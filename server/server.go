// Package server
package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rromanowicz/mockery/context"
	"github.com/rromanowicz/mockery/db"
	"github.com/rromanowicz/mockery/routing"
)

type Database string

const (
	SqLite   Database = "SqLite"
	Postgres Database = "Postgres"
)

var ctx context.Context

func StartMockServer(port *int, dbType *string) error {
	var repo db.MockRepoInt

	switch Database(*dbType) {
	case SqLite:
		repo = context.SqLite
	case Postgres:
		repo = context.Postgres
	default:
		return fmt.Errorf("unsupported DB Type")
	}

	log.Printf("Starting server [Port: %v, DB: %v]", *port, *dbType)

	ctx = context.InitContext(repo)
	defer ctx.Close()

	handler := &routing.RegexpHandler{}
	routing.RegisterRoutes(ctx, handler)

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", *port),
		Handler: handler,
	}

	log.Println("Service Started")
	return server.ListenAndServe()
}
