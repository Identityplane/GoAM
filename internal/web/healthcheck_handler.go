package web

import (
	"encoding/json"
	"goiam/internal/service"
	"reflect"
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

	allRealms, err := service.GetServices().RealmService.GetAllRealms()

	if err != nil {
		ready["Realms"] = "cannot load realms"
		isReady = false
	}

	if len(allRealms) == 0 {
		ready["Realms"] = "not at least 1 realm available"
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
	// Create the main info structure
	info := map[string]interface{}{
		"name":       "GoIAM",
		"go_version": runtime.Version(),
	}

	// list available services and their implementaitons with service.GetServices()
	services := service.GetServices()
	servicesInfo := make(map[string]string)

	// Use reflection to get the type information
	svcValue := reflect.ValueOf(services).Elem()
	svcType := svcValue.Type()

	for i := 0; i < svcValue.NumField(); i++ {
		field := svcValue.Field(i)
		fieldType := svcType.Field(i)

		// Get the concrete type name
		implType := "nil"
		if !field.IsNil() {
			implType = reflect.TypeOf(field.Interface()).String()
		}

		servicesInfo[fieldType.Name] = implType
	}

	// Add services info directly to the main structure
	info["services"] = servicesInfo

	// Get database implementations
	databases := service.Databases
	databasesInfo := make(map[string]string)

	// Use reflection to get the type information for databases
	dbValue := reflect.ValueOf(databases).Elem()
	dbType := dbValue.Type()

	for i := 0; i < dbValue.NumField(); i++ {
		field := dbValue.Field(i)
		fieldType := dbType.Field(i)

		// Get the concrete type name
		implType := "nil"
		if !field.IsNil() {
			implType = reflect.TypeOf(field.Interface()).String()
		}

		databasesInfo[fieldType.Name] = implType
	}

	// Add databases info directly to the main structure
	info["databases"] = databasesInfo

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	jsonData, _ := json.MarshalIndent(info, "", "  ")
	ctx.SetBody(jsonData)
}
