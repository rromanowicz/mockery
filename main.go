package main

import (
	"log"
	"net/http"

	"github.com/rromanowicz/mockery/context"
	"github.com/rromanowicz/mockery/routing"
)

var ctx context.Context

func main() {
	log.Println("Starting Service")
	ctx = context.InitContext()
	defer ctx.Close()

	handler := &routing.RegexpHandler{}
	routing.RegisterRoutes(ctx, handler)

	server := http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println("Service Started")
	err := server.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
