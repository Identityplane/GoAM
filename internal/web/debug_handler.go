package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goiam/internal"
	"goiam/internal/auth/graph/visual"
	"os/exec"

	"github.com/valyala/fasthttp"
)

// HandleListFlows responds with a list of available flow names and routes
func HandleListFlows(ctx *fasthttp.RequestCtx) {
	var flowList []map[string]string

	// Iterate over FlowRegistry to collect names and routes
	for _, flowWithRoute := range internal.FlowRegistry {
		flowList = append(flowList, map[string]string{
			"name":  flowWithRoute.Flow.Name,
			"route": flowWithRoute.Route,
		})
	}

	// Set the response type to JSON and send the flow list
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)

	if err := json.NewEncoder(ctx).Encode(flowList); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Failed to encode flow list: " + err.Error())
	}
}

// HandleFlowGraphPNG generates and serves a PNG image of the requested flow graph.
// Usage: GET /debug/flow/graph.png?flow=flow_name
func HandleFlowGraphPNG(ctx *fasthttp.RequestCtx) {
	// Get flow name from the query parameter
	flowName := string(ctx.QueryArgs().Peek("flow"))
	if flowName == "" {
		// Return a bad request if flow name is missing
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("Missing query parameter: ?flow=")
		return
	}

	// Look up the flow in the registry
	flowWithRoute, ok := internal.FlowRegistry[flowName]
	if !ok || flowWithRoute.Flow == nil {
		// Return 404 if flow is not found
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString(fmt.Sprintf("Flow not found: %q", flowName))
		return
	}

	// Generate the DOT representation for the flow graph
	dot := visual.RenderDOTGraph(flowWithRoute.Flow)

	// Prepare the PNG output buffer
	var out bytes.Buffer

	// Use the `dot` command to convert the DOT string into a PNG image
	cmd := exec.Command("dot", "-Tpng")
	cmd.Stdin = bytes.NewReader([]byte(dot)) // Pass the DOT data as input
	cmd.Stdout = &out                        // Capture the PNG output in the buffer

	// Run the command and check for errors
	if err := cmd.Run(); err != nil {
		// Return an internal server error if Graphviz fails
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(fmt.Sprintf("Failed to generate PNG: %v", err))
		return
	}

	// Set the content type to image/png and return the image data
	ctx.SetContentType("image/png")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(out.Bytes())
}

// HandleFlowGraphSVG generates and serves an SVG image of the requested flow graph.
func HandleFlowGraphSVG(ctx *fasthttp.RequestCtx) {
	// Get flow name from the query parameter
	flowName := string(ctx.QueryArgs().Peek("flow"))
	if flowName == "" {
		// Return a bad request if the flow name is missing
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("Missing query parameter: ?flow=")
		return
	}

	// Look up the flow in the registry
	flowWithRoute, ok := internal.FlowRegistry[flowName]
	if !ok || flowWithRoute.Flow == nil {
		// Return 404 if the flow is not found
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString(fmt.Sprintf("Flow not found: %q", flowName))
		return
	}

	// Generate the DOT representation for the flow graph
	dot := visual.RenderDOTGraph(flowWithRoute.Flow)

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
