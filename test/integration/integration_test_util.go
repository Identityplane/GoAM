package integration

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Identityplane/GoAM/internal"
	"github.com/Identityplane/GoAM/internal/logger"
	"github.com/Identityplane/GoAM/internal/web"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/Identityplane/GoAM/pkg/server_settings"
	"github.com/spf13/viper"

	"github.com/PuerkitoBio/goquery"
	"github.com/fasthttp/router"
	"github.com/gavv/httpexpect/v2"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	"github.com/Identityplane/GoAM/internal/service"
	dbinit "github.com/Identityplane/GoAM/pkg/db/init"
	services_init "github.com/Identityplane/GoAM/pkg/services/init"
)

var (
	DefaultTenant = "acme"
	DefaultRealm  = "customers"
)

var Router *router.Router = nil

func SetupIntegrationTest(t *testing.T, flowYaml string) *httpexpect.Expect {

	log := logger.GetGoamLogger()

	// Debug print current working directory
	pwd, _ := os.Getwd()
	fmt.Println("Current working directory:", pwd)

	// Set config file to default goam.yaml file
	projectRoot := findProjectRoot("README.md")
	configFile := filepath.Join(projectRoot, "goam.yaml")
	viper.SetConfigFile(configFile)

	// Load server settings
	serverSettings, err := server_settings.InitWithViper()
	if err != nil {
		t.Fatalf("failed to initialize server settings: %v", err)
	}

	log.Debug().Msg("Initializing in-memory database")

	// Reset singleton factories to ensure each test gets a fresh database
	// This is critical because the factories are singletons and would otherwise
	// reuse the same database connection across tests
	dbinit.SetDBConnectionsFactory(nil)
	services_init.SetServicesFactory(nil)

	// Overwrite settings for testing
	serverSettings.UnsafeDisableAdminAuth = true
	serverSettings.RealmConfigurationFolder = filepath.Join(projectRoot, "test/integration/config")
	serverSettings.DBConnString = ":memory:?_foreign_keys=on"

	// Disable the default listeners
	serverSettings.ListenerHttp = ""
	serverSettings.ListenerHTTPS = ""

	// Call init function
	internal.Initialize(serverSettings)

	// if present manually add the flow to the realm
	if flowYaml != "" {

		flow := &model.Flow{
			Id:             "test_flow",
			Route:          "test_flow",
			Active:         true,
			DefinitionYaml: flowYaml,
		}

		// Add the flow to the loaded flows
		err := service.GetServices().FlowService.CreateFlow(DefaultTenant, DefaultRealm, *flow)
		if err != nil {
			t.Fatalf("failed to create flow: %v", err)
		}
	}

	// Setup Http
	Router = web.New()
	handler := Router.Handler
	ln := fasthttputil.NewInmemoryListener()

	// Serve fasthttp using the in-memory listener
	go func() {
		_ = fasthttp.Serve(ln, handler)
	}()

	// Convert the in-memory listener to a net/http-compatible client
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return ln.Dial()
			},
		},
		// Configure client to not follow redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	e := httpexpect.WithConfig(httpexpect.Config{
		Client:   client,
		BaseURL:  "http://integration-test.com", // just a placeholder
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})

	return e
}

func findProjectRoot(markerFile string) string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, markerFile)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("Project root not found")
		}
		dir = parent
	}
}

func CreateAccessTokenSession(t *testing.T, user model.User) string {

	token, _, err := service.GetServices().SessionsService.CreateAccessTokenSession(
		context.Background(),
		user.Tenant, user.Realm,
		"clientid",
		user.ID,
		[]string{},
		"test",
		1000,
		nil)

	if err != nil {
		t.Fatalf("failed to create access token session: %v", err)
	}

	return token
}

// ExtractStepFromHTML extracts the step value from the hidden input field in the HTML response
func ExtractStepFromHTML(t *testing.T, htmlContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	stepInput := doc.Find("input[type='hidden'][name='step']")
	if stepInput.Length() == 0 {
		t.Fatal("Expected to find hidden step field in HTML response")
	}

	stepValue, exists := stepInput.Attr("value")
	if !exists {
		t.Fatal("Step input field found but has no value attribute")
	}

	return stepValue
}

// SubmitAuthForm submits a form with the given field values, automatically extracting the step from the response
func SubmitAuthForm(t *testing.T, e *httpexpect.Expect, url string, sessionCookie string, fields map[string]string) *httpexpect.Request {
	// First, get the current form to extract the step value
	getResp := e.GET(url).
		WithCookie("session_id", sessionCookie).
		Expect().
		Status(http.StatusOK).
		Body()

	htmlContent := getResp.Raw()
	step := ExtractStepFromHTML(t, htmlContent)

	// Build form fields with step
	formFields := map[string]string{"step": step}
	for k, v := range fields {
		formFields[k] = v
	}

	// Submit the form
	req := e.POST(url).
		WithHeader("Content-Type", "application/x-www-form-urlencoded").
		WithCookie("session_id", sessionCookie)

	for k, v := range formFields {
		req = req.WithFormField(k, v)
	}

	return req
}
