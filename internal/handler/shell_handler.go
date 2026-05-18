package handler

import (
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/errcode"
	"weicloud-backend/internal/middleware"
	"weicloud-backend/internal/service"
)

type UserShellHandler struct {
	vmService    *service.VMService
	shellService *service.ShellService
}

func NewUserShellHandler(vmService *service.VMService, shellService *service.ShellService) *UserShellHandler {
	return &UserShellHandler{
		vmService:    vmService,
		shellService: shellService,
	}
}

func (h *UserShellHandler) IssueToken(c *gin.Context) {
	claims, exists := middleware.CurrentUserClaims(c)
	if !exists {
		fail(c, http.StatusUnauthorized, errcode.ShellIssueTokenUnauthorized, "unauthorized")
		return
	}
	token, err := h.vmService.IssueShellOneTimeToken(h.shellService, c.Param("id"), claims.UserID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.ShellIssueTokenNotFound, "vm not found")
			return
		}
		fail(c, http.StatusBadRequest, errcode.ShellIssueTokenFailed, "issue shell token failed")
		return
	}
	ok(c, dto.VncTokenResponse{Token: token})
}

type ShellWSHandler struct {
	vmService    *service.VMService
	shellService *service.ShellService
	upgrader     websocket.Upgrader
}

func NewShellWSHandler(vmService *service.VMService, shellService *service.ShellService) *ShellWSHandler {
	return &ShellWSHandler{
		vmService:    vmService,
		shellService: shellService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
	}
}

func (h *ShellWSHandler) Proxy(c *gin.Context) {
	vmID := c.Param("id")
	token := c.Query("token")
	if token == "" {
		fail(c, http.StatusUnauthorized, errcode.ShellMissingToken, "missing shell token")
		return
	}

	userID, err := h.shellService.ConsumeToken(token, vmID)
	if err != nil {
		fail(c, http.StatusUnauthorized, errcode.ShellInvalidToken, "invalid shell token")
		return
	}
	defer h.shellService.ReleaseSession(vmID)

	remoteConn, err := h.vmService.OpenShellForOwner(c.Request.Context(), vmID, userID)
	if err != nil {
		fail(c, http.StatusBadGateway, errcode.ShellConnectBackendFailed, "connect shell backend failed")
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
		closeWith(websocket.CloseInternalServerErr, "shell proxy closed")
	}
}
