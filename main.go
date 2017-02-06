package main

import (
	"github.com/barthv/epoxy/handlers/consulmemcached"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/netflix/rend/handlers"
	"github.com/netflix/rend/orcas"
	"github.com/netflix/rend/server"
)

func main() {
	//go ConsulPoller()
	var MemcachedList memcache.ServerList
	go consulmemcached.ConsulPoller(&MemcachedList)
	mc := memcache.NewFromSelector(&MemcachedList)

	server.ListenAndServe(
		server.ListenArgs{
			Type: server.ListenTCP,
			Port: 11211,
		},
		server.Default,
		orcas.L1Only,
		consulmemcached.New(mc),
		handlers.NilHandler,
	)
}
