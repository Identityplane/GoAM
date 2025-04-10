package main

import (
	"goiam/internal"
	"goiam/internal/auth/graph"
	"goiam/internal/db/sqlite"
	db "goiam/internal/db/sqlite"
	"goiam/internal/web"
	"log"

	"github.com/valyala/fasthttp"
)

func main() {

	// Init Flows
	internal.InitializeFlows()

	// Init Database
	err := db.Init(db.Config{
		Driver: "sqlite",
		DSN:    "goiam.db?_foreign_keys=on",
	})
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
		return
	}

	// Init Services
	graph.Services.UserRepo = sqlite.NewUserRepository()

	// Init web adapter
	r := web.New()

	log.Println("Server running on http://localhost:8080")
	if err := fasthttp.ListenAndServe(":8080", r.Handler); err != nil {
		log.Fatalf("Error: %s", err)
	}
}
