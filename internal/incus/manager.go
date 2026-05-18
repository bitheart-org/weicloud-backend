package incus

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"weicloud-backend/internal/model"
)

type ResourceSnapshot struct {
	CPUCores    int
	MemoryBytes int64
	Status      string
	LastSeenAt  *time.Time
}

type hostClient struct {
	baseURL    string
	httpClient *http.Client
}

type Manager struct {
	mu      sync.RWMutex
	clients map[string]*hostClient
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
		return ResourceSnapshot{
			Status: model.HostStatusOffline,
		}, err
	}

	resources, err := m.fetchResources(ctx, client)
	if err != nil {
		return ResourceSnapshot{
			Status: model.HostStatusOffline,
		}, err
	}

	now := time.Now()
	return ResourceSnapshot{
		CPUCores:    resources.CPUCores,
		MemoryBytes: resources.MemoryBytes,
		Status:      model.HostStatusOnline,
		LastSeenAt:  &now,
	}, nil
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

type incusEnvelope struct {
	Metadata json.RawMessage `json:"metadata"`
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

	var envelope incusEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return ResourceSnapshot{}, fmt.Errorf("decode resources envelope: %w", err)
	}

	var metadata map[string]any
	if err := json.Unmarshal(envelope.Metadata, &metadata); err != nil {
		return ResourceSnapshot{}, fmt.Errorf("decode resources metadata: %w", err)
	}

	return ResourceSnapshot{
		CPUCores:    extractCPUCores(metadata),
		MemoryBytes: extractMemoryTotal(metadata),
	}, nil
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
