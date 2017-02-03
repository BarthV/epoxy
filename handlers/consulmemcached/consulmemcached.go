package consulmemcached

import (
	"sync"
	"time"

	"github.com/netflix/rend/common"
	"github.com/netflix/rend/handlers"
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
	mutex *sync.RWMutex
}

var singleton = &Handler{
	data:  make(map[string]entry),
	mutex: new(sync.RWMutex),
}

func New() (handlers.Handler, error) {
	// return the same singleton map each time so all connections see the same data
	return singleton, nil
}

func (h *Handler) Set(cmd common.SetRequest) error {
	h.mutex.Lock()

	var exptime uint32
	if cmd.Exptime > 0 {
		exptime = uint32(time.Now().Unix()) + cmd.Exptime
	}

	h.data[string(cmd.Key)] = entry{
		data:    cmd.Data,
		exptime: exptime,
		flags:   cmd.Flags,
	}

	h.mutex.Unlock()
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

	h.mutex.RLock()

	for idx, bk := range cmd.Keys {
		e, ok := h.data[string(bk)]

		if !ok || e.isExpired() {
			delete(h.data, string(bk))
			dataOut <- common.GetResponse{
				Miss:   true,
				Quiet:  cmd.Quiet[idx],
				Opaque: cmd.Opaques[idx],
				Key:    bk,
			}
			continue
		}

		dataOut <- common.GetResponse{
			Miss:   false,
			Quiet:  cmd.Quiet[idx],
			Opaque: cmd.Opaques[idx],
			Flags:  e.flags,
			Key:    bk,
			Data:   e.data,
		}
	}

	h.mutex.RUnlock()

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
