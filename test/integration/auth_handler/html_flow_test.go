package authhandler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Identityplane/GoAM/internal/service"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/Identityplane/GoAM/test/integration"
	"github.com/PuerkitoBio/goquery"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func TestHTMLFlow_SuccessLeadToSessionCookie(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")
	ctx := context.Background()

	// Create a new appliaction with simple-cookie and cookie settings
	application := &model.Application{
		ClientId:                   "test-app",
		AllowedGrants:              []string{model.GRANT_SIMPLE_AUTH_COOKIE},
		AllowedScopes:              []string{"openid", "profile", "email"},
		AllowedAuthenticationFlows: []string{"mock_success"},
		RedirectUris:               []string{"https://example.com/success"},
		AccessTokenLifetime:        3600,
		Settings: &model.ApplicationExtensionSettings{
			Cookie: &model.CookieSpecification{
				Name:          "access_token",
				Domain:        "example.com",
				Path:          "/",
				Secure:        true,
				HttpOnly:      true,
				SameSite:      "Strict",
				SessionExpiry: false,
			},
		},
	}

	err := service.GetServices().ApplicationService.CreateApplication("acme", "customers", *application)
	assert.NoError(t, err)

	t.Run("Success Leads to Session Cookie", func(t *testing.T) {
		resp := e.GET("/acme/customers/auth/mock-success").
			WithQuery("client_id", "test-app").
			WithQuery("scope", "profile email").
			Expect().
			Status(http.StatusSeeOther)

		// Validate the the cookie is set correctly
		header := resp.Header("Set-Cookie").Raw()
		assert.NotNil(t, header)

		cookie := resp.Cookie(application.Settings.Cookie.Name)
		cookie_raw := cookie.Raw()
		assert.NotNil(t, cookie_raw)

		cookie.Domain().IsEqual(application.Settings.Cookie.Domain)
		cookie.Name().IsEqual(application.Settings.Cookie.Name)
		cookie.MaxAge().IsEqual(time.Duration(application.AccessTokenLifetime) * time.Second)
		cookie.Path().IsEqual(application.Settings.Cookie.Path)
		assert.Equal(t, cookie.Raw().Secure, application.Settings.Cookie.Secure)
		assert.Equal(t, cookie.Raw().HttpOnly, application.Settings.Cookie.HttpOnly)
		assert.Equal(t, cookie.Raw().SameSite, http.SameSiteStrictMode)

		// Check if the access token is valid
		access_token := cookie.Raw().Value

		sesssion, err := service.GetServices().SessionsService.GetClientSessionByAccessToken(ctx, "acme", "customers", access_token)
		assert.NoError(t, err)
		assert.NotNil(t, sesssion)
		assert.Equal(t, application.ClientId, sesssion.ClientID)
		assert.Equal(t, "profile email", sesssion.Scope)
		assert.Equal(t, "simple-cookie", sesssion.GrantType)
		assert.NotNil(t, sesssion.UserID)

		// Validate the the redirection is also correct
		assert.Equal(t, "https://example.com/success", string(resp.Header("Location").Raw()))
	})
}

func TestHTMLFlow_UserRegistration(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	t.Run("Complete Registration Flow", func(t *testing.T) {
		var sessionCookie string

		// Step 1: GET request - expect HTML with username field
		t.Run("Step 1: GET - Username Field", func(t *testing.T) {
			// Make the request and capture both cookie and response
			req := e.GET("/acme/customers/auth/username-password-register")
			expect := req.Expect().Status(http.StatusOK)
			sessionCookie = expect.Cookie("session_id").Value().Raw()
			resp := expect.Body()

			// Parse HTML and validate
			doc := parseHTMLResponse(t, resp)
			assertInputFieldExists(t, doc, "text", "username")
			assertStepValue(t, doc, "askUsername")
		})

		// Step 2: POST username - expect HTML with password field
		t.Run("Step 2: POST Username - Password Field", func(t *testing.T) {
			resp := e.POST("/acme/customers/auth/username-password-register").
				WithHeader("Content-Type", "application/x-www-form-urlencoded").
				WithCookie("session_id", sessionCookie).
				WithFormField("step", "askUsername").
				WithFormField("username", "testuser").
				Expect().
				Status(http.StatusOK).
				Body()

			// Parse HTML and validate
			doc := parseHTMLResponse(t, resp)
			assertInputFieldExists(t, doc, "password", "password")
			assertStepValue(t, doc, "askPassword")
		})

		// Step 3: POST password - expect success message
		t.Run("Step 3: POST Password - Success Message", func(t *testing.T) {
			resp := e.POST("/acme/customers/auth/username-password-register").
				WithHeader("Content-Type", "application/x-www-form-urlencoded").
				WithCookie("session_id", sessionCookie).
				WithFormField("step", "askPassword").
				WithFormField("password", "testuser").
				Expect().
				Status(http.StatusOK).
				Body()

			// Parse HTML and validate
			doc := parseHTMLResponse(t, resp)
			assertSuccessMessage(t, doc, "Registration successful!")
		})
	})
}

// parseHTMLResponse parses the HTML response body and returns a goquery document
func parseHTMLResponse(t *testing.T, resp *httpexpect.String) *goquery.Document {
	htmlContent := resp.Raw()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}
	return doc
}

// assertInputFieldExists checks if an input field with the specified type and name exists
func assertInputFieldExists(t *testing.T, doc *goquery.Document, inputType, name string) {
	selector := fmt.Sprintf("input[type='%s'][name='%s']", inputType, name)
	inputField := doc.Find(selector)
	if inputField.Length() == 0 {
		t.Errorf("Expected to find input field with type='%s' and name='%s'", inputType, name)
	} else {
		t.Logf("Found %d input field(s) with type='%s' and name='%s'", inputField.Length(), inputType, name)
	}
}

// assertStepValue checks if the hidden step field has the expected value
func assertStepValue(t *testing.T, doc *goquery.Document, expectedStep string) {
	stepInput := doc.Find("input[type='hidden'][name='step']")
	if stepInput.Length() == 0 {
		t.Error("Expected to find hidden step field")
		return
	}

	stepValue, _ := stepInput.Attr("value")
	if stepValue != expectedStep {
		t.Errorf("Expected step value '%s', got '%s'", expectedStep, stepValue)
	} else {
		t.Logf("Step value is correct: '%s'", stepValue)
	}
}

// assertSuccessMessage checks if the success message is present in the response
func assertSuccessMessage(t *testing.T, doc *goquery.Document, expectedMessage string) {
	bodyText := doc.Find("body").Text()
	if !strings.Contains(bodyText, expectedMessage) {
		t.Errorf("Expected to find '%s' message in body. Got: %s", expectedMessage, bodyText)
	} else {
		t.Logf("Found success message: '%s'", expectedMessage)
	}
}
