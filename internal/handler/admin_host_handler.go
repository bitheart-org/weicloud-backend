package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/dto"
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
		fail(c, http.StatusInternalServerError, 50030, "query hosts failed")
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
		fail(c, http.StatusBadRequest, 40030, "invalid request")
		return
	}

	host, err := h.hostService.Create(c.Request.Context(), req)
	if err != nil {
		fail(c, http.StatusBadRequest, 40031, "create host failed")
		return
	}
	ok(c, service.ToHostPayload(host))
}

func (h *AdminHostHandler) Update(c *gin.Context) {
	var req dto.UpdateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40032, "invalid request")
		return
	}

	hostID := c.Param("id")
	host, err := h.hostService.Update(c.Request.Context(), hostID, req)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40430, "host not found")
			return
		}
		fail(c, http.StatusBadRequest, 40033, "update host failed")
		return
	}
	ok(c, service.ToHostPayload(host))
}

func (h *AdminHostHandler) Delete(c *gin.Context) {
	hostID := c.Param("id")
	if err := h.hostService.Delete(hostID); err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40431, "host not found")
			return
		}
		fail(c, http.StatusInternalServerError, 50031, "delete host failed")
		return
	}
	ok(c, gin.H{"id": hostID})
}

func (h *AdminHostHandler) Sync(c *gin.Context) {
	hostID := c.Param("id")
	host, err := h.hostService.Sync(c.Request.Context(), hostID)
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, 40432, "host not found")
			return
		}
		fail(c, http.StatusBadGateway, 50230, "sync host failed")
		return
	}
	ok(c, service.ToHostPayload(host))
}
