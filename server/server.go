// Package server
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rromanowicz/mockery/context"
	"github.com/rromanowicz/mockery/model"
	"github.com/rromanowicz/mockery/routing"
)

var ctx context.Context

func StartMockServer(portOverride *int, dbTypeOverride *string) error {
	config, err := ReadConfig(portOverride, dbTypeOverride)
	if err != nil {
		panic(err)
	}

	if len(*dbTypeOverride) != 0 {
		config.DBType = model.Database(*dbTypeOverride)
	}
	if *portOverride != 0 {
		config.Port = *portOverride
	}

	ctx, err = context.InitContext(config)
	if err != nil {
		panic(err)
	}

	defer ctx.Close()

	handler := &routing.RegexpHandler{}
	routing.RegisterRoutes(ctx, handler)

	print()
	server := http.Server{
		Addr:    fmt.Sprintf(":%v", config.Port),
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
