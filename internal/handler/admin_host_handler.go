package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/dto"
	"weicloud-backend/internal/errcode"
	"weicloud-backend/internal/service"
)

type AdminHostHandler struct {
	hostService *service.HostService
}

func NewAdminHostHandler(hostService *service.HostService) *AdminHostHandler {
	return &AdminHostHandler{
		hostService: hostService,
	}
}

func (h *AdminHostHandler) List(c *gin.Context) {
	hosts, err := h.hostService.List()
	if err != nil {
		fail(c, http.StatusInternalServerError, errcode.AdminHostQueryFailed, "query hosts failed")
		return
	}

	items := make([]dto.HostPayload, 0, len(hosts))
	for _, host := range hosts {
		items = append(items, service.ToHostPayload(host))
	}
	ok(c, gin.H{"items": items})
}

func (h *AdminHostHandler) Create(c *gin.Context) {
	var req dto.CreateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminHostCreateInvalidInput, "invalid request")
		return
	}

	host, err := h.hostService.Create(c.Request.Context(), req)
	if err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminHostCreateFailed, "create host failed")
		return
	}
	ok(c, service.ToHostPayload(host))
}

func (h *AdminHostHandler) Update(c *gin.Context) {
	var req dto.UpdateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, errcode.AdminHostUpdateInvalidInput, "invalid request")
		return
	}

	hostID := c.Param("id")
	host, err := h.hostService.Update(c.Request.Context(), hostID, req)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminHostNotFoundForUpdate, "host not found")
			return
		}
		fail(c, http.StatusBadRequest, errcode.AdminHostUpdateFailed, "update host failed")
		return
	}
	ok(c, service.ToHostPayload(host))
}

func (h *AdminHostHandler) Delete(c *gin.Context) {
	hostID := c.Param("id")
	if err := h.hostService.Delete(hostID); err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminHostNotFoundForDelete, "host not found")
			return
		}
		fail(c, http.StatusInternalServerError, errcode.AdminHostDeleteFailed, "delete host failed")
		return
	}
	ok(c, gin.H{"id": hostID})
}

func (h *AdminHostHandler) Sync(c *gin.Context) {
	hostID := c.Param("id")
	host, err := h.hostService.Sync(c.Request.Context(), hostID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminHostNotFoundForSync, "host not found")
			return
		}
		fail(c, http.StatusBadGateway, errcode.AdminHostSyncFailed, "sync host failed")
		return
	}
	ok(c, service.ToHostPayload(host))
}
