package main

import (
	"log"

	"github.com/rromanowicz/mockery/server"
)

func main() {
	err := server.StartMockServer(server.SqLite)
	if err != nil {
		log.Panic(err)
	}
}
