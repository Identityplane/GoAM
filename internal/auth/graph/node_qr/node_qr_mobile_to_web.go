package node_qr

import (
	"github.com/Identityplane/GoAM/pkg/model"
)

const (
	CONTEXT_KEY_QR_LOGIN_MOBILE_TO_WEB_SECRET = "qr_login_mobile_to_web_secret"
	CONTEXT_KEY_CONTINUE_ON_PRIMARY_DEVICE    = "continue_on_primary_device"
)

var QrMobileToWebNode = &model.NodeDefinition{
	Name:          "qrMobileToWeb",
	PrettyName:    "QR Mobile to Web",
	Description:   "Generates a QR code for the user to scan with their mobile device to login on the web.",
	Category:      "QR",
	Type:          model.NodeTypeQueryWithLogic,
	OutputContext: []string{"qr_code"},
	PossibleResultStates: []string{
		model.ResultStateSuccess,
		"abort",
	},
	Run: RunQrMobileToWebNode,
}

func RunQrMobileToWebNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	// Secondary device completed: advance primary without regenerating the QR URL.
	if state.Context[CONTEXT_KEY_CONTINUE_ON_PRIMARY_DEVICE] == "true" {

		state.Context[CONTEXT_KEY_CONTINUE_ON_PRIMARY_DEVICE] = ""
		return model.NewNodeResultWithCondition(model.ResultStateSuccess)
	}

	// Reuse persisted URL so polling does not rotate the secondary session id.
	qrCodeUrl := state.Context[CONTEXT_KEY_QR_LOGIN_MOBILE_TO_WEB_SECRET]
	if qrCodeUrl == "" {
		secondarySecret := state.InitSecondarySessionID()
		qrCodeUrl = getQrLoginUrl(state.LoginUriBase, secondarySecret)
		state.Context[CONTEXT_KEY_QR_LOGIN_MOBILE_TO_WEB_SECRET] = qrCodeUrl
	}

	return model.NewNodeResultWithPrompts(map[string]string{
		"qr_code": qrCodeUrl,
	})

}
