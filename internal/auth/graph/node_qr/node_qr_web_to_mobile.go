package node_qr

import (
	"errors"
	"fmt"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	CONTEXT_KEY_QR_LOGIN_WEB_TO_MOBILE_SECRET = "qr_login_web_to_mobile_secret"
)

var QrWebToMobileNode = &model.NodeDefinition{
	Name:            "qrWebToMobile",
	PrettyName:      "QR Web to Mobile",
	Description:     "Generates a QR code for the user to scan with their mobile device to login. The user must be set in the context.",
	Category:        "QR",
	Type:            model.NodeTypeQueryWithLogic,
	RequiredContext: []string{"user"},
	OutputContext:   []string{"qr_code"},
	PossibleResultStates: []string{
		"user_not_found",
		"success",
		"abort",
		"action_token",
	},
	Run: RunQrWebToMobileNode,
}

func RunQrWebToMobileNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Get the base URL from context or extract from LoginUriBase
	loginUrl := state.LoginUriBase
	if loginUrl == "" {
		return model.NewNodeResultWithError(errors.New("login url not found in context"))
	}

	// Get the user from the context
	user, err := node_utils.LoadUserFromContext(state, services)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}
	if user == nil {
		return model.NewNodeResultWithError(errors.New("user not found in context"))
	}

	// Initialize the secondary session id and set the current node to initSecondary
	secondarySecret := state.InitSecondarySessionID()
	state.CurrentOnSecondaryDevice = "initSecondary"

	return model.NewNodeResultWithPrompts(map[string]string{
		"qr_code": getQrLoginUrl(loginUrl, secondarySecret),
	})
}

func getQrLoginUrl(baseUrl string, secret string) string {
	return fmt.Sprintf("%s?secondary=%s", baseUrl, secret)
}
