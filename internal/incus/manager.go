package incus

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"weicloud-backend/internal/model"
)

type ResourceSnapshot struct {
	CPUCores    int
	MemoryBytes int64
	Status      string
	LastSeenAt  *time.Time
}

type InstanceCreateRequest struct {
	InstanceName   string
	Image          string
	CPUCores       int
	MemoryBytes    int64
	DiskRootBytes  int64
	NetworkIngress string
	NetworkEgress  string
	LoginUsername  string
	LoginPassword  string
	FRPSServerAddr string
	FRPSServerPort int
	FRPSToken      string
	SSHRemotePort  int
}

type ImageInfo struct {
	Alias        string `json:"alias"`
	Architecture string `json:"architecture"`
	Description  string `json:"description"`
	Type         string `json:"type"`
}

type InstanceMetrics struct {
	CPUNanoseconds int64
	MemoryBytes    int64
}

type hostClient struct {
	baseURL    string
	httpClient *http.Client
}

type Manager struct {
	mu      sync.RWMutex
	clients map[string]*hostClient
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "status=404")
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*hostClient),
	}
}

func (m *Manager) RegisterHost(host model.Host) error {
	tlsCert, err := tls.X509KeyPair([]byte(host.Certificate), []byte(host.Key))
	if err != nil {
		return fmt.Errorf("invalid host tls cert/key: %w", err)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{tlsCert},
				InsecureSkipVerify: true,
			},
		},
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[host.ID.String()] = &hostClient{
		baseURL:    strings.TrimSuffix(host.Address, "/"),
		httpClient: client,
	}
	return nil
}

func (m *Manager) RemoveHost(hostID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, hostID)
}

func (m *Manager) SyncHost(ctx context.Context, host model.Host) (ResourceSnapshot, error) {
	client, err := m.getOrCreateClient(host)
	if err != nil {
		return ResourceSnapshot{Status: model.HostStatusOffline}, err
	}

	resources, err := m.fetchResources(ctx, client)
	if err != nil {
		return ResourceSnapshot{Status: model.HostStatusOffline}, err
	}

	now := time.Now()
	return ResourceSnapshot{
		CPUCores:    resources.CPUCores,
		MemoryBytes: resources.MemoryBytes,
		Status:      model.HostStatusOnline,
		LastSeenAt:  &now,
	}, nil
}

func (m *Manager) CreateInstance(ctx context.Context, host model.Host, req InstanceCreateRequest) error {
	client, err := m.getOrCreateClient(host)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"name": req.InstanceName,
		"type": "virtual-machine",
		"source": map[string]any{
			"type":     "image",
			"alias":    req.Image,
			"mode":     "pull",
			"protocol": "simplestreams",
			"server":   "https://images.linuxcontainers.org",
		},
		"config": map[string]string{
			"limits.cpu":     fmt.Sprintf("%d", req.CPUCores),
			"limits.memory":  fmt.Sprintf("%d", req.MemoryBytes),
			"agent":          "enabled",
			"user.user-data": buildCloudInit(req),
		},
		"devices": map[string]map[string]string{
			"root": {
				"type": "disk",
				"path": "/",
				"size": fmt.Sprintf("%d", req.DiskRootBytes),
				"pool": "default",
			},
			"eth0": {
				"type":    "nic",
				"network": "incusbr0",
				"name":    "eth0",
			},
		},
	}
	if req.NetworkIngress != "" {
		payload["devices"].(map[string]map[string]string)["eth0"]["limits.ingress"] = req.NetworkIngress
	}
	if req.NetworkEgress != "" {
		payload["devices"].(map[string]map[string]string)["eth0"]["limits.egress"] = req.NetworkEgress
	}
	return m.doJSON(ctx, client, http.MethodPost, "/1.0/instances", payload)
}

func buildCloudInit(req InstanceCreateRequest) string {
	return fmt.Sprintf(`#cloud-config
users:
  - default
  - name: %s
    lock_passwd: false
    shell: /bin/bash
    sudo: ALL=(ALL) NOPASSWD:ALL
ssh_pwauth: true
chpasswd:
  expire: false
  list:
    - %s:%s
write_files:
  - path: /etc/frp/frpc.toml
    permissions: "0600"
    content: |
      serverAddr = "%s"
      serverPort = %d
      auth.method = "token"
      auth.token = "%s"

      [[proxies]]
      name = "ssh-%s"
      type = "tcp"
      localIP = "127.0.0.1"
      localPort = 22
      remotePort = %d
  - path: /etc/systemd/system/frpc.service
    permissions: "0644"
    content: |
      [Unit]
      Description=frp client
      After=network.target
      Wants=network-online.target

      [Service]
      Type=simple
      ExecStart=/usr/local/bin/frpc -c /etc/frp/frpc.toml
      Restart=always
      RestartSec=5s

      [Install]
      WantedBy=multi-user.target
runcmd:
  - [bash, -lc, "set -euo pipefail; arch=$(uname -m); case $arch in x86_64) frp_arch=amd64 ;; aarch64|arm64) frp_arch=arm64 ;; *) exit 0 ;; esac; ver=0.58.1; curl -fsSL -o /tmp/frp.tgz https://github.com/fatedier/frp/releases/download/v${ver}/frp_${ver}_linux_${frp_arch}.tar.gz; tar -xzf /tmp/frp.tgz -C /tmp; install -m 0755 /tmp/frp_${ver}_linux_${frp_arch}/frpc /usr/local/bin/frpc; systemctl daemon-reload; systemctl enable --now frpc"]
`,
		req.LoginUsername,
		req.LoginUsername,
		req.LoginPassword,
		req.FRPSServerAddr,
		req.FRPSServerPort,
		req.FRPSToken,
		req.InstanceName,
		req.SSHRemotePort,
	)
}

func (m *Manager) DeleteInstance(ctx context.Context, host model.Host, instanceName string) error {
	client, err := m.getOrCreateClient(host)
	if err != nil {
		return err
	}
	path := "/1.0/instances/" + url.PathEscape(instanceName)
	return m.doJSON(ctx, client, http.MethodDelete, path, nil)
}

func (m *Manager) SetInstanceState(ctx context.Context, host model.Host, instanceName, action string, force bool) error {
	client, err := m.getOrCreateClient(host)
	if err != nil {
		return err
	}
	path := "/1.0/instances/" + url.PathEscape(instanceName) + "/state"
	payload := map[string]any{
		"action":  action,
		"timeout": 30,
		"force":   force,
	}
	return m.doJSON(ctx, client, http.MethodPut, path, payload)
}

func (m *Manager) UpdateInstanceConfig(ctx context.Context, host model.Host, instanceName string, cpuCores int, memoryBytes int64) error {
	client, err := m.getOrCreateClient(host)
	if err != nil {
		return err
	}
	instance, err := m.getInstance(ctx, client, instanceName)
	if err != nil {
		return err
	}

	instance.Metadata.Config["limits.cpu"] = fmt.Sprintf("%d", cpuCores)
	instance.Metadata.Config["limits.memory"] = fmt.Sprintf("%d", memoryBytes)
	path := "/1.0/instances/" + url.PathEscape(instanceName)
	return m.doJSON(ctx, client, http.MethodPut, path, map[string]any{
		"config":  instance.Metadata.Config,
		"devices": instance.Metadata.Devices,
	})
}

func (m *Manager) ResizeRootDisk(ctx context.Context, host model.Host, instanceName string, diskRootBytes int64) error {
	client, err := m.getOrCreateClient(host)
	if err != nil {
		return err
	}
	instance, err := m.getInstance(ctx, client, instanceName)
	if err != nil {
		return err
	}
	rootDevice, ok := instance.Metadata.Devices["root"]
	if !ok {
		return fmt.Errorf("instance %s missing root device", instanceName)
	}
	rootDevice["size"] = fmt.Sprintf("%d", diskRootBytes)
	instance.Metadata.Devices["root"] = rootDevice

	path := "/1.0/instances/" + url.PathEscape(instanceName)
	return m.doJSON(ctx, client, http.MethodPut, path, map[string]any{
		"config":  instance.Metadata.Config,
		"devices": instance.Metadata.Devices,
	})
}

func (m *Manager) UpdateNetworkLimits(ctx context.Context, host model.Host, instanceName string, ingress, egress string) error {
	client, err := m.getOrCreateClient(host)
	if err != nil {
		return err
	}
	instance, err := m.getInstance(ctx, client, instanceName)
	if err != nil {
		return err
	}

	nicName := ""
	for name, device := range instance.Metadata.Devices {
		if device["type"] == "nic" {
			nicName = name
			break
		}
	}
	if nicName == "" {
		return fmt.Errorf("instance %s missing nic device", instanceName)
	}
	device := instance.Metadata.Devices[nicName]
	if ingress != "" {
		device["limits.ingress"] = ingress
	}
	if egress != "" {
		device["limits.egress"] = egress
	}
	instance.Metadata.Devices[nicName] = device

	path := "/1.0/instances/" + url.PathEscape(instanceName)
	return m.doJSON(ctx, client, http.MethodPut, path, map[string]any{
		"config":  instance.Metadata.Config,
		"devices": instance.Metadata.Devices,
	})
}

func (m *Manager) ListImages(ctx context.Context) ([]ImageInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://images.linuxcontainers.org/streams/v1/images.json", nil)
	if err != nil {
		return nil, fmt.Errorf("build image list request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request image list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("image list status: %d", resp.StatusCode)
	}

	var payload struct {
		Products map[string]struct {
			Aliases      string `json:"aliases"`
			Architecture string `json:"arch"`
			OS           string `json:"os"`
			Release      string `json:"release"`
			Variant      string `json:"variant"`
		} `json:"products"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode image list: %w", err)
	}

	items := make([]ImageInfo, 0, 100)
	for _, p := range payload.Products {
		if len(items) >= 100 {
			break
		}
		alias := strings.TrimSpace(strings.Split(p.Aliases, ",")[0])
		if alias == "" {
			continue
		}
		items = append(items, ImageInfo{
			Alias:        alias,
			Architecture: p.Architecture,
			Description:  strings.TrimSpace(p.OS + " " + p.Release + " " + p.Variant),
			Type:         "virtual-machine",
		})
	}
	return items, nil
}

func (m *Manager) GetInstanceMetrics(ctx context.Context, host model.Host, instanceName string) (InstanceMetrics, error) {
	client, err := m.getOrCreateClient(host)
	if err != nil {
		return InstanceMetrics{}, err
	}
	path := "/1.0/instances/" + url.PathEscape(instanceName) + "/state"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.baseURL+path, nil)
	if err != nil {
		return InstanceMetrics{}, fmt.Errorf("build metrics request: %w", err)
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return InstanceMetrics{}, fmt.Errorf("request instance metrics: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return InstanceMetrics{}, fmt.Errorf("instance metrics status: %d", resp.StatusCode)
	}

	var envelope struct {
		Metadata struct {
			CPU struct {
				Usage int64 `json:"usage"`
			} `json:"cpu"`
			Memory struct {
				Usage int64 `json:"usage"`
			} `json:"memory"`
		} `json:"metadata"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return InstanceMetrics{}, fmt.Errorf("decode instance metrics: %w", err)
	}

	return InstanceMetrics{
		CPUNanoseconds: envelope.Metadata.CPU.Usage,
		MemoryBytes:    envelope.Metadata.Memory.Usage,
	}, nil
}

func (m *Manager) ResetRootPassword(ctx context.Context, host model.Host, instanceName string, password string) error {
	return m.SetUserPassword(ctx, host, instanceName, "root", password)
}

func (m *Manager) SetUserPassword(ctx context.Context, host model.Host, instanceName, username, password string) error {
	client, err := m.getOrCreateClient(host)
	if err != nil {
		return err
	}

	pair := strings.ReplaceAll(username+":"+password, "'", `'\''`)
	command := fmt.Sprintf("printf '%%s\n' '%s' | chpasswd", pair)
	payload := map[string]any{
		"command":            []string{"/bin/sh", "-c", command},
		"interactive":        false,
		"wait-for-websocket": false,
	}
	raw, err := m.doJSONRaw(ctx, client, http.MethodPost, "/1.0/instances/"+url.PathEscape(instanceName)+"/exec", payload)
	if err != nil {
		return err
	}

	var operationResp struct {
		Operation string `json:"operation"`
	}
	if err := json.Unmarshal(raw, &operationResp); err != nil {
		return fmt.Errorf("decode exec operation response: %w", err)
	}
	if operationResp.Operation == "" {
		return fmt.Errorf("missing exec operation id")
	}

	return m.waitOperation(ctx, client, operationResp.Operation, 30)
}

func (m *Manager) OpenVNCConnection(ctx context.Context, host model.Host, instanceName string) (*websocket.Conn, error) {
	return m.openConsoleConnection(ctx, host, instanceName, "vga")
}

func (m *Manager) OpenShellConnection(ctx context.Context, host model.Host, instanceName string) (*websocket.Conn, error) {
	return m.openConsoleConnection(ctx, host, instanceName, "console")
}

func (m *Manager) openConsoleConnection(ctx context.Context, host model.Host, instanceName, consoleType string) (*websocket.Conn, error) {
	client, err := m.getOrCreateClient(host)
	if err != nil {
		return nil, err
	}

	raw, err := m.doJSONRaw(
		ctx,
		client,
		http.MethodPost,
		"/1.0/instances/"+url.PathEscape(instanceName)+"/console",
		map[string]any{
			"type": consoleType,
		},
	)
	if err != nil {
		return nil, err
	}

	var consoleResp struct {
		Operation string `json:"operation"`
	}
	if err := json.Unmarshal(raw, &consoleResp); err != nil {
		return nil, fmt.Errorf("decode open console response: %w", err)
	}
	if consoleResp.Operation == "" {
		return nil, fmt.Errorf("missing console operation id")
	}

	secret, err := m.getOperationWebsocketSecret(ctx, client, consoleResp.Operation)
	if err != nil {
		return nil, err
	}

	wsURL, err := buildOperationWebsocketURL(client.baseURL, consoleResp.Operation, secret)
	if err != nil {
		return nil, err
	}

	dialer := websocket.Dialer{
		TLSClientConfig: extractTLSConfig(client),
	}
	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("dial %s websocket: %w", consoleType, err)
	}
	return conn, nil
}

func (m *Manager) getOrCreateClient(host model.Host) (*hostClient, error) {
	hostID := host.ID.String()
	m.mu.RLock()
	client, exists := m.clients[hostID]
	m.mu.RUnlock()
	if exists {
		return client, nil
	}

	if err := m.RegisterHost(host); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[hostID], nil
}

type resourcesEnvelope struct {
	Metadata map[string]any `json:"metadata"`
}

type instanceEnvelope struct {
	Metadata struct {
		Config  map[string]string            `json:"config"`
		Devices map[string]map[string]string `json:"devices"`
		Status  string                       `json:"status"`
	} `json:"metadata"`
}

func (m *Manager) fetchResources(ctx context.Context, client *hostClient) (ResourceSnapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.baseURL+"/1.0/resources", nil)
	if err != nil {
		return ResourceSnapshot{}, fmt.Errorf("build resources request: %w", err)
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return ResourceSnapshot{}, fmt.Errorf("request incus resources: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ResourceSnapshot{}, fmt.Errorf("incus resources status: %d", resp.StatusCode)
	}

	var envelope resourcesEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return ResourceSnapshot{}, fmt.Errorf("decode resources envelope: %w", err)
	}
	return ResourceSnapshot{
		CPUCores:    extractCPUCores(envelope.Metadata),
		MemoryBytes: extractMemoryTotal(envelope.Metadata),
	}, nil
}

func (m *Manager) getInstance(ctx context.Context, client *hostClient, instanceName string) (instanceEnvelope, error) {
	path := "/1.0/instances/" + url.PathEscape(instanceName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.baseURL+path, nil)
	if err != nil {
		return instanceEnvelope{}, fmt.Errorf("build get instance request: %w", err)
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return instanceEnvelope{}, fmt.Errorf("request get instance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return instanceEnvelope{}, fmt.Errorf("get instance status: %d", resp.StatusCode)
	}

	var envelope instanceEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return instanceEnvelope{}, fmt.Errorf("decode get instance response: %w", err)
	}
	if envelope.Metadata.Config == nil {
		envelope.Metadata.Config = map[string]string{}
	}
	if envelope.Metadata.Devices == nil {
		envelope.Metadata.Devices = map[string]map[string]string{}
	}
	return envelope, nil
}

func (m *Manager) doJSON(ctx context.Context, client *hostClient, method, path string, body any) error {
	_, err := m.doJSONRaw(ctx, client, method, path, body)
	return err
}

func (m *Manager) doJSONRaw(ctx context.Context, client *hostClient, method, path string, body any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, client.baseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request incus: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("incus request failed status=%d body=%s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func (m *Manager) waitOperation(ctx context.Context, client *hostClient, operationPath string, timeoutSec int) error {
	waitPath := operationPath + "/wait?timeout=" + fmt.Sprintf("%d", timeoutSec)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.baseURL+waitPath, nil)
	if err != nil {
		return fmt.Errorf("build wait operation request: %w", err)
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("wait operation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("wait operation failed status=%d body=%s", resp.StatusCode, string(body))
	}

	var operationResp struct {
		Metadata struct {
			Status string `json:"status"`
			Err    string `json:"err"`
		} `json:"metadata"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&operationResp); err != nil {
		return fmt.Errorf("decode wait operation response: %w", err)
	}
	if strings.ToLower(operationResp.Metadata.Status) != "success" {
		if operationResp.Metadata.Err != "" {
			return fmt.Errorf("operation failed: %s", operationResp.Metadata.Err)
		}
		return fmt.Errorf("operation failed status: %s", operationResp.Metadata.Status)
	}
	return nil
}

func (m *Manager) getOperationWebsocketSecret(ctx context.Context, client *hostClient, operationPath string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.baseURL+operationPath, nil)
	if err != nil {
		return "", fmt.Errorf("build operation request: %w", err)
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request operation metadata: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("operation metadata status=%d body=%s", resp.StatusCode, string(body))
	}

	var operationResp struct {
		Metadata struct {
			Fds map[string]string `json:"fds"`
		} `json:"metadata"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&operationResp); err != nil {
		return "", fmt.Errorf("decode operation metadata: %w", err)
	}
	secret := operationResp.Metadata.Fds["0"]
	if secret == "" {
		return "", fmt.Errorf("missing operation websocket secret")
	}
	return secret, nil
}

func buildOperationWebsocketURL(baseURL, operationPath, secret string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse host url: %w", err)
	}
	switch parsed.Scheme {
	case "https":
		parsed.Scheme = "wss"
	case "http":
		parsed.Scheme = "ws"
	default:
		return "", fmt.Errorf("unsupported host scheme: %s", parsed.Scheme)
	}
	parsed.Path = operationPath + "/websocket"
	query := url.Values{}
	query.Set("secret", secret)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func extractTLSConfig(client *hostClient) *tls.Config {
	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		return nil
	}
	if transport.TLSClientConfig == nil {
		return nil
	}
	return transport.TLSClientConfig.Clone()
}

func extractCPUCores(metadata map[string]any) int {
	cpuRaw, ok := metadata["cpu"].(map[string]any)
	if !ok {
		return 0
	}
	if total, ok := cpuRaw["total"]; ok {
		switch v := total.(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
	}

	sockets, ok := cpuRaw["sockets"].([]any)
	if !ok {
		return 0
	}
	totalCores := 0
	for _, item := range sockets {
		socket, ok := item.(map[string]any)
		if !ok {
			continue
		}
		cores, ok := socket["cores"].([]any)
		if !ok {
			totalCores++
			continue
		}
		totalCores += len(cores)
	}
	return totalCores
}

func extractMemoryTotal(metadata map[string]any) int64 {
	memoryRaw, ok := metadata["memory"].(map[string]any)
	if !ok {
		return 0
	}
	total, ok := memoryRaw["total"]
	if !ok {
		return 0
	}
	switch v := total.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}
