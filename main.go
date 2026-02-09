package main

import (
	"flag"
	"log"

	"github.com/rromanowicz/mockery/server"
)

func main() {
	serverPort := flag.Int("port", 0, "Server port.")
	serverDB := flag.String("db", "", "Database type.")
	flag.Parse()

	err := server.StartMockServer(serverPort, serverDB)
	if err != nil {
		log.Panic(err)
	}
}
