package admin_api

import (
	"encoding/json"
	"goiam/internal/model"
	"goiam/internal/realms"
	"net/http"

	"github.com/valyala/fasthttp"
)

// DashboardResponse represents the combined dashboard data
type DashboardResponse struct {
	UserStats *model.UserStats `json:"user_stats"`
	Flows     FlowInfo         `json:"flows"`
}

// FlowInfo represents information about a flow
type FlowInfo struct {
	TotalFlows  int `json:"total_flows"`
	ActiveFlows int `json:"active_flows"`
}

// HandleDashboard returns combined dashboard data for a tenant/realm
// @Summary Get dashboard data
// @Description Returns combined user stats and flow information for a tenant/realm
// @Tags dashboard
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 200 {object} DashboardResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /{tenant}/{realm}/admin/dashboard [get]
func (h *Handler) HandleDashboard(ctx *fasthttp.RequestCtx) {
	// Get tenant and realm from path parameters
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// Lookup the loaded realm
	_, ok := realms.GetRealm(tenant + "/" + realm)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetBodyString("Realm not found")
		return
	}

	// Get user stats
	stats, err := h.userService.GetUserStats(ctx, tenant, realm)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to get user stats: " + err.Error())
		return
	}

	// Get flows
	flows, err := realms.ListFlowsPerRealm(tenant, realm)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetBodyString("Failed to list flows: " + err.Error())
		return
	}

	// Create response
	response := DashboardResponse{
		UserStats: stats,
		Flows: FlowInfo{
			TotalFlows:  len(flows),
			ActiveFlows: len(flows),
		},
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
