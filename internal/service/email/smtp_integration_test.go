package email

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Identityplane/GoAM/pkg/model"
)

// TestMailpitSMTP_OTPEmail spins up Mailpit (SMTP + HTTP API), sends an OTP template
// email through DefaultEmailService, and asserts the message was captured.
//
// Skipped with -short (requires Docker). Mirrors the pattern used for Postgres testcontainers.
func TestMailpitSMTP_OTPEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Mailpit SMTP test in short mode (requires Docker)")
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "axllent/mailpit:v1.21",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("1025/tcp"),
			wait.ForListeningPort("8025/tcp"),
		).WithDeadline(60 * time.Second),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = c.Terminate(context.Background())
	})

	host, err := c.Host(ctx)
	require.NoError(t, err)
	smtpPort, err := c.MappedPort(ctx, "1025")
	require.NoError(t, err)
	httpPort, err := c.MappedPort(ctx, "8025")
	require.NoError(t, err)

	smtpHost := host
	if smtpHost == "localhost" {
		smtpHost = "127.0.0.1"
	}

	portNum, err := strconv.Atoi(smtpPort.Port())
	require.NoError(t, err)

	svc := NewDefaultEmailService(SmtpConfig{
		Host:           smtpHost,
		Port:           portNum,
		Username:       "",
		Password:       "",
		FromEmail:      "goam-test@local.test",
		FromName:       "GoAM Test",
		UseImplicitTLS: false,
	})

	const otp = "424242"
	err = svc.SendEmail("tenant", "realm", &model.SendEmailParams{
		Template: model.EmailTemplateOTP,
		To:       []model.EmailAddress{{Email: "recipient@local.test", Name: "Recipient"}},
		Params:   map[string]any{"otp": otp},
	})
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, httpPort.Port())
	waitForMailpitMessage(t, baseURL, "Your verification code", otp)
}

func waitForMailpitMessage(t *testing.T, baseURL, wantSubjectSubstring, wantBodySubstring string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	client := &http.Client{Timeout: 5 * time.Second}

	for time.Now().Before(deadline) {
		msgID := pollMailpitMessageIDBySubject(t, client, baseURL, wantSubjectSubstring)
		if msgID != "" && mailpitMessageContains(t, client, baseURL, msgID, wantBodySubstring) {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for Mailpit message subject=%q body=%q", wantSubjectSubstring, wantBodySubstring)
}

func pollMailpitMessageIDBySubject(t *testing.T, client *http.Client, baseURL, wantSubjectSubstring string) string {
	t.Helper()
	resp, err := client.Get(baseURL + "/api/v1/messages?limit=20")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		return ""
	}
	msgs, _ := root["messages"].([]any)
	for _, m := range msgs {
		mm, ok := m.(map[string]any)
		if !ok {
			continue
		}
		subj, _ := mm["Subject"].(string)
		if !strings.Contains(subj, wantSubjectSubstring) {
			continue
		}
		if id, ok := mm["ID"].(string); ok && id != "" {
			return id
		}
	}
	return ""
}

func mailpitMessageContains(t *testing.T, client *http.Client, baseURL, id, want string) bool {
	t.Helper()
	resp, err := client.Get(baseURL + "/api/v1/message/" + id)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return strings.Contains(string(raw), want)
}

// TestManualSendSMTP_FromEnvFile sends one real email using repo-root .env and explicit opt-in.
//
// Requirements:
//   - Do not run with -short.
//   - Set MANUAL_SMTP_TEST=1 (in the shell or in .env).
//   - Repo-root .env should define SMTP (e.g. GOAM_SMTP_HOST, GOAM_SMTP_PORT, GOAM_SMTP_FROM_EMAIL,
//     GOAM_SMTP_USERNAME / GOAM_SMTP_PASSWORD if your relay needs auth) and TEST_EMAIL_TO
//     (recipient address).
//
// Example:
//
//	MANUAL_SMTP_TEST=1 go test ./internal/service/email -run TestManualSendSMTP_FromEnvFile -count=1 -v
func TestManualSendSMTP_FromEnvFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping manual SMTP test in short mode")
	}

	root := findRepoRoot(t)
	envPath := filepath.Join(root, ".env")
	if _, err := os.Stat(envPath); err != nil {
		t.Skipf("no .env at %s: %v", envPath, err)
	}
	if err := godotenv.Load(envPath); err != nil {
		t.Fatalf("load .env: %v", err)
	}
	if os.Getenv("MANUAL_SMTP_TEST") != "1" {
		t.Skip("set MANUAL_SMTP_TEST=1 (in .env or the shell) to run this manual email send")
	}

	cfg := SmtpConfig{
		Host:           strings.TrimSpace(os.Getenv("SMTP_HOST")),
		Port:           atoiDef(t, os.Getenv("GOAM_SMTP_PORT"), 587),
		Username:       strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		Password:       os.Getenv("SMTP_PASSWORD"),
		FromEmail:      strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL")),
		FromName:       strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
		UseImplicitTLS: envBool(os.Getenv("SMTP_USE_IMPLICIT_TLS")),
	}
	to := strings.TrimSpace(os.Getenv("TEST_EMAIL_TO"))

	if cfg.Host == "" || cfg.FromEmail == "" || to == "" {
		t.Fatalf("missing required env after loading .env: SMTP_HOST, SMTP_FROM_EMAIL, TEST_EMAIL_TO")
	}

	svc := NewDefaultEmailService(cfg)
	err := svc.SendEmail("manual", "test", GenericHTMLTextEmailParams(
		"GoAM manual SMTP test",
		[]model.EmailAddress{{Email: to}},
		nil,
		nil,
		"This is a manual GoAM SMTP test (plain text).",
		"<p>This is a <strong>manual</strong> GoAM SMTP test (HTML).</p>",
	))
	require.NoError(t, err)
	fmt.Printf("sent manual test email to %s\n", to)
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not locate repo root (go.mod)")
	return ""
}

func atoiDef(t *testing.T, s string, def int) int {
	t.Helper()
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	require.NoError(t, err)
	return n
}

func envBool(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "1" || s == "true" || s == "yes"
}
