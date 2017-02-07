package consulmemcached

import (
	"fmt"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/hashicorp/consul/api"
)

var consulOptions = api.QueryOptions{}

func ConsulPoller(list *memcache.ServerList) {
	consulconf := api.DefaultConfig()
	consulconf.Address = "consul01-par.central.criteo.preprod:8500"
	consul, _ := api.NewClient(consulconf)

	for {
		cluster := []string{}
		res, resqry, err := consul.Catalog().Service("memcached-mesos-test1", "", &consulOptions)
		if err != nil {
			fmt.Println("Consul Catalog query failed")
			continue
		}
		for _, service := range res {
			//fmt.Println(service.Node + ":" + strconv.Itoa(service.ServicePort))
			cluster = append(cluster, service.Node+":"+strconv.Itoa(service.ServicePort))
		}
		fmt.Println(">> Cluster update <<")
		err = list.SetServers(cluster...)
		if err != nil {
			fmt.Println("Memcached client serverlist update failed")
			continue
		}
		fmt.Println(cluster)
		fmt.Println("")
		consulOptions.WaitIndex = resqry.LastIndex
	}
}
