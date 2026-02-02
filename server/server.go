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

func StartMockServer(dbType Database) error {
	log.Println("Starting Service")

	var repo db.MockRepoInt
	switch dbType {
	case SqLite:
		repo = context.SqLite
	case Postgres:
		repo = context.Postgres
	default:
		return fmt.Errorf("unsupported DB Type")
	}

	ctx = context.InitContext(repo)
	defer ctx.Close()

	handler := &routing.RegexpHandler{}
	routing.RegisterRoutes(ctx, handler)

	server := http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println("Service Started")
	return server.ListenAndServe()
}
