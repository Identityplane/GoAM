package node_device

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	CONDITION_DEVICE_ALREADY_EXISTS = "device_already_exists"
)

var AddKnownDeviceNode = &model.NodeDefinition{
	Name:                 "addKnownDevice",
	PrettyName:           "Add Known Device",
	Description:          "Adds a known device to the user",
	Category:             "Device",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{},
	OutputContext:        []string{"device", "user"},
	PossibleResultStates: []string{"success", "failure", CONDITION_DEVICE_ALREADY_EXISTS},
	CustomConfigOptions: map[string]string{
		CONFIG_COOKIE_NAME: "The name of the cookie to check for the device id",
	},
	Run: RunAddKnownDeviceNode,
}

func RunAddKnownDeviceNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	ctx := context.Background()
	now := time.Now()
	var err error

	cookieName := node.CustomConfig[CONFIG_COOKIE_NAME]
	if cookieName == "" {
		cookieName = DEFAULT_COOKIE_NAME
	}

	// if there is already an device id in the state context we return already exists
	if state.Context["device"] != "" {
		return model.NewNodeResultWithCondition(CONDITION_DEVICE_ALREADY_EXISTS)
	}

	// get the user agent header from the request
	userAgent := state.HttpAuthContext.RequestHeaders["user-agent"]
	deviceIp := state.HttpAuthContext.RequestIP

	// generate a new random device secret
	deviceId := lib.GenerateSecureSessionID()
	deviceSecret := lib.GenerateSecureSessionID()
	deviceSecretHash := lib.HashString(deviceSecret)

	// create a new device attribute value
	device := model.DeviceAttributeValue{
		DeviceID:         deviceId,
		DeviceSecretHash: deviceSecretHash,
		DeviceName:       userAgent,
		DeviceIP:         deviceIp,
		DeviceUserAgent:  userAgent,
		CookieName:       cookieName,
	}

	err = initSessions(&device, now, state, node)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// set the cookie expires
	device.CookieExpires = device.LatestExpiry(now)

	// create the device attribute
	attribute := &model.UserAttribute{
		Type:      model.AttributeTypeDevice,
		UserID:    state.User.ID,
		Value:     device,
		Index:     &deviceSecretHash,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// save the attribute to the user
	err = services.UserRepo.CreateUserAttribute(ctx, attribute)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// create a new cookie with the device secret hash
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    deviceSecret,
		Expires:  device.CookieExpires,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		HttpOnly: true,
		Secure:   true,
	}

	state.HttpAuthContext.AdditionalResponseCookies[cookieName] = *cookie

	// Set the device ID in the context
	state.Context["device"] = deviceId

	return model.NewNodeResultWithCondition("success")
}

func initSessions(device *model.DeviceAttributeValue, now time.Time, state *model.AuthenticationSession, node *model.GraphNode) error {

	// Get the level of assurance  from the authentication context

	loa := 0
	if state.Context["loa"] != "" {
		var err error
		loaStr := state.Context["loa"]
		loa, err = strconv.Atoi(loaStr)
		if err != nil {
			return fmt.Errorf("failed to convert loa to int: %w", err)
		}
	}

	mappings := getLoaToExpiryMapping()

	// We always init the LOA0 session
	device.SessionLoa0 = *model.InitSession(now, mappings[0])

	// We init the LOA1 session if the loa is 1 or 2
	if loa >= 1 && len(mappings) > 1 {
		device.SessionLoa1 = model.InitSession(now, mappings[1])
	}

	// We init the LOA2 session if the loa is 2
	if loa >= 2 && len(mappings) > 2 {
		device.SessionLoa2 = model.InitSession(now, mappings[2])
	}

	return nil
}

func getLoaToExpiryMapping() []model.LoaToExpiryMapping {
	return model.DEFAULT_LOA_TO_EXPIRY_MAPPINGS
}
