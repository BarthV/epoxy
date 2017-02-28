// Copyright © 2017 Barthelemy Vessemont
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	log "github.com/Sirupsen/logrus"

	"github.com/BarthV/epoxy/handlers/consulmemcached"
	"github.com/bradfitz/gomemcache/memcache"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/netflix/rend/handlers"
	"github.com/netflix/rend/orcas"
	"github.com/netflix/rend/server"
	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// proxyCmd represents the serve command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Run proxy server",
	Long: `Proxify memcached request to cluster
Request list to consul agent`,
	Run: proxy,
}

func init() {
	RootCmd.AddCommand(proxyCmd)

	proxyCmd.Flags().IntP("port", "p", 11211, "Port listen")
	if err := viper.BindPFlag("port", proxyCmd.Flags().Lookup("port")); err != nil {
		log.WithError(err).Fatal("port")
	}

	proxyCmd.Flags().StringP("timeout", "t", "100ms", "Memcache backend timeout")
	if err := viper.BindPFlag("timeout", proxyCmd.Flags().Lookup("timeout")); err != nil {
		log.WithError(err).Fatal("timeout")
	}
}

func proxy(cmd *cobra.Command, args []string) {
	if viper.GetBool("profile") {
		defer profile.Start().Stop()
	}

	var MemcachedList memcache.ServerList
	var consulOptions = consulApi.QueryOptions{}

	go consulmemcached.ConsulPoller(&MemcachedList, consulOptions)
	mc := memcache.NewFromSelector(&MemcachedList)
	mc.Timeout = viper.GetDuration("timeout")

	server.ListenAndServe(
		server.ListenArgs{
			Type: server.ListenTCP,
			Port: viper.GetInt("port"),
		},
		server.Default,
		orcas.L1Only,
		consulmemcached.New(mc),
		handlers.NilHandler,
	)
}
