package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"weicloud-backend/internal/errcode"
	"weicloud-backend/internal/service"
)

type AdminSystemHandler struct {
	systemService *service.SystemService
	vmService     *service.VMService
}

func NewAdminSystemHandler(systemService *service.SystemService, vmService *service.VMService) *AdminSystemHandler {
	return &AdminSystemHandler{
		systemService: systemService,
		vmService:     vmService,
	}
}

func (h *AdminSystemHandler) Dashboard(c *gin.Context) {
	payload, err := h.systemService.Dashboard()
	if err != nil {
		fail(c, http.StatusInternalServerError, errcode.AdminSystemDashboardFailed, "query dashboard failed")
		return
	}
	ok(c, payload)
}

func (h *AdminSystemHandler) Logs(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("page_size"), 20)
	if pageSize > 100 {
		pageSize = 100
	}

	result, err := h.systemService.ListOperationLogs(page, pageSize)
	if err != nil {
		fail(c, http.StatusInternalServerError, errcode.AdminSystemLogsFailed, "query logs failed")
		return
	}
	ok(c, result)
}

func (h *AdminSystemHandler) VMResource(c *gin.Context) {
	payload, err := h.vmService.GetResourceByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if service.IsNotFound(err) {
			fail(c, http.StatusNotFound, errcode.AdminSystemVMNotFound, "vm not found")
			return
		}
		fail(c, http.StatusBadGateway, errcode.AdminSystemVMResourceFailed, "query vm resource failed")
		return
	}
	ok(c, payload)
}

func parseIntDefault(raw string, defaultValue int) int {
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return defaultValue
	}
	return v
}
