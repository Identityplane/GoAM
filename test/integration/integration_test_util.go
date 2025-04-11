package integration

import (
	"context"
	"fmt"
	"goiam/internal"
	"goiam/internal/auth/graph"
	"goiam/internal/auth/graph/yaml"
	"goiam/internal/db/sqlite"
	"goiam/internal/web"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	db "goiam/internal/db/sqlite"
)

func SetupIntegrationTest(t *testing.T, flowYaml string) httpexpect.Expect {

	// Overwrite config dir
	if flowYaml == "" {

		// Load standard flow when no flow is given
		internal.FlowsDir = "../../config/flows"
		internal.InitializeFlows()
	} else {

		// Overwrite flow registry with this flow
		flow, err := yaml.LoadFlowFromYAMLString(flowYaml)
		if err != nil {
			log.Fatalf("failed to load flows: %v", err)
		}

		flowWithRoute := &graph.FlowWithRoute{Route: flow.Route, Flow: flow.Flow}

		internal.FlowRegistry[flow.Flow.Name] = flowWithRoute

	}

	// Init Database
	err := db.Init(db.Config{
		Driver: "sqlite",
		DSN:    ":memory:?_foreign_keys=on",
	})

	// Check db
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
		t.Fail()
	}

	// Migrate database
	err = RunTestMigrations()
	if err != nil {
		t.Fatal(err)
	}

	// Create user repo object
	graph.Services.UserRepo = sqlite.NewUserRepository()

	// Setup Http
	handler := web.New().Handler
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
		BaseURL:  "http://example.com", // just a placeholder
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})

	return *e
}

func RunTestMigrations() error {
	sqlBytes, err := os.ReadFile("../../internal/db/sqlite/migrations/001_create_users.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration: %w", err)
	}

	_, err = db.DB.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}
	return nil
}
