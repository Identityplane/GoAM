package node_device

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/Identityplane/GoAM/pkg/model"
)

var ListAvailableUsersNode = &model.NodeDefinition{
	Name:            "listAvailableUsers",
	PrettyName:      "List Available Users",
	Description:     "Lists all available users for the device",
	Category:        "Device",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{},
	OutputContext:   []string{"device", "user", "loa"},
	PossiblePrompts: map[string]string{
		"userid": "The user id to continue with",
		"action": "The action to take",
	},
	PossibleResultStates: []string{"success", "failure", "none"},
	Run:                  RunListAvailableUsersNode,
}

func RunListAvailableUsersNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	now := time.Now()

	// If the user wants to login with another account, we return none
	if action, ok := input["action"]; ok && action == "another" {
		return model.NewNodeResultWithCondition("none")
	}

	// If a user id is provided in the input, we continue with that user
	if selectedUserID, ok := input["userid"]; ok && selectedUserID != "" {
		deviceInfos, err := getDevicesFromRequest(state, services, node)
		if err != nil {
			return model.NewNodeResultWithError(err)
		}

		var bestInfo *deviceInfo
		for _, info := range deviceInfos {
			if info.User.ID == selectedUserID {
				if bestInfo == nil || info.Device.CurrentLoa(now) > bestInfo.Device.CurrentLoa(now) {
					infoCopy := info
					bestInfo = &infoCopy
				}
			}
		}

		if bestInfo != nil {
			state.User = bestInfo.User
			state.Context["device"] = bestInfo.Device.DeviceID
			loa := bestInfo.Device.CurrentLoa(now)
			state.Context["loa"] = strconv.Itoa(loa)
			return model.NewNodeResultWithCondition("success")
		}

		return model.NewNodeResultWithTextError("selected user not found on this device")
	}

	// Otherwise we list all available users
	deviceInfos, err := getDevicesFromRequest(state, services, node)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	if len(deviceInfos) == 0 {
		return model.NewNodeResultWithCondition("none")
	}

	type UserInfo struct {
		UserID   string `json:"userid"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Loa      int    `json:"loa"`
	}

	userMap := make(map[string]UserInfo)
	for _, info := range deviceInfos {
		username := info.User.ID // Default to ID if no username attribute
		email := ""

		// Try to get username
		usernames := info.User.GetAttributesByType(model.AttributeTypeUsername)
		if len(usernames) > 0 {
			if val, ok := usernames[0].Value.(model.UsernameAttributeValue); ok {
				username = val.PreferredUsername
			} else if val, ok := usernames[0].Value.(*model.UsernameAttributeValue); ok {
				username = val.PreferredUsername
			}
		}

		// Try to get email
		emails := info.User.GetAttributesByType(model.AttributeTypeEmail)
		if len(emails) > 0 {
			if val, ok := emails[0].Value.(model.EmailAttributeValue); ok {
				email = val.Email
			} else if val, ok := emails[0].Value.(*model.EmailAttributeValue); ok {
				email = val.Email
			}
		}

		loa := info.Device.CurrentLoa(now)
		if existing, ok := userMap[info.User.ID]; !ok || loa > existing.Loa {
			userMap[info.User.ID] = UserInfo{
				UserID:   info.User.ID,
				Username: username,
				Email:    email,
				Loa:      loa,
			}
		}
	}

	var users []UserInfo
	for _, u := range userMap {
		users = append(users, u)
	}

	usersJson, err := json.Marshal(users)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	return model.NewNodeResultWithPrompts(map[string]string{
		"users": string(usersJson),
	})
}
