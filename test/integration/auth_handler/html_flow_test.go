package authhandler

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"
	"github.com/PuerkitoBio/goquery"
	"github.com/gavv/httpexpect/v2"
)

func TestHTMLFlow_Integration(t *testing.T) {
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
