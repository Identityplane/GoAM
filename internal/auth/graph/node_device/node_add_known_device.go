package node_device

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	CONDITION_DEVICE_ALREADY_EXISTS = "device_already_exists"

	DEFAULT_SESSION_DURATION      = 3600 * 24 * 30
	DEFAULT_SESSION_REFRESH_AFTER = 3600 * 24 * 30 / 2

	CONFIG_SESSION_DURATION      = "session_duration"
	CONFIG_SESSION_REFRESH_AFTER = "session_refresh_after"
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

	// Get the config for the node
	sessionDurationInt := DEFAULT_SESSION_DURATION
	sessionDuration := node.CustomConfig[CONFIG_SESSION_DURATION]
	if sessionDuration != "" {
		sessionDurationInt, err = strconv.Atoi(sessionDuration)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}
	}

	sessionRefreshAfterInt := DEFAULT_SESSION_REFRESH_AFTER
	sessionRefreshAfter := node.CustomConfig[CONFIG_SESSION_REFRESH_AFTER]
	if sessionRefreshAfter != "" {
		sessionRefreshAfterInt, err = strconv.Atoi(sessionRefreshAfter)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}
	}

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

	expiry := now.Add(time.Duration(sessionDurationInt) * time.Second)

	// create a new device attribute value
	device := model.DeviceAttributeValue{
		DeviceID:            deviceId,
		DeviceSecretHash:    deviceSecretHash,
		DeviceName:          userAgent,
		DeviceIP:            deviceIp,
		DeviceUserAgent:     userAgent,
		SessionDuration:     sessionDurationInt,
		SessionExpiry:       &expiry,
		SessionRefreshAfter: sessionRefreshAfterInt,
		SessionFirstLogin:   &now,
		SessionLastActivity: &now,
		CookieName:          cookieName,
	}

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
		Expires:  expiry,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}

	state.HttpAuthContext.AdditionalResponseCookies[cookieName] = *cookie

	// Set the device ID in the context
	state.Context["device"] = deviceId

	return model.NewNodeResultWithCondition("success")
}
