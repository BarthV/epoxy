package consulmemcached

import (
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

var consulOptions = api.QueryOptions{}

func ConsulPoller(list *memcache.ServerList) {
	consulconf := api.DefaultConfig()
	consulconf.Address = viper.GetString("consul.address")
	consul, _ := api.NewClient(consulconf)

	for {
		cluster := []string{}
		res, resqry, err := consul.Catalog().Service(viper.GetString("consul.service"), "", &consulOptions)
		if err != nil {
			log.WithError(err).Error("Consul Catalog query failed")
			continue
		}
		for _, service := range res {
			//fmt.Println(service.Node + ":" + strconv.Itoa(service.ServicePort))
			cluster = append(cluster, service.Node+":"+strconv.Itoa(service.ServicePort))
		}
		err = list.SetServers(cluster...)
		if err != nil {
			log.WithError(err).Error("Memcached client serverlist update failed")
			continue
		}
		log.WithFields(log.Fields{
			"cluster": cluster,
		}).Info("Cluster update")

		consulOptions.WaitIndex = resqry.LastIndex
	}
}
