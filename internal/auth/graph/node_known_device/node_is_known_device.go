package node_device

import (
	"context"
	"fmt"
	"time"

	"github.com/Identityplane/GoAM/internal/lib"
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	DEFAULT_COOKIE_NAME = "session"

	CONFIG_COOKIE_NAME = "cookie_name"

	CONDITION_KNOWN_DEVICE   = "known_device"
	CONDITION_UNKNOWN_DEVICE = "unknown_device"
)

var IsKnownDeviceNode = &model.NodeDefinition{
	Name:                 "isKnownDevice",
	PrettyName:           "Is Known Device",
	Description:          "Checks if the user is using a known device",
	Category:             "Device",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{},
	OutputContext:        []string{"device", "user"},
	PossibleResultStates: []string{CONDITION_KNOWN_DEVICE, CONDITION_UNKNOWN_DEVICE},
	CustomConfigOptions: map[string]string{
		CONFIG_COOKIE_NAME: "The name of the cookie to check for the device id (required)",
	},
	Run: RunIsKnownDeviceNode,
}

func RunIsKnownDeviceNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

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
	now := time.Now()
	if device.SessionExpiry.Before(now) {
		return model.NewNodeResultWithCondition(CONDITION_UNKNOWN_DEVICE)
	}

	// Update the last activity timestamp
	device.SessionLastActivity = &now

	// Update the user attribute
	attr.Value = &device
	err = services.UserRepo.UpdateUserAttribute(ctx, attr)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	// Set the authentication context with the user and device
	state.User = user
	state.Context["device"] = device.DeviceID

	return model.NewNodeResultWithCondition(CONDITION_KNOWN_DEVICE)
}

func getDeviceFromRequest(state *model.AuthenticationSession, services *model.Repositories, node *model.GraphNode) (*model.DeviceAttributeValue, *model.UserAttribute, *model.User, error) {
	cookieName := node.CustomConfig[CONFIG_COOKIE_NAME]
	if cookieName == "" {
		cookieName = DEFAULT_COOKIE_NAME
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
	for _, deviceAttribute := range attributes {
		if *deviceAttribute.Index == deviceHash {
			attribute = attributes[0]
			break
		}
	}

	device, ok := attribute.Value.(model.DeviceAttributeValue)
	if !ok {
		return nil, nil, nil, fmt.Errorf("attribute value is not of type DeviceAttributeValue")
	}

	return &device, attribute, user, nil
}
