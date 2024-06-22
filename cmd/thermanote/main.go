package main

import (
	"github.com/nulltrope/thermanote/internal/pkg/server"
)

func main() {
	srv, err := server.NewServer("127.0.0.1:8081")
	if err != nil {
		panic(err)
	}
	err = srv.Start()
	if err != nil {
		panic(err)
	}
}
