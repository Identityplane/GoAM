package admin_api

import (
	"encoding/json"
	"fmt"
	"goiam/internal/auth/graph"
	"goiam/internal/model"
	"goiam/internal/service"
	"net/http"

	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"
)

// HandleListFlows returns all flows for a realm
// @Summary List all flows
// @Description Returns a list of all flows in a realm
// @Tags Flows
// @Accept json
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Success 200 {array} model.Flow
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

	// Ensure flows is never nil
	if flows == nil {
		flows = []model.Flow{}
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(flows, "", "  ")
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
// @Success 200 {object} model.Flow
// @Failure 404 {string} string "Flow not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/flows/{flow} [get]
func HandleGetFlow(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flowId := ctx.UserValue("flow").(string)

	flow, ok := service.GetServices().FlowService.GetFlowById(tenant, realm, flowId)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Flow not found",
		})
		return
	}

	// Marshal response to JSON with pretty printing
	jsonData, err := json.MarshalIndent(flow, "", "  ")
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
// @Param request body model.Flow true "Flow configuration"
// @Success 201 {object} model.Flow
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/flows/{flow} [post]
func HandleCreateFlow(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flowId := ctx.UserValue("flow").(string)

	// Parse JSON request body
	var flow model.Flow
	if err := json.Unmarshal(ctx.PostBody(), &flow); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Invalid JSON content: " + err.Error(),
		})
		return
	}

	// Set route from URL parameter
	flow.Route = flowId
	flow.Tenant = tenant
	flow.Realm = realm
	flow.Id = flowId

	if err := service.GetServices().FlowService.CreateFlow(tenant, realm, flow); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to create flow: " + err.Error(),
		})
		return
	}

	// Return created flow
	jsonData, err := json.MarshalIndent(flow, "", "  ")
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
// @Success 200 {object} model.Flow
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

	// Update flow by creating a new one with the same route
	if err := service.GetServices().FlowService.CreateFlow(tenant, realm, *existingFlow); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to update flow: " + err.Error(),
		})
		return
	}

	// Return updated flow
	jsonData, err := json.MarshalIndent(existingFlow, "", "  ")
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
			Use:                  node.Name,
			Type:                 string(node.Type),
			PrettyName:           node.Name, // TODO: Add pretty name
			Category:             "",        // TODO: Add category
			RequiredContext:      node.RequiredContext,
			OutputContext:        node.OutputContext,
			PossibleResultStates: node.PossibleResultStates,
			Description:          "",  // TODO: Add description
			CustomConfigOptions:  nil, // TODO: Add custom config options
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

// HandleGetFlowDefintion returns the flow definition for a given flow id as yaml
// @Summary Get a flow definition
// @Description Returns the flow definition for a given flow id as yaml
// @Tags Flows
// @Accept json
// @Produce text/yaml
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param flow path string true "Flow ID"
// @Success 200 {string} string "Flow definition as YAML"
// @Failure 404 {string} string "Flow not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/flows/{flow}/definition [get]
func HandleGetFlowDefintion(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flowId := ctx.UserValue("flow").(string)

	flow, ok := service.GetServices().FlowService.GetFlowById(tenant, realm, flowId)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Flow not found",
		})
		return
	}

	if flow.Definition == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Flow definition not found",
		})
		return
	}

	ctx.SetContentType("text/yaml")
	ctx.SetBody([]byte(flow.DefintionYaml))
}

// HandleValidateFlowDefinition handles the validation of a YAML flow definition
// @Summary Validate a flow definition
// @Description Validates a YAML flow definition
// @Tags Flows
// @Accept text/yaml
// @Produce json
// @Param request body string true "Flow definition as YAML"
// @Success 200 {object} map[string]interface{} "Validation results"
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/flows/validate [post]
func HandleValidateFlowDefinition(ctx *fasthttp.RequestCtx) {
	// Get the flow service from the context
	flowService := service.GetServices().FlowService

	// Get the YAML content from the request body
	content := string(ctx.PostBody())
	if content == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "YAML content is required in the request body",
		})
		return
	}

	// Validate the flow definition
	validationErrors, err := flowService.ValidateFlowDefinition(content)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": fmt.Sprintf("failed to validate flow definition: %v", err),
		})
		return
	}

	if validationErrors == nil {
		validationErrors = []service.FlowLintError{}
	}

	// Return the validation results
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	_ = json.NewEncoder(ctx).Encode(map[string]interface{}{
		"valid":  len(validationErrors) == 0,
		"errors": validationErrors,
	})
}

// HandlePutFlowDefintion creates or updates a yaml flow defintion for a flow
// @Summary Create or update a flow definition
// @Description Creates or updates a flow definition for a given flow ID
// @Tags Flows
// @Accept text/yaml
// @Produce json
// @Param tenant path string true "Tenant ID"
// @Param realm path string true "Realm ID"
// @Param flow path string true "Flow ID"
// @Param request body string true "Flow definition as YAML"
// @Success 200 {object} model.Flow
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /admin/{tenant}/{realm}/flows/{flow}/definition [put]
func HandlePutFlowDefintion(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)
	flowId := ctx.UserValue("flow").(string)

	// Get existing flow
	existingFlow, ok := service.GetServices().FlowService.GetFlowById(tenant, realm, flowId)
	if !ok {
		ctx.SetStatusCode(http.StatusNotFound)
		ctx.SetContentType("application/json")

		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Flow not found",
		})
		return
	}

	yamlDefinition := string(ctx.PostBody())

	// Validate the flow definition
	validationErrors, err := service.GetServices().FlowService.ValidateFlowDefinition(yamlDefinition)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": fmt.Sprintf("failed to validate flow definition: %v", err),
		})
		return
	}

	if validationErrors != nil && len(validationErrors) > 0 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Invalid flow definition: " + validationErrors[0].Message,
		})
		return
	}

	// Unmarsahl the yaml and update the flow definition
	var flowDef model.FlowDefinition
	if err := yaml.Unmarshal([]byte(yamlDefinition), &flowDef); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to unmarshal flow definition: " + err.Error(),
		})
		return
	}

	existingFlow.Definition = &flowDef
	existingFlow.DefintionYaml = yamlDefinition

	// Update flow by creating a new one with the same route
	if err := service.GetServices().FlowService.CreateFlow(tenant, realm, *existingFlow); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.SetContentType("application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"error": "Failed to update flow: " + err.Error(),
		})
		return
	}

	// Return 200 OK
	ctx.SetStatusCode(http.StatusOK)
}
