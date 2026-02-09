// Package server
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rromanowicz/mockery/context"
	"github.com/rromanowicz/mockery/db"
	"github.com/rromanowicz/mockery/model"
	"github.com/rromanowicz/mockery/routing"
)

var ctx context.Context

func StartMockServer(portOverride *int, dbTypeOverride *string) error {
	config, err := ReadConfig(portOverride, dbTypeOverride)
	if err != nil {
		panic(err)
	}
	var repo db.MockRepoInt
	var dbParams model.DBParams
	port := portOverride
	dbType := model.Database(*dbTypeOverride)

	if len(*dbTypeOverride) == 0 {
		dbType = config.DBType
	}
	if *portOverride == 0 {
		port = &config.Port
	}

	switch dbType {
	case model.SqLite:
		repo = context.SqLite
		dbParams = config.DBConfig.SqLite
	case model.SqLiteORM:
		repo = context.SqLiteORM
		dbParams = config.DBConfig.SqLite
	case model.Postgres:
		repo = context.Postgres
		dbParams = config.DBConfig.Postgres
	default:
		if len(dbType) == 0 {
			return fmt.Errorf("db type not set")
		}
		return fmt.Errorf("unsupported DB Type")
	}

	log.Printf("Starting server [Port: %v, DB: %s]", *port, dbType)

	ctx = context.InitContext(dbParams, repo)
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

func ReadConfig(port *int, dbType *string) (model.Config, error) {
	var config model.Config
	contents, err := os.ReadFile("./.config")
	if err != nil {
		log.Println("Failed to read or missing config file. Setting default values.")
	} else {
		err = json.Unmarshal(contents, &config)
		if err != nil {
			log.Println("Failed to process config file. Setting default values.")
		}
	}
	err = config.Validate(port, dbType)
	return config, err
}
