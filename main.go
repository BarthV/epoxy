package main

import (

	//"time"

	"github.com/barthv/epoxy/handlers/consulmemcached"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/netflix/rend/handlers"
	"github.com/netflix/rend/orcas"
	"github.com/netflix/rend/server"
	//"github.com/wrighty/ketama"
)

func main() {
	//go ConsulPoller()
	var memcacheList memcache.ServerList
	go consulmemcached.ConsulPoller(&memcacheList)
	mc := memcache.NewFromSelector(&memcacheList)

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
	//time.Sleep(30 * time.Second)
}
