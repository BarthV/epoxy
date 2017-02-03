package main

import (
	"github.com/netflix/rend/handlers"
	"github.com/barthv/epoxy/handlers/consulmemcached"
	"github.com/netflix/rend/orcas"
	"github.com/netflix/rend/server"
)

func main() {
	server.ListenAndServe(
		server.ListenArgs{
			Type: server.ListenTCP,
			Port: 11211,
		},
		server.Default,
		orcas.L1Only,
		consulmemcached.New,
                handlers.NilHandler,
	)
}
