package integration

import (
	"context"
	"fmt"
	"goiam/internal"
	"goiam/internal/config"
	"goiam/internal/model"
	"goiam/internal/web"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/fasthttp/router"
	"github.com/gavv/httpexpect/v2"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	"goiam/internal/service"
)

var (
	DefaultTenant = "acme"
	DefaultRealm  = "customers"
)

var Router *router.Router = nil

func SetupIntegrationTest(t *testing.T, flowYaml string) *httpexpect.Expect {

	// Debug print current working directory
	pwd, _ := os.Getwd()
	fmt.Println("Current working directory:", pwd)

	projectRoot := findProjectRoot("README.md")

	config.UnsafeDisableAdminAuthzCheck = true
	config.ConfigPath = filepath.Join(projectRoot, "test/integration/config")
	config.DBConnString = ":memory:?_foreign_keys=on"

	// Call init function
	internal.Initialize()

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
		1000)

	if err != nil {
		t.Fatalf("failed to create access token session: %v", err)
	}

	return token
}
