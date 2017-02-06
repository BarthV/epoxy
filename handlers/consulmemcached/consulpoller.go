package consulmemcached

import (
	"fmt"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/hashicorp/consul/api"
)

func ConsulPoller(list *memcache.ServerList) {
	consulconf := api.DefaultConfig()
	consulconf.Address = "consul01-par.central.criteo.preprod:8500"
	consul, _ := api.NewClient(consulconf)

	options := api.QueryOptions{}

	for {
		cluster := []string{}
		res, resqry, _ := consul.Catalog().Service("memcached-mesos-test1", "", &options)
		for _, service := range res {
			//fmt.Println(service.Node + ":" + strconv.Itoa(service.ServicePort))
			cluster = append(cluster, service.Node+":"+strconv.Itoa(service.ServicePort))
		}
		fmt.Println(">> Cluster update <<")
		list.SetServers(cluster...)
		fmt.Println(cluster)
		fmt.Println("")
		options.WaitIndex = resqry.LastIndex
	}
}
