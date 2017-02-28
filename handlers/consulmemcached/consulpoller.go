package consulmemcached

import (
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

func ConsulPoller(list *memcache.ServerList, consulOptions api.QueryOptions) {
	consulconf := api.DefaultConfig()
	consulconf.Address = viper.GetString("consul.address")
	consul, _ := api.NewClient(consulconf)

	for {
		cluster := []string{}
		res, resqry, err := consul.Health().Service(viper.GetString("consul.service"), "", true, &consulOptions)
		if err != nil {
			log.WithError(err).Error("Consul Services query failed")
			continue
		}
		for _, service := range res {
			ip := service.Service.Address
			if ip == "" {
				ip = service.Node.Address
			}
			cluster = append(cluster, ip+":"+strconv.Itoa(service.Service.Port))
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
