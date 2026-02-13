// Package server
package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rromanowicz/mockery/context"
	"github.com/rromanowicz/mockery/model"
	"github.com/rromanowicz/mockery/routing"
	"gopkg.in/yaml.v3"
)

const configFilePath string = "mockery.yml"

var ctx context.Context

func StartMockServer(portOverride *int, dbTypeOverride *string) error {
	config := ReadConfig(portOverride, dbTypeOverride)
	port, handler := SetupServer(&config)

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: handler,
	}

	log.Println("Service Started")
	return server.ListenAndServe()
}

func SetupServer(config *model.Config) (int, *routing.RegexpHandler) {
	var err error
	ctx, err = context.InitContext(config)
	if err != nil {
		panic(err)
	}

	defer ctx.Close()

	handler := &routing.RegexpHandler{}
	routing.RegisterRoutes(ctx, handler)

	return config.Port, handler
}

func ReadConfig(portOverride *int, dbTypeOverride *string) model.Config {
	var config model.Config
	contents, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Println("Failed to read or missing config file. Setting default values.")
	} else {
		err = yaml.Unmarshal(contents, &config)
		if err != nil {
			log.Println("Failed to process config file. Setting default values.")
		}
	}
	err = config.Validate()
	if err != nil {
		panic(err)
	}
	if len(*dbTypeOverride) != 0 {
		config.DBType = model.Database(*dbTypeOverride)
	}
	if *portOverride != 0 {
		config.Port = *portOverride
	}
	return config
}
