package web

import (
	"encoding/json"
	"goiam/internal/realms"
	"runtime"
	"strconv"

	"github.com/valyala/fasthttp"
)

// handleLiveness checks if the service is alive
// @Summary Check service liveness
// @Description Returns a simple status indicating if the service is alive
// @Tags Health
// @Produce json
// @Success 200 {object} object "Service status"
// @Router /healthz [get]
func handleLiveness(ctx *fasthttp.RequestCtx) {
	resp := map[string]string{"status": "alive"}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")

	jsonData, _ := json.MarshalIndent(resp, "", "  ")
	ctx.SetBody(jsonData)
}

// handleReadiness checks if the service is ready to handle requests
// @Summary Check service readiness
// @Description Returns the readiness status of the service and its components
// @Tags Health
// @Produce json
// @Success 200 {object} object "Service and components readiness status"
// @Failure 503 {object} object "Service is not ready"
// @Router /readyz [get]
func handleReadiness(ctx *fasthttp.RequestCtx) {
	ready := map[string]string{}
	isReady := true

	if len(realms.GetAllRealms()) == 0 {
		ready["Realms"] = "not ready"
		isReady = false
	} else {
		ready["Realms"] = "ready"
	}

	ready["Ready"] = strconv.FormatBool(isReady)

	if isReady {
		ctx.SetStatusCode(fasthttp.StatusOK)
	} else {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
	}

	ctx.SetContentType("application/json")
	jsonData, _ := json.MarshalIndent(ready, "", "  ")
	ctx.SetBody(jsonData)
}

// handleInfo returns basic service information
// @Summary Get service information
// @Description Returns basic information about the service including version
// @Tags Health
// @Produce json
// @Success 200 {object} object "Service information"
// @Router /info [get]
func handleInfo(ctx *fasthttp.RequestCtx) {
	info := map[string]string{
		"name":       "GoIAM",
		"go_version": runtime.Version(),
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	jsonData, _ := json.MarshalIndent(info, "", "  ")
	ctx.SetBody(jsonData)
}
