package admin_api

import (
	"encoding/json"
	"goiam/internal/model"
	"goiam/internal/service"
	"net/http"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/valyala/fasthttp"
)

// DashboardResponse represents the response structure for the dashboard endpoint
type DashboardResponse struct {
	UserStats          *model.UserStats `json:"user_stats"`
	Flows              FlowInfo         `json:"flows"`
	NumberApplications int              `json:"number_applications"`
}

// FlowInfo represents flow statistics in the dashboard
type FlowInfo struct {
	TotalFlows  int `json:"total_flows"`
	ActiveFlows int `json:"active_flows"`
}

// HandleDashboard returns combined dashboard data for a tenant/realm
// @Summary Get dashboard data
// @Description Returns combined user stats and flow information for a tenant/realm
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 200 {object} DashboardResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/dashboard [get]
func HandleDashboard(ctx *fasthttp.RequestCtx) {
	// Get tenant and realm from path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	services := service.GetServices()

	// Lookup the loaded realm
	_, ok := services.RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Get user stats
	stats, err := service.GetServices().UserService.GetUserStats(ctx, tenant, realm)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get user stats: " + err.Error())
		return
	}

	// Get number of applications
	applications, err := service.GetServices().ApplicationService.ListApplications(tenant, realm)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get number of applications: " + err.Error())
		return
	}

	// Get flows
	flows, err := services.FlowService.ListFlows(tenant, realm)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to list flows: " + err.Error())
		return
	}

	// number of active flows
	activeFlows := 0
	for _, flow := range flows {
		if flow.Active {
			activeFlows++
		}
	}

	// Create response
	response := DashboardResponse{
		UserStats: stats,
		Flows: FlowInfo{
			TotalFlows:  len(flows),
			ActiveFlows: activeFlows,
		},
		NumberApplications: len(applications),
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	// Set response headers and body
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

type SystemStats struct {
	CacheMetrics service.CacheMetrics `json:"cache_metrics"`
	MemoryUsage  MemoryUsage          `json:"memory_usage"`
}

type MemoryUsage struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// HandleSystemStats returns system-wide statistics including cache metrics and memory usage
// @Summary Get system statistics
// @Description Returns system-wide statistics including cache metrics and memory usage
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} SystemStats
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/system/stats [get]
func HandleSystemStats(ctx *fasthttp.RequestCtx) {

	cacheMetrics := service.GetServices().CacheService.GetMetrics()

	// Calculate current memory usage of the server
	v, err := mem.VirtualMemory()

	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get memory usage: " + err.Error())
		return
	}

	memoryUsage := MemoryUsage{
		Total:       v.Total,
		Used:        v.Used,
		Free:        v.Free,
		UsedPercent: v.UsedPercent,
	}

	response := SystemStats{
		CacheMetrics: cacheMetrics,
		MemoryUsage:  memoryUsage,
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to marshal response: " + err.Error())
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}
