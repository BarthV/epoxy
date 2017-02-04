package consulmemcached

import (
	"time"
	"fmt"
        "strconv"

        "github.com/hashicorp/consul/api"
	"github.com/netflix/rend/common"
	"github.com/netflix/rend/handlers"
        "github.com/wrighty/ketama"
)

type entry struct {
   exptime uint32
   flags   uint32
   data    []byte
}

func (e entry) isExpired() bool {
	return e.exptime != 0 && e.exptime < uint32(time.Now().Unix())
}


type Handler struct {
  data  map[string]entry
}

var singleton = &Handler{
	data:  make(map[string]entry),
}

var MemcachedCluster []string


func ConsulPoller() {
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
                fmt.Println(MemcachedCluster)
                fmt.Println("")
                options.WaitIndex = resqry.LastIndex
        }
}

func New() (handlers.Handler, error) {
        go ConsulPoller()
	// return the same singleton map each time so all connections see the same data
	return singleton, nil
}

func (h *Handler) Set(cmd common.SetRequest) error {
        fmt.Println(">> Set operation <<")
        fmt.Println("key = " + string(cmd.Key))
        fmt.Println("data = " + string(cmd.Data))
        fmt.Printf("ttl = "+ strconv.FormatInt(int64(cmd.Exptime),10) + "\n")
        fmt.Println("  -----  ")

        c := ketama.Make(MemcachedCluster)
        host := c.GetHost(string(cmd.Key))
        fmt.Println("CONSISTENT HASHING'S HOST = " + host)
        fmt.Println("")
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
