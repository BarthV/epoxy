package main

import (
	"fmt"
        //"time"
	"strconv"
        "github.com/barthv/epoxy/handlers/consulmemcached"
        "github.com/hashicorp/consul/api"
        "github.com/netflix/rend/handlers"
        "github.com/netflix/rend/orcas"
        "github.com/netflix/rend/server"
        //"github.com/wrighty/ketama"
)

func ConsulPoller() {
	config := api.DefaultConfig()
	config.Address = "consul01-par.central.criteo.preprod:8500"
	consul, _ := api.NewClient(config)

	options := api.QueryOptions{}

	for {
		res, resqry, _ := consul.Catalog().Service("memcached-mesos-test1", "", &options)
		for _, service := range res {
			fmt.Println(service.Node + ":" + strconv.Itoa(service.ServicePort))
		}
                fmt.Println(resqry.LastIndex)
                fmt.Println("")
                options.WaitIndex = resqry.LastIndex
	}
}

func main() {
  //go ConsulPoller()
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
  //time.Sleep(30 * time.Second)
}
