package integration

import (
	"context"
	"goiam/internal/web"
	"log"
	"net"
	"net/http"
	"testing"

	db "goiam/internal/db/sqlite"

	"github.com/gavv/httpexpect/v2"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func SetupIntegrationTest(t *testing.T) httpexpect.Expect {

	// Init Database
	err := db.Init(db.Config{
		Driver: "sqlite",
		DSN:    "goiam.db?_foreign_keys=on",
	})
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
		t.Fail()
	}

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
	})

	return *e
}
