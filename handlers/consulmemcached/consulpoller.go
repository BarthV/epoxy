package consulmemcached

import (
	"fmt"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/hashicorp/consul/api"
)

var MemcachedList memcache.ServerList

func ConsulPoller(list *memcache.ServerList) {
	config := api.DefaultConfig()
	config.Address = "consul01-par.central.criteo.preprod:8500"
	consul, _ := api.NewClient(config)

	options := api.QueryOptions{}

	for {
		cluster := []string{}
		res, resqry, _ := consul.Catalog().Service("memcached-mesos-test1", "", &options)
		for _, service := range res {
			//fmt.Println(service.Node + ":" + strconv.Itoa(service.ServicePort))
			cluster = append(cluster, service.Node+":"+strconv.Itoa(service.ServicePort))
		}
		MemcachedCluster = cluster
		fmt.Println(">> Cluster update <<")
		list.SetServers(cluster...)
		fmt.Println(MemcachedCluster)
		fmt.Println("")
		options.WaitIndex = resqry.LastIndex
	}
}
