package admin_api

import (
	"encoding/json"
	"goiam/internal/auth/graph"
	"goiam/internal/service"
	"net/http"

	"github.com/valyala/fasthttp"
)

// HandleListFlows returns all flows for a realm
// @Summary List all flows
// @Description Returns a list of all flows in a realm
// @Tags Flows
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 200 {array} FlowWithYAML
// @Failure 404 {string} string "Realm not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/flows [get]
func HandleListFlows(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	flows, err := service.GetServices().FlowService.ListFlows(tenant, realm)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to list flows: " + err.Error(),
		})
		return
	}

	// Convert flows to API format
	apiFlows := make([]FlowWithYAML, len(flows))
	for i, flow := range flows {
		yaml, err := service.ConvertFlowToYAML(flow.Definition)
		if err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			ctx.SetContentType("application/json")
			_ = json.NewEncoder(ctx).Encode(map[string]string{
				"error": "Failed to convert flow to YAML: " + err.Error(),
			})
			return
		}
		apiFlows[i] = FlowWithYAML{
			FlowId: flow.Id,
			Route:  flow.Route,
			Realm:  flow.Realm,
			Tenant: flow.Tenant,
			YAML:   yaml,
		}
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(apiFlows, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to marshal response: " + err.Error(),
		})
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleGetFlow returns a specific flow
// @Summary Get a flow
// @Description Returns a specific flow configuration
// @Tags Flows
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param flow path string true "Flow ID"
// @Success 200 {object} FlowWithYAML
// @Failure 404 {string} string "Flow not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/flows/{flow} [get]
func HandleGetFlow(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flow := ctx.UserValue("flow").(string)

	flowWithRoute, ok := service.GetServices().FlowService.GetFlowById(tenant, realm, flow)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Flow not found",
		})
		return
	}

	// Convert flow to YAML
	yaml, err := service.ConvertFlowToYAML(flowWithRoute.Definition)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to convert flow to YAML: " + err.Error(),
		})
		return
	}

	// Create API response
	apiFlow := FlowWithYAML{
		FlowId: flowWithRoute.Id,
		Route:  flowWithRoute.Route,
		Realm:  flowWithRoute.Realm,
		Tenant: flowWithRoute.Tenant,
		YAML:   yaml,
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(apiFlow, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to marshal response: " + err.Error(),
		})
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleCreateFlow creates a new flow
// @Summary Create a flow
// @Description Creates a new flow configuration
// @Tags Flows
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param flow path string true "Flow ID"
// @Param request body FlowWithYAML true "Flow configuration"
// @Success 201 {object} FlowWithYAML
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/flows/{flow} [post]
func HandleCreateFlow(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flow := ctx.UserValue("flow").(string)

	// Parse JSON request body into FlowWithYAML
	var apiFlow FlowWithYAML
	if err := json.Unmarshal(ctx.PostBody(), &apiFlow); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Invalid JSON content: " + err.Error(),
		})
		return
	}

	// Set route from URL parameter
	apiFlow.Route = flow
	apiFlow.Tenant = tenant
	apiFlow.Realm = realm
	apiFlow.FlowId = flow

	// Convert YAML to FlowWithRoute
	flowWithRoute, err := service.LoadFlowFromYAMLString(apiFlow.YAML)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Invalid YAML content: " + err.Error(),
		})
		return
	}

	// Check if flow definition is valid
	if err := graph.ValidateFlowDefinition(flowWithRoute.Definition); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Invalid flow definition: " + err.Error(),
		})
		return
	}

	flowWithRoute.Route = apiFlow.Route
	flowWithRoute.Tenant = apiFlow.Tenant
	flowWithRoute.Realm = apiFlow.Realm
	flowWithRoute.Id = apiFlow.FlowId

	if err := service.GetServices().FlowService.CreateFlow(tenant, realm, flowWithRoute); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to create flow: " + err.Error(),
		})
		return
	}

	// Return created flow
	jsonData, err := json.MarshalIndent(apiFlow, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to marshal response: " + err.Error(),
		})
		return
	}

	ctx.SetStatusCode(http.StatusCreated)
	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleUpdateFlow updates an existing flow
// @Summary Update a flow
// @Description Updates an existing flow configuration
// @Tags Flows
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param flow path string true "Flow ID"
// @Param request body FlowPatch true "Flow patch object with fields to update"
// @Success 200 {object} FlowWithYAML
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Flow not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/flows/{flow} [patch]
func HandleUpdateFlow(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flow := ctx.UserValue("flow").(string)

	// Get existing flow
	existingFlow, ok := service.GetServices().FlowService.GetFlowById(tenant, realm, flow)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Flow not found",
		})
		return
	}

	// Parse patch request body
	var patch FlowPatch
	if err := json.Unmarshal(ctx.PostBody(), &patch); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Invalid JSON content: " + err.Error(),
		})
		return
	}

	// Apply patch updates
	if patch.Route != nil {
		existingFlow.Route = *patch.Route
	}
	if patch.YAML != nil {
		flowDef, err := service.LoadFlowFromYAMLString(*patch.YAML)
		if err != nil {
			ctx.SetStatusCode(http.StatusBadRequest)
			ctx.SetContentType("application/json")
			_ = json.NewEncoder(ctx).Encode(map[string]string{
				"error": "Invalid YAML content: " + err.Error(),
			})
			return
		}
		existingFlow.Definition = flowDef.Definition

		// Check if flow definition is valid
		if err := graph.ValidateFlowDefinition(flowDef.Definition); err != nil {
			ctx.SetStatusCode(http.StatusBadRequest)
			ctx.SetContentType("application/json")
			_ = json.NewEncoder(ctx).Encode(map[string]string{
				"error": "Invalid flow definition: " + err.Error(),
			})
			return
		}
	}

	// Update flow by creating a new one with the same route
	if err := service.GetServices().FlowService.CreateFlow(tenant, realm, existingFlow); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to update flow: " + err.Error(),
		})
		return
	}

	// Convert updated flow to YAML
	yaml, err := service.ConvertFlowToYAML(existingFlow.Definition)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to convert flow to YAML: " + err.Error(),
		})
		return
	}

	// Create API response
	apiFlow := FlowWithYAML{
		FlowId: existingFlow.Id,
		Route:  existingFlow.Route,
		Realm:  existingFlow.Realm,
		Tenant: existingFlow.Tenant,
		YAML:   yaml,
	}

	// Return updated flow
	jsonData, err := json.MarshalIndent(apiFlow, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to marshal response: " + err.Error(),
		})
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}

// HandleDeleteFlow deletes a flow
// @Summary Delete a flow
// @Description Deletes an existing flow
// @Tags Flows
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param flow path string true "Flow ID"
// @Success 204
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/flows/{flow} [delete]
func HandleDeleteFlow(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flow := ctx.UserValue("flow").(string)

	// Get existing flow
	_, ok := service.GetServices().FlowService.GetFlowById(tenant, realm, flow)
	if !ok {
		ctx.SetStatusCode(http.StatusNoContent)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Flow not found",
		})
		return
	}

	// Delete flow
	err := service.GetServices().FlowService.DeleteFlow(tenant, realm, flow)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to delete flow: " + err.Error(),
		})
		return
	}

	ctx.SetStatusCode(http.StatusOK)
}

// HandleListNodes returns all available node definitions
// @Summary List all nodes
// @Description Returns a list of all available node definitions in the system
// @Tags Nodes
// @Accept json
// @Produce json
// @Success 200 {array} NodeInfo
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/nodes [get]
func HandleListNodes(ctx *fasthttp.RequestCtx) {
	// Get all node definitions and convert to API format
	nodes := make([]NodeInfo, 0, len(graph.NodeDefinitions))
	for _, node := range graph.NodeDefinitions {
		nodes = append(nodes, NodeInfo{
			Name:                 node.Name,
			Type:                 string(node.Type),
			PossibleResultStates: node.PossibleResultStates,
		})
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to marshal response: " + err.Error(),
		})
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetBody(jsonData)
}
