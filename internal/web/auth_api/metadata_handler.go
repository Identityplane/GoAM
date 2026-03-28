package auth_api

import (
	"encoding/json"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/valyala/fasthttp"
)

// FlowInfo representing basic information about a flow
type FlowInfo struct {
	ID          string `json:"id"`
	Route       string `json:"route"`
	Description string `json:"description"`
}

// RealmInfo representing public realm settings
type RealmInfo struct {
	Name     string            `json:"name"`
	BaseUrl  string            `json:"baseUrl"`
	Settings map[string]string `json:"settings"`
}

// MetadataResponse represents the response for the metadata endpoint
type MetadataResponse struct {
	Flows []FlowInfo `json:"flows"`
	Realm RealmInfo  `json:"realm"`
}

// HandleMetadataRequest returns metadata about the realm
func HandleMetadataRequest(ctx *fasthttp.RequestCtx) {
	tenant := ctx.UserValue("tenant").(string)
	realm := ctx.UserValue("realm").(string)

	// Set JSON content type
	ctx.SetContentType("application/json")

	// Load realm
	loadedRealm, ok := service.GetServices().RealmService.GetRealm(tenant, realm)
	if !ok {
		sendErrorResponse(ctx, nil, fasthttp.StatusNotFound, "REALM_NOT_FOUND", "Realm not found", "")
		return
	}

	// Load flows
	flows, err := service.GetServices().FlowService.ListFlows(tenant, realm)
	if err != nil {
		sendErrorResponse(ctx, nil, fasthttp.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Could not list flows", "")
		return
	}

	// Filter active flows and convert to FlowInfo
	activeFlows := []FlowInfo{}
	for _, flow := range flows {
		if flow.Active {
			description := ""
			if flow.Definition != nil {
				description = flow.Definition.Description
			}
			activeFlows = append(activeFlows, FlowInfo{
				ID:          flow.Id,
				Route:       flow.Route,
				Description: description,
			})
		}
	}

	// Construct response
	response := MetadataResponse{
		Flows: activeFlows,
		Realm: RealmInfo{
			Name:     loadedRealm.Config.RealmName,
			BaseUrl:  loadedRealm.Config.BaseUrl,
			Settings: loadedRealm.Config.RealmSettings,
		},
	}

	// Send JSON response
	ctx.SetStatusCode(fasthttp.StatusOK)
	json.NewEncoder(ctx).Encode(response)
}
