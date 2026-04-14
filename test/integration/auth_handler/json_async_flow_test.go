package authhandler

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Identityplane/GoAM/test/integration"
	"github.com/google/uuid"
)

func TestJSONAsyncFlow_QRMobileToWeb_ResumeWithSecondaryID(t *testing.T) {
	e := integration.SetupIntegrationTest(t, "")

	const tenant = "acme"
	const realm = "customers"
	const flowRoute = "qr-mobile-to-web"

	t.Run("init flow -> extract secondary -> resume via secondary id -> primary poll completes", func(t *testing.T) {
		// 1) Init flow (primary device)
		initHTTP := e.GET("/" + tenant + "/" + realm + "/api/v1/" + flowRoute).
			WithQuery("init", "true").
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK)

		initResp := initHTTP.JSON()
		primaryCookie := initHTTP.Cookie("session_id").Value().Raw()
		if primaryCookie == "" {
			t.Fatal("expected session_id cookie from init")
		}

		initResp.Object().Value("executionId").String().NotEmpty()
		primarySessionID := initResp.Object().Value("sessionId").String().Raw()
		initResp.Object().HasValue("currentNode", "qr")
		initResp.Object().Value("prompts").Object().ContainsKey("qr_code")

		qrURL := initResp.Object().Value("prompts").Object().Value("qr_code").String().Raw()
		secondaryID := extractSecondaryID(t, qrURL)
		if secondaryID == "" {
			t.Fatalf("expected secondary id from qr url, got empty (qr_url=%q)", qrURL)
		}

		// 2) Resume using secondary id (secondary device)
		resumeResp := e.POST("/" + tenant + "/" + realm + "/api/v1/").
			WithHeader("Content-Type", "application/json").
			WithJSON(map[string]string{"sessionId": secondaryID}).
			Expect().
			Status(http.StatusOK).
			JSON()

		// The resume endpoint should echo a valid sessionId.
		resumeResp.Object().Value("executionId").String().NotEmpty()
		resumeResp.Object().Value("sessionId").String().IsEqual(secondaryID)
		resumeResp.Object().HasValue("currentNode", "askUserID")
		resumeResp.Object().HasValue("isSecondaryDevice", true)
		resumeResp.Object().Value("prompts").Object().HasValue("user_id", "text")

		executionID := resumeResp.Object().Value("executionId").String().Raw()

		// 3) Continue: submit user_id (unknown id -> create user, then "continue on other device" message)
		newUserID := uuid.NewString()
		cont := FlowRequest{
			SessionID:   secondaryID,
			CurrentNode: "askUserID",
			Responses: map[string]string{
				"user_id": newUserID,
			},
		}

		afterAsk := e.POST("/"+tenant+"/"+realm+"/api/v1/"+flowRoute).
			WithHeader("Content-Type", "application/json").
			WithJSON(cont).
			Expect().
			Status(http.StatusOK).
			JSON()

		afterAsk.Object().Value("executionId").String().IsEqual(executionID)
		afterAsk.Object().Value("sessionId").String().IsEqual(secondaryID)
		afterAsk.Object().HasValue("currentNode", "continueOnOther")
		afterAsk.Object().HasValue("currentNodeType", "continueOnOtherDevice")
		afterAsk.Object().Value("prompts").Object().ContainsKey("message")

		// 3b) Acknowledge — sets primary flag and completes secondary success
		ack := FlowRequest{
			SessionID:   secondaryID,
			CurrentNode: "continueOnOther",
			Responses: map[string]string{
				"ack": "true",
			},
		}

		done := e.POST("/"+tenant+"/"+realm+"/api/v1/"+flowRoute).
			WithHeader("Content-Type", "application/json").
			WithJSON(ack).
			Expect().
			Status(http.StatusOK).
			JSON()

		done.Object().Value("executionId").String().IsEqual(executionID)
		done.Object().Value("sessionId").String().IsEqual(secondaryID)
		done.Object().HasValue("currentNode", "successResult")
		done.Object().Value("result").Object().HasValue("success", true)

		// 4) Primary device: poll until graph advances (same session as init)
		poll := e.GET("/" + tenant + "/" + realm + "/api/v1/" + flowRoute).
			WithQuery("continue", "true").
			WithQuery("poll", "true").
			WithCookie("session_id", primaryCookie).
			WithHeader("Accept", "application/json").
			Expect().
			Status(http.StatusOK).
			JSON()

		poll.Object().Value("sessionId").String().IsEqual(primarySessionID)
		poll.Object().HasValue("currentNode", "webSuccess")
		poll.Object().Value("result").Object().HasValue("success", true)
	})
}

func extractSecondaryID(t *testing.T, qrURL string) string {
	t.Helper()

	parsed, err := url.Parse(qrURL)
	if err == nil {
		if v := parsed.Query().Get("secondary"); v != "" {
			return v
		}
	}

	// Fallback for non-absolute / odd URLs
	if i := strings.Index(qrURL, "secondary="); i >= 0 {
		v := qrURL[i+len("secondary="):]
		if j := strings.IndexByte(v, '&'); j >= 0 {
			v = v[:j]
		}
		if j := strings.IndexByte(v, '#'); j >= 0 {
			v = v[:j]
		}
		return v
	}

	return ""
}

