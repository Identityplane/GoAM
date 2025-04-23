package integration

import (
	"context"
	"fmt"
	"goiam/internal"
	"goiam/internal/config"
	"goiam/internal/web"
	"net"
	"net/http"
	"os"
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
	ConfigPath    = "../../config"
)

var Router *router.Router = nil

func SetupIntegrationTest(t *testing.T, flowYaml string) *httpexpect.Expect {

	// Debug print current working directory
	pwd, _ := os.Getwd()
	fmt.Println("Current working directory:", pwd)

	config.ConfigPath = ConfigPath
	config.DBConnString = ":memory:?_foreign_keys=on"

	// Call init function
	internal.Initialize()

	// if present manually add the flow to the realm
	if flowYaml != "" {
		flow, err := service.LoadFlowFromYAMLString(flowYaml)

		if err != nil {
			t.Fatalf("failed to load flow from YAML: %v", err)
		}

		// Add the flow to the realm
		realm, _ := service.GetServices().RealmService.GetRealm(DefaultTenant + "/" + DefaultRealm)
		if realm == nil {
			t.Fatalf("failed to get realm %s/%s", DefaultTenant, DefaultRealm)
		}
		realm.Config.Flows = append(realm.Config.Flows, *flow)
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
