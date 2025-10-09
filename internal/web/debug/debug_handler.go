package debug

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/Identityplane/GoAM/internal/auth/graph/visual"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/service"

	"github.com/valyala/fasthttp"
)

// HandleListAllFlows returns a list of all flows across all tenants/realms.
// @Summary List all flows across all tenants/realms
// @Description Returns a plain text list of all authentication flows across all tenants and realms
// @Tags Debug
// @Produce text/plain
// @Success 200 {string} string "List of flows in plain text format"
// @Router /debug/flows/all [get]
func HandleListAllFlows(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/plain; charset=utf-8")

	allFlows, err := service.GetServices().FlowService.ListAllFlows()

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Failed to list flows: " + err.Error())
		return
	}

	// Set the response type to JSON and send the flow list
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)

	// use pretty printing for better readability
	allFlowsJSON, err := json.MarshalIndent(allFlows, "", "  ")

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Failed to encode flow list: " + err.Error())
		return
	}

	ctx.SetBody(allFlowsJSON)
}

// HandleListFlows responds with a list of available flow names and routes
// @Summary List flows for a specific tenant and realm
// @Description Returns a JSON list of available authentication flows for the specified tenant and realm
// @Tags Debug
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 200 {array} object "List of flows with name and route"
// @Failure 500 {string} string "Internal server error"
// @Router /{tenant}/{realm}/debug/flows [get]
func HandleListFlows(ctx *fasthttp.RequestCtx) {

	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	flows, err := service.GetServices().FlowService.ListFlows(tenant, realm)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Failed to list flows: " + err.Error())
		return
	}

	// Set the response type to JSON and send the flow list
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)

	// use pretty printing for better readability
	activeFlowsJSON, err := json.MarshalIndent(flows, "", "  ")
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Failed to encode flow list: " + err.Error())
		return
	}

	// Set the response body to the JSON-encoded flow list
	ctx.SetBody(activeFlowsJSON)
}

// HandleFlowGraphSVG generates and serves an SVG image of the requested flow graph.
// @Summary Generate SVG graph of a flow
// @Description Generates and returns an SVG image visualization of the specified authentication flow
// @Tags Debug
// @Produce image/svg+xml
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param flow path string true "Flow name"
// @Success 200 {file} binary "SVG image of the flow graph"
// @Failure 400 {string} string "Bad request - missing flow parameter"
// @Failure 404 {string} string "Flow not found"
// @Failure 500 {string} string "Internal server error"
// @Router /{tenant}/{realm}/debug/{flow}/graph.svg [get]
func HandleFlowGraphSVG(ctx *fasthttp.RequestCtx) {

	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flowPath := ctx.UserValue("flow").(string)

	if flowPath == "" {
		// Return a bad request if the flow name is missing
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("Missing query flow")
		return
	}

	loadedRealm, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString("realm not found")
		return
	}

	// Look up the flow in the registry
	flow, ok := service.GetServices().FlowService.GetFlowForExecution(flowPath, loadedRealm)

	if !ok {
		// Return 404 if flow is not found
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString(fmt.Sprintf("Flow not found: %q", flowPath))
		return
	}

	// Generate the DOT representation for the flow graph
	dot, err := visual.RenderDOTGraph(flow.Definition)
	if err != nil {

		logger.GetRequestLogger(ctx).Warn().Err(err).Msg("Failed to generate SVG")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(fmt.Sprintf("Failed to generate SVG: %v", err))
		return
	}

	// Prepare the SVG output buffer
	var out bytes.Buffer

	// Use the `dot` command to convert the DOT string into an SVG image
	cmd := exec.Command("dot", "-Tsvg")
	cmd.Stdin = bytes.NewReader([]byte(dot)) // Pass the DOT data as input
	cmd.Stdout = &out                        // Capture the SVG output in the buffer

	// Run the command and check for errors
	if err := cmd.Run(); err != nil {
		// Return an internal server error if Graphviz fails
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(fmt.Sprintf("Failed to generate SVG: %v", err))
		return
	}

	// Set the content type to image/svg+xml and return the SVG data
	ctx.SetContentType("image/svg+xml")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(out.Bytes())
}
