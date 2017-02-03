// Copyright 2016 Netflix, Inc.
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

package orcas

import (
	"hash"
	"hash/fnv"
	"sync"
	"sync/atomic"

	"github.com/netflix/rend/common"
	"github.com/netflix/rend/handlers"
)

const maxLockSets = 1024

var (
	locks   = make([][]sync.Locker, maxLockSets)
	rlocks  = make([][]sync.Locker, maxLockSets)
	curslot uint32
)

func getNewLocks(multipleReaders bool, concurrency uint8) (slot uint32) {
	slot = atomic.AddUint32(&curslot, 1)

	if slot > maxLockSets {
		panic("Too many lock sets!")
	}

	locks[slot] = make([]sync.Locker, 1<<concurrency)
	rlocks[slot] = make([]sync.Locker, 1<<concurrency)

	if multipleReaders {
		for idx := range locks[slot] {
			temp := &sync.RWMutex{}
			locks[slot][idx] = temp
			rlocks[slot][idx] = temp.RLocker()
		}
	} else {
		for idx := range locks[slot] {
			temp := &sync.Mutex{}
			locks[slot][idx] = temp
			rlocks[slot][idx] = temp
		}
	}

	return
}

type LockedOrca struct {
	wrapped Orca
	locks   []sync.Locker
	rlocks  []sync.Locker
	hpool   *sync.Pool
	//counts  []uint32
}

// Locking wraps an orcas.Orca to provide locking around operations on the same
// key. When multipleReaders is true, operations will allow many readers and
// only a single writer at a time. When false, only a single reader is allowed.
// The concurrency param allows 2^(concurrency) operations to happen in
// parallel. E.g. concurrency of 1 would allow 2 parallel operations, while a
// concurrency of 4 allows 2^4 = 16 parallel operations.
func Locked(oc OrcaConst, multipleReaders bool, concurrency uint8) (OrcaConst, uint32) {
	if concurrency < 0 {
		panic("Concurrency level must be at least 0")
	}

	slot := getNewLocks(multipleReaders, concurrency)

	//counts := make([]uint32, 1<<concurrency)

	hashpool := &sync.Pool{
		New: func() interface{} {
			return fnv.New32a()
		},
	}

	return func(l1, l2 handlers.Handler, res common.Responder) Orca {
		return &LockedOrca{
			wrapped: oc(l1, l2, res),
			locks:   locks[slot],
			rlocks:  rlocks[slot],
			hpool:   hashpool,
			//counts:  counts,
		}
	}, slot
}

func LockedWithExisting(oc OrcaConst, locksetID uint32) OrcaConst {
	if cur := atomic.LoadUint32(&curslot); cur < locksetID {
		panic("Asked for lock set that does not exist!")
	}

	hashpool := &sync.Pool{
		New: func() interface{} {
			return fnv.New32a()
		},
	}

	return func(l1, l2 handlers.Handler, res common.Responder) Orca {
		return &LockedOrca{
			wrapped: oc(l1, l2, res),
			locks:   locks[locksetID],
			rlocks:  rlocks[locksetID],
			hpool:   hashpool,
		}
	}
}

//var numops uint64 = 0

func (l *LockedOrca) getlock(key []byte, read bool) sync.Locker {
	h := l.hpool.Get().(hash.Hash32)
	h.Reset()

	// Calculate bucket using hash and mod. hash.Hash.Write() never returns an error.
	h.Write(key)
	bucket := int(h.Sum32())
	bucket &= len(l.locks) - 1

	//atomic.AddUint32(&l.counts[bucket], 1)

	//if (atomic.AddUint64(&numops, 1) % 10000) == 0 {
	//	for idx, count := range l.counts {
	//		fmt.Printf("%d: %d\n", idx, count)
	//	}
	//}

	if read {
		return l.rlocks[bucket]
	}

	return l.locks[bucket]
}

func (l *LockedOrca) Set(req common.SetRequest) error {
	lock := l.getlock(req.Key, false)
	lock.Lock()
	defer lock.Unlock()
	ret := l.wrapped.Set(req)
	return ret
}

func (l *LockedOrca) Add(req common.SetRequest) error {
	lock := l.getlock(req.Key, false)
	lock.Lock()
	defer lock.Unlock()
	ret := l.wrapped.Add(req)
	return ret
}

func (l *LockedOrca) Replace(req common.SetRequest) error {
	lock := l.getlock(req.Key, false)
	lock.Lock()
	defer lock.Unlock()
	ret := l.wrapped.Replace(req)
	return ret
}

func (l *LockedOrca) Append(req common.SetRequest) error {
	lock := l.getlock(req.Key, false)
	lock.Lock()
	defer lock.Unlock()
	ret := l.wrapped.Append(req)
	return ret
}

func (l *LockedOrca) Prepend(req common.SetRequest) error {
	lock := l.getlock(req.Key, false)
	lock.Lock()
	defer lock.Unlock()
	ret := l.wrapped.Prepend(req)
	return ret
}

func (l *LockedOrca) Delete(req common.DeleteRequest) error {
	lock := l.getlock(req.Key, false)
	lock.Lock()
	defer lock.Unlock()
	ret := l.wrapped.Delete(req)
	return ret
}

func (l *LockedOrca) Touch(req common.TouchRequest) error {
	lock := l.getlock(req.Key, false)
	lock.Lock()
	defer lock.Unlock()
	ret := l.wrapped.Touch(req)
	return ret
}

func (l *LockedOrca) Get(req common.GetRequest) error {
	// Lock for each read key, complete the read, and then move on.
	// The last key sent through should have a noop at the end to complete the
	// whole interaction between the client and this server.
	var ret error
	var lock sync.Locker

	// guarantee that an operation that failed with a panic will unlock its lock
	defer func() {
		if r := recover(); r != nil {
			if lock != nil {
				lock.Unlock()
			}
		}
	}()

	for idx, key := range req.Keys {
		// Acquire read lock (true == read)
		lock = l.getlock(key, true)
		lock.Lock()

		// The last request will have these set to complete the interaction
		noopOpaque := uint32(0)
		noopEnd := false
		if idx == len(req.Keys)-1 {
			noopOpaque = req.NoopOpaque
			noopEnd = req.NoopEnd
		}

		subreq := common.GetRequest{
			Keys:       [][]byte{key},
			Opaques:    []uint32{req.Opaques[idx]},
			Quiet:      []bool{req.Quiet[idx]},
			NoopOpaque: noopOpaque,
			NoopEnd:    noopEnd,
		}

		// Make the actual request
		ret = l.wrapped.Get(subreq)

		// release read lock
		lock.Unlock()

		// Bail out early if there was an error (misses are not errors in this sense)
		// This will probably end up breaking the connection anyway, so no worries
		// about leaving the gets half-done.
		if ret != nil {
			break
		}
	}

	return ret
}

func (l *LockedOrca) GetE(req common.GetRequest) error {
	// Lock for each read key, complete the read, and then move on.
	// The last key sent through should have a noop at the end to complete the
	// whole interaction between the client and this server.
	var ret error
	var lock sync.Locker

	// guarantee that an operation that failed with a panic will unlock its lock
	defer func() {
		if r := recover(); r != nil {
			if lock != nil {
				lock.Unlock()
			}

			panic(r)
		}
	}()

	for idx, key := range req.Keys {
		// Acquire read lock (true == read)
		lock = l.getlock(key, true)
		lock.Lock()

		// The last request will have these set to complete the interaction
		noopOpaque := uint32(0)
		noopEnd := false
		if idx == len(req.Keys)-1 {
			noopOpaque = req.NoopOpaque
			noopEnd = req.NoopEnd
		}

		subreq := common.GetRequest{
			Keys:       [][]byte{key},
			Opaques:    []uint32{req.Opaques[idx]},
			Quiet:      []bool{req.Quiet[idx]},
			NoopOpaque: noopOpaque,
			NoopEnd:    noopEnd,
		}

		// Make the actual request
		ret = l.wrapped.GetE(subreq)

		// release read lock
		lock.Unlock()

		// Bail out early if there was an error (misses are not errors in this sense)
		// This will probably end up breaking the connection anyway, so no worries
		// about leaving the gets half-done.
		if ret != nil {
			break
		}
	}

	return ret
}

func (l *LockedOrca) Gat(req common.GATRequest) error {
	lock := l.getlock(req.Key, false)
	lock.Lock()
	defer lock.Unlock()
	ret := l.wrapped.Gat(req)
	return ret
}

func (l *LockedOrca) Noop(req common.NoopRequest) error {
	return l.wrapped.Noop(req)
}

func (l *LockedOrca) Quit(req common.QuitRequest) error {
	return l.wrapped.Quit(req)
}

func (l *LockedOrca) Version(req common.VersionRequest) error {
	return l.wrapped.Version(req)
}

func (l *LockedOrca) Unknown(req common.Request) error {
	return l.wrapped.Unknown(req)
}

func (l *LockedOrca) Error(req common.Request, reqType common.RequestType, err error) {
	l.wrapped.Error(req, reqType, err)
}
