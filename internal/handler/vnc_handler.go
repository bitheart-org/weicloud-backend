package handler

import (
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/middleware"
	"weicloud-backend/internal/service"
)

type UserVNCHandler struct {
	vmService  *service.VMService
	vncService *service.VNCService
}

func NewUserVNCHandler(vmService *service.VMService, vncService *service.VNCService) *UserVNCHandler {
	return &UserVNCHandler{
		vmService:  vmService,
		vncService: vncService,
	}
}

func (h *UserVNCHandler) IssueToken(c *gin.Context) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, 40135, "unauthorized")
		return
	}
	token, err := h.vmService.IssueVNCOneTimeToken(h.vncService, c.Param("id"), claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40454, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, 40061, "issue vnc token failed")
		return
	}
	ok(c, dto.VncTokenResponse{Token: token})
}

type VNCWSHandler struct {
	vmService  *service.VMService
	vncService *service.VNCService
	upgrader   websocket.Upgrader
}

func NewVNCWSHandler(vmService *service.VMService, vncService *service.VNCService) *VNCWSHandler {
	return &VNCWSHandler{
		vmService:  vmService,
		vncService: vncService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
	}
}

func (h *VNCWSHandler) Proxy(c *gin.Context) {
	vmID := c.Param("id")
	token := c.Query("token")
	if token == "" {
		fail(c, http.StatusUnauthorized, 40136, "missing vnc token")
		return
	}

	userID, err := h.vncService.ConsumeToken(token, vmID)
	if err != nil {
		fail(c, http.StatusUnauthorized, 40137, "invalid vnc token")
		return
	}
	defer h.vncService.ReleaseSession(vmID)

	remoteConn, err := h.vmService.OpenVNCForOwner(c.Request.Context(), vmID, userID)
	if err != nil {
		fail(c, http.StatusBadGateway, 50270, "connect vnc backend failed")
		return
	}
	defer remoteConn.Close()

	clientConn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer clientConn.Close()

	errCh := make(chan error, 2)
	var once sync.Once
	closeWith := func(closeCode int, text string) {
		once.Do(func() {
			_ = clientConn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(closeCode, text),
				time.Now().Add(time.Second),
			)
		})
	}

	go proxyWS(clientConn, remoteConn, errCh)
	go proxyWS(remoteConn, clientConn, errCh)

	if err := <-errCh; err != nil && err != io.EOF {
		closeWith(websocket.CloseInternalServerErr, "vnc proxy closed")
	}
}

func proxyWS(src, dst *websocket.Conn, errCh chan<- error) {
	for {
		messageType, data, err := src.ReadMessage()
		if err != nil {
			errCh <- err
			return
		}
		if err := dst.WriteMessage(messageType, data); err != nil {
			errCh <- err
			return
		}
	}
}
