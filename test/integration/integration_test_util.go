package integration

import (
	"context"
	"database/sql"
	"fmt"
	"goiam/internal/auth/graph"
	"goiam/internal/logger"
	"goiam/internal/realms"
	"goiam/internal/web"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/fasthttp/router"
	"github.com/gavv/httpexpect/v2"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	"goiam/internal/db/sqlite_adapter"    // lint:ignore ST1019 (This should be fixed, but is not a priority)
	db "goiam/internal/db/sqlite_adapter" // lint:ignore
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

	// Setup Realm
	realms.InitRealms(ConfigPath) // #nosec

	// if present manually add the flow to the realm
	if flowYaml != "" {
		flow, err := realms.LoadFlowFromYAMLString(flowYaml)

		if err != nil {
			t.Fatalf("failed to load flow from YAML: %v", err)
		}

		// Add the flow to the realm
		realm, _ := realms.GetRealm(DefaultTenant + "/" + DefaultRealm)
		if realm == nil {
			t.Fatalf("failed to get realm %s/%s", DefaultTenant, DefaultRealm)
		}
		realm.Config.Flows = append(realm.Config.Flows, *flow)
	}

	// Overwrite Template Dirs
	//web.LayoutTemplatePath = "../../internal/web/templates/layout.html"
	//web.NodeTemplatesPath = "../../internal/web/templates/nodes"

	// Init Database
	database, err := db.Init(db.Config{
		Driver: "sqlite",
		DSN:    ":memory:?_foreign_keys=on",
	})

	// Check db
	if err != nil {
		logger.PanicNoContext("DB init failed: %v", err)
		t.Fail()
	}

	// Migrate database
	err = RunTestMigrations(database)
	if err != nil {
		t.Fatal(err)
	}

	// Create user repo object
	userDb, err := sqlite_adapter.NewSQLiteUserDB(database)
	if err != nil {
		t.Fatalf("failed to create sqlite user db: %v", err)
	}
	userRepo := sqlite_adapter.NewUserRepository(DefaultTenant, DefaultRealm, userDb)

	// get the loaded realm and init the service registry
	realm, _ := realms.GetRealm(DefaultTenant + "/" + DefaultRealm)
	realm.Services = &graph.ServiceRegistry{
		UserRepo: userRepo,
	}

	// Setup UserAdminService
	UserAdminService := service.NewUserService(userDb)

	// Setup Http
	Router = web.New(UserAdminService)
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
		BaseURL:  "http://example.com", // just a placeholder
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})

	return e
}

func RunTestMigrations(database *sql.DB) error {
	sqlBytes, err := os.ReadFile("../../internal/db/sqlite_adapter/migrations/001_create_users.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration: %w", err)
	}

	_, err = database.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}
	return nil
}
