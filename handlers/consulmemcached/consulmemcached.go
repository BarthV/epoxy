package consulmemcached

import (
	"fmt"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/netflix/rend/common"
	"github.com/netflix/rend/handlers"
)

type Handler struct {
	mc *memcache.Client
}

func New(mclient *memcache.Client) handlers.HandlerConst {
	return func() (handlers.Handler, error) {
		fmt.Println(">> Starting Proxy ! <<")
		handler := &Handler{
			mc: mclient,
		}
		return handler, nil
	}
}

func (h *Handler) Set(cmd common.SetRequest) error {
	fmt.Println(">> Set operation <<")
	fmt.Println("key = " + string(cmd.Key))
	fmt.Println("data = " + string(cmd.Data))
	fmt.Printf("ttl = " + strconv.FormatInt(int64(cmd.Exptime), 10) + "\n")
	fmt.Println("  -----  ")
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
	dataOut := make(chan common.GetResponse, len(cmd.Keys))
	errorOut := make(chan error)

	fmt.Println(">> Get operation <<")

	for idx, bk := range cmd.Keys {
		item, err := h.mc.Get(string(bk))

		if err != nil {
			//if err == common.ErrKeyNotFound {
			if err.Error() == "memcache: cache miss" {
				fmt.Println("Key not found !")
				dataOut <- common.GetResponse{
					Miss:   true,
					Quiet:  cmd.Quiet[idx],
					Opaque: cmd.Opaques[idx],
					Key:    bk,
				}
				continue
			}
			fmt.Println("Unkown error !")
			fmt.Println(err)
			errorOut <- err
			continue
		}

		fmt.Println("key = " + item.Key)
		fmt.Println("data = " + string(item.Value))
		fmt.Printf("ttl = " + strconv.FormatInt(int64(item.Expiration), 10) + "\n")
		fmt.Println("  -----  ")
		fmt.Println("")
		dataOut <- common.GetResponse{
			Miss:   false,
			Quiet:  cmd.Quiet[idx],
			Opaque: cmd.Opaques[idx],
			Flags:  item.Flags,
			Key:    bk,
			Data:   item.Value,
		}
	}
	close(dataOut)
	close(errorOut)
	return dataOut, errorOut
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
