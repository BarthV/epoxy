package main

import (
      "fmt"
      "strconv"
//	"github.com/barthv/epoxy/handlers/consulmemcached"
      "github.com/hashicorp/consul/api"
//	"github.com/netflix/rend/handlers"
//	"github.com/netflix/rend/orcas"
//	"github.com/netflix/rend/server"
//      "https://github.com/wrighty/ketama"
)

func main() {
/*	server.ListenAndServe(
		server.ListenArgs{
			Type: server.ListenTCP,
			Port: 11211,
		},
		server.Default,
		orcas.L1Only,
		consulmemcached.New,
                handlers.NilHandler,
	) */
      config := api.DefaultConfig()
      config.Address = "consul01-par.central.criteo.preprod:8500"
      consul, _ := api.NewClient(config)

      r, q, _ := consul.Catalog().Service("memcached-mesos-test1", "", nil)
      for _, v := range r {
            fmt.Println(v.Node + ":" + strconv.Itoa(v.ServicePort))
      }
      fmt.Println(q.LastIndex)

      fmt.Println(" -- ")

      options := api.QueryOptions{
            Datacenter: "par",
            WaitIndex: q.LastIndex,
      }
      r2, q2, _ := consul.Catalog().Service("memcached-mesos-test1", "", &options)
      for _, v := range r2 {
            fmt.Println(v.Node + ":" + strconv.Itoa(v.ServicePort))
      }
      fmt.Println(q2.LastIndex)
}
