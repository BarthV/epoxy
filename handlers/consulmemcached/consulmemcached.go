package consulmemcached

import (
	"fmt"
        "strconv"

        "github.com/hashicorp/consul/api"
	"github.com/netflix/rend/common"
	"github.com/netflix/rend/handlers"
        "github.com/wrighty/ketama"
        "github.com/bradfitz/gomemcache/memcache"
)

type Handler struct {
  mc *memcache.Client
}

var singleton = &Handler{
  mc: memcache.NewFromSelector(&MemcachedList),
}

var MemcachedList memcache.ServerList
var MemcachedCluster []string

func (h *Handler) ConsulPoller() {
        config := api.DefaultConfig()
        config.Address = "consul01-par.central.criteo.preprod:8500"
        consul, _ := api.NewClient(config)

        options := api.QueryOptions{}

        for {
                cluster := []string{}
                res, resqry, _ := consul.Catalog().Service("memcached-mesos-test1", "", &options)
                for _, service := range res {
                        //fmt.Println(service.Node + ":" + strconv.Itoa(service.ServicePort))
                        cluster = append(cluster, service.Node + ":" + strconv.Itoa(service.ServicePort))
                }
                MemcachedCluster = cluster
                fmt.Println(">> Cluster update <<")
                MemcachedList.SetServers(cluster...)
                fmt.Println(MemcachedCluster)
                fmt.Println("")
                options.WaitIndex = resqry.LastIndex
        }
}

func New() (handlers.Handler, error) {
        fmt.Println(">> Starting Proxy ! <<")
        go singleton.ConsulPoller()
	return singleton, nil
}

func (h *Handler) Set(cmd common.SetRequest) error {
        fmt.Println(">> Set operation <<")
        fmt.Println("key = " + string(cmd.Key))
        fmt.Println("data = " + string(cmd.Data))
        fmt.Printf("ttl = "+ strconv.FormatInt(int64(cmd.Exptime),10) + "\n")
        fmt.Println("  -----  ")

        c := ketama.Make(MemcachedCluster)
        hostketama := c.GetHost(string(cmd.Key))
        fmt.Println("KETAMA HOST = " + hostketama)
        hostclient, _ := MemcachedList.PickServer(string(cmd.Key))
        fmt.Println("GO LIB HOST = " + hostclient.String())
        fmt.Println("")
        h.mc.Set(&memcache.Item{Key: string(cmd.Key), Value: cmd.Data})
	return nil
}

func (h *Handler) Add(cmd common.SetRequest) error {
	return nil
}

func (h *Handler) Replace(cmd common.SetRequest) error {
	return nil
}

func (h *Handler) Append(cmd common.SetRequest) error {
	return nil
}

func (h *Handler) Prepend(cmd common.SetRequest) error {
	return nil
}

func (h *Handler) Get(cmd common.GetRequest) (<-chan common.GetResponse, <-chan error) {
	return nil, nil
}

func (h *Handler) GetE(cmd common.GetRequest) (<-chan common.GetEResponse, <-chan error) {
	return nil, nil
}

func (h *Handler) GAT(cmd common.GATRequest) (common.GetResponse, error) {
	return common.GetResponse{}, nil
}

func (h *Handler) Delete(cmd common.DeleteRequest) error {
	return nil
}

func (h *Handler) Touch(cmd common.TouchRequest) error {
	return nil
}

func (h *Handler) Close() error {
	return nil
}
