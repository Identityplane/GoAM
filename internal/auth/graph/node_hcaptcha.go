package graph

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/Identityplane/GoAM/internal/auth/repository"
	"github.com/Identityplane/GoAM/internal/model"
)

// HCaptchaVerifier defines the interface for hCaptcha verification
type HCaptchaVerifier interface {
	Verify(response, sitekey, secret string) bool
}

// DefaultHCaptchaVerifier implements HCaptchaVerifier using the hCaptcha API
type DefaultHCaptchaVerifier struct{}

var HcaptchaNode = &NodeDefinition{
	Name:                 "hcaptcha",
	Type:                 model.NodeTypeQueryWithLogic,
	RequiredContext:      []string{"hcaptcha"},
	PossiblePrompts:      map[string]string{"hcaptcha": "text"},
	OutputContext:        []string{},
	PossibleResultStates: []string{"success", "failure"},
	CustomConfigOptions:  []string{"hcaptcha_sitekey", "hcaptcha_secret"},
	Run:                  RunHcaptchaNode,
}

// Global verifier instance
var hcaptchaVerifier HCaptchaVerifier = &DefaultHCaptchaVerifier{}

// SetHCaptchaVerifier allows setting a custom verifier for testing
func SetHCaptchaVerifier(verifier HCaptchaVerifier) {
	hcaptchaVerifier = verifier
}

func RunHcaptchaNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *repository.Repositories) (*model.NodeResult, error) {
	response := input["hcaptcha"]

	// Check if custom config hcaptcha_sitekey is set
	hcaptchaSitekey := node.CustomConfig["hcaptcha_sitekey"]
	if hcaptchaSitekey == "" {
		return model.NewNodeResultWithError(errors.New("hcaptcha_sitekey is not set"))
	}

	// Check if custom config hcaptcha_secret is set
	hcaptchaSecret := node.CustomConfig["hcaptcha_secret"]
	if hcaptchaSecret == "" {
		return model.NewNodeResultWithError(errors.New("hcaptcha_secret is not set"))
	}

	if response == "" {
		return model.NewNodeResultWithPrompts(map[string]string{"hcaptcha": "text"})
	}

	// Verify hcaptcha response
	if !hcaptchaVerifier.Verify(response, hcaptchaSitekey, hcaptchaSecret) {
		return model.NewNodeResultWithCondition("failure")
	}

	return model.NewNodeResultWithCondition("success")
}

func (v *DefaultHCaptchaVerifier) Verify(response, sitekey, secret string) bool {
	// Create form data
	formData := url.Values{}
	formData.Set("secret", secret)
	formData.Set("response", response)
	formData.Set("sitekey", sitekey)

	// Make POST request to hCaptcha verification endpoint
	resp, err := http.PostForm("https://api.hcaptcha.com/siteverify", formData)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Parse JSON response
	var result struct {
		Success     bool     `json:"success"`
		ChallengeTs string   `json:"challenge_ts"`
		Hostname    string   `json:"hostname"`
		ErrorCodes  []string `json:"error-codes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	// Check if verification was successful
	return result.Success
}
