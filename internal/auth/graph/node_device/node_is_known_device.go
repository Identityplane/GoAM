package node_device

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	DEFAULT_COOKIE_NAME = "device"

	CONFIG_COOKIE_NAME = "cookie_name"

	CONDITION_KNOWN_DEVICE        = "known_device"
	CONDITION_UNKNOWN_DEVICE      = "unknown_device"
	CONDITION_OIDC_REQUIRES_LOGIN = "oidc_requires_login"
)

var IsKnownDeviceNode = &model.NodeDefinition{
	Name:                 "isKnownDevice",
	PrettyName:           "Is Known Device",
	Description:          "Checks if the user is using a known device",
	Category:             "Device",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{},
	OutputContext:        []string{"device", "user"},
	PossibleResultStates: []string{CONDITION_KNOWN_DEVICE, CONDITION_UNKNOWN_DEVICE, CONDITION_OIDC_REQUIRES_LOGIN},
	CustomConfigOptions: map[string]string{
		CONFIG_COOKIE_NAME: "The name of the cookie to check for the device id (required)",
	},
	Run: RunIsKnownDeviceNode,
}

func RunIsKnownDeviceNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	now := time.Now()
	ctx := context.Background()

	// Get the device from the request
	device, attr, user, err := getDeviceFromRequest(state, services, node)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	if user == nil {
		return model.NewNodeResultWithCondition(CONDITION_UNKNOWN_DEVICE)
	}

	// Check if the device is still valid
	if device.LatestExpiry(now).Before(now) {
		return model.NewNodeResultWithCondition(CONDITION_UNKNOWN_DEVICE)
	}

	// Update the last activity timestamp and sessions
	device.Refresh(now)
	loa := device.CurrentLoa(now)
	state.Context["loa"] = strconv.Itoa(loa)

	// Check if we need to refresh the cookie
	// If the latest expiry is after the cookie expires we need to set the cookie again with a new value
	// This happens if the rolling session has been extended and the cookie expires before the session expires
	latestExpiry := device.LatestExpiry(now)
	if latestExpiry.After(device.CookieExpires) {
		device.CookieExpires = latestExpiry
		// create a new cookie with the device secret hash
		cookie := &http.Cookie{
			Name:     device.CookieName,
			Value:    device.DeviceSecretHash,
			Expires:  device.CookieExpires,
			SameSite: getSameSiteModeFromString(device.CookieSameSite),
			HttpOnly: device.CookieHttpOnly,
			Secure:   device.CookieSecure,
		}
		state.HttpAuthContext.AdditionalResponseCookies[device.CookieName] = *cookie
	}

	// Update the user attribute
	attr.Value = &device
	err = services.UserRepo.UpdateUserAttribute(ctx, attr)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// Set the authentication context with the user and device
	state.User = user
	state.Context["device"] = device.DeviceID

	// Get the latest authentication time
	// TODO, atm its unclear which LOA the auth time refers to, so we use just the latest auth time in general
	latestAuthTime := device.GetLatestAuthTime(now)

	// If there is a OIDC context we update the auth_time information
	if state.Oauth2SessionInformation != nil && state.Oauth2SessionInformation.AuthorizeRequest != nil {

		// If oidc has promot=login we return that oidc requires a login
		if state.Oauth2SessionInformation.AuthorizeRequest.Prompt == "login" {

			return model.NewNodeResultWithCondition(CONDITION_OIDC_REQUIRES_LOGIN)
		}

		// If the latest auth time is more than max_age ago we return that oidc requires a login
		if state.Oauth2SessionInformation.AuthorizeRequest.MaxAge != nil && latestAuthTime.Before(now.Add(-time.Duration(*state.Oauth2SessionInformation.AuthorizeRequest.MaxAge)*time.Second)) {
			return model.NewNodeResultWithCondition(CONDITION_OIDC_REQUIRES_LOGIN)
		}
	}

	// If there is a OIDC context we update the auth_time information in the context
	if state.Oauth2SessionInformation != nil {
		state.Oauth2SessionInformation.AuthTime = latestAuthTime
	}

	return model.NewNodeResultWithCondition(CONDITION_KNOWN_DEVICE)
}

func getDeviceFromRequest(state *model.AuthenticationSession, services *model.Repositories, node *model.GraphNode) (*model.DeviceAttributeValue, *model.UserAttribute, *model.User, error) {
	cookieName := node.CustomConfig[CONFIG_COOKIE_NAME]
	if cookieName == "" {
		cookieName = DEFAULT_COOKIE_NAME
	}

	if state.HttpAuthContext == nil {
		return nil, nil, nil, errors.New("http auth context is not set")
	}

	cookieValue := state.HttpAuthContext.RequestCookies[cookieName]
	if cookieValue == "" {
		return nil, nil, nil, nil
	}

	// Hash the device cookie and retreive the attribute if present
	deviceHash := lib.HashString(cookieValue)
	user, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypeDevice, deviceHash)
	if err != nil {
		return nil, nil, nil, err
	}

	// Check if the device is know for a user
	if user == nil {
		return nil, nil, nil, nil
	}

	// Get the right device attribute of the user
	devices, attributes, err := model.GetAttributes[model.DeviceAttributeValue](user, model.AttributeTypeDevice)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(devices) == 0 {
		return nil, nil, nil, nil
	}

	// find the right device attribute by the device hash
	var attribute *model.UserAttribute
	var device *model.DeviceAttributeValue
	for i, deviceAttribute := range attributes {
		if *deviceAttribute.Index == deviceHash {
			attribute = attributes[i]
			device = &devices[i]
			break
		}
	}

	return device, attribute, user, nil
}

func getSameSiteModeFromString(sameSite string) http.SameSite {
	switch sameSite {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
