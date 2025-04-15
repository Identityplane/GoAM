package web

import (
	"encoding/json"
	"goiam/internal/realms"
	"runtime"
	"strconv"

	"github.com/valyala/fasthttp"
)

func handleLiveness(ctx *fasthttp.RequestCtx) {
	resp := map[string]string{"status": "alive"}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")

	jsonData, _ := json.MarshalIndent(resp, "", "  ")
	ctx.SetBody(jsonData)
}

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
