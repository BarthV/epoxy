// Copyright 2015 Netflix, Inc.
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
	"log"

	"github.com/netflix/rend/common"
	"github.com/netflix/rend/handlers"
	"github.com/netflix/rend/metrics"
	"github.com/netflix/rend/timer"
)

type L1L2BatchOrca struct {
	l1  handlers.Handler
	l2  handlers.Handler
	res common.Responder
}

func L1L2Batch(l1, l2 handlers.Handler, res common.Responder) Orca {
	return &L1L2BatchOrca{
		l1:  l1,
		l2:  l2,
		res: res,
	}
}

func (l *L1L2BatchOrca) Set(req common.SetRequest) error {
	//log.Println("set", string(req.Key))

	// Try L2 first
	metrics.IncCounter(MetricCmdSetL2)
	start := timer.Now()

	err := l.l2.Set(req)

	metrics.ObserveHist(HistSetL2, timer.Since(start))

	// If we fail to set in L2, don't do anything in L1
	if err != nil {
		metrics.IncCounter(MetricCmdSetErrorsL2)
		metrics.IncCounter(MetricCmdSetErrors)
		return err
	}
	metrics.IncCounter(MetricCmdSetSuccessL2)

	// Replace the entry in L1.
	metrics.IncCounter(MetricCmdSetReplaceL1)
	start = timer.Now()

	err = l.l1.Replace(req)

	metrics.ObserveHist(HistReplaceL1, timer.Since(start))

	if err != nil {
		if err == common.ErrKeyNotFound {
			// For a replace not stored in L1, there's no problem.
			// There is no hot data to replace
			metrics.IncCounter(MetricCmdSetReplaceNotStoredL1)
		} else {
			metrics.IncCounter(MetricCmdSetReplaceErrorsL1)
			metrics.IncCounter(MetricCmdSetErrors)
			return err
		}
	} else {
		metrics.IncCounter(MetricCmdSetReplaceStoredL1)
	}

	metrics.IncCounter(MetricCmdSetSuccess)

	return l.res.Set(req.Opaque, req.Quiet)
}

func (l *L1L2BatchOrca) Add(req common.SetRequest) error {
	//log.Println("add", string(req.Key))

	// Add in L2 first, since it has the larger state
	metrics.IncCounter(MetricCmdAddL2)
	start := timer.Now()

	err := l.l2.Add(req)

	metrics.ObserveHist(HistAddL2, timer.Since(start))

	if err != nil {
		// A key already existing is not an error per se, it's a part of the
		// functionality of the add command to respond with a "not stored" in
		// the form of a ErrKeyExists. Hence no error metrics.
		if err == common.ErrKeyExists {
			metrics.IncCounter(MetricCmdAddNotStoredL2)
			metrics.IncCounter(MetricCmdAddNotStored)
			return err
		}

		// otherwise we have a real error on our hands
		metrics.IncCounter(MetricCmdAddErrorsL2)
		metrics.IncCounter(MetricCmdAddErrors)
		return err
	}

	metrics.IncCounter(MetricCmdAddStoredL2)

	// Replace the entry in L1.
	metrics.IncCounter(MetricCmdAddReplaceL1)
	start = timer.Now()

	err = l.l1.Replace(req)

	metrics.ObserveHist(HistReplaceL1, timer.Since(start))

	if err != nil {
		if err == common.ErrKeyNotFound {
			// For a replace not stored in L1, there's no problem.
			// There is no hot data to replace
			metrics.IncCounter(MetricCmdAddReplaceNotStoredL1)
		} else {
			metrics.IncCounter(MetricCmdAddReplaceErrorsL1)
			metrics.IncCounter(MetricCmdAddErrors)
			return err
		}
	} else {
		metrics.IncCounter(MetricCmdAddReplaceStoredL1)
	}

	metrics.IncCounter(MetricCmdAddStored)

	return l.res.Add(req.Opaque, req.Quiet)
}

func (l *L1L2BatchOrca) Replace(req common.SetRequest) error {
	//log.Println("replace", string(req.Key))

	// Add in L2 first, since it has the larger state
	metrics.IncCounter(MetricCmdReplaceL2)
	start := timer.Now()

	err := l.l2.Replace(req)

	metrics.ObserveHist(HistReplaceL2, timer.Since(start))

	if err != nil {
		// A key already existing is not an error per se, it's a part of the
		// functionality of the replace command to respond with a "not stored"
		// in the form of an ErrKeyNotFound. Hence no error metrics.
		if err == common.ErrKeyNotFound {
			metrics.IncCounter(MetricCmdReplaceNotStoredL2)
			metrics.IncCounter(MetricCmdReplaceNotStored)
			return err
		}

		// otherwise we have a real error on our hands
		metrics.IncCounter(MetricCmdReplaceErrorsL2)
		metrics.IncCounter(MetricCmdReplaceErrors)
		return err
	}

	metrics.IncCounter(MetricCmdReplaceStoredL2)

	// Replace the entry in L1.
	metrics.IncCounter(MetricCmdReplaceReplaceL1)
	start = timer.Now()

	err = l.l1.Replace(req)

	metrics.ObserveHist(HistReplaceL1, timer.Since(start))

	if err != nil {
		if err == common.ErrKeyNotFound {
			// For a replace not stored in L1, there's no problem.
			// There is no hot data to replace
			metrics.IncCounter(MetricCmdReplaceReplaceNotStoredL1)
		} else {
			metrics.IncCounter(MetricCmdReplaceReplaceErrorsL1)
			metrics.IncCounter(MetricCmdReplaceErrors)
			return err
		}
	} else {
		metrics.IncCounter(MetricCmdReplaceReplaceStoredL1)
	}

	metrics.IncCounter(MetricCmdReplaceStored)

	return l.res.Replace(req.Opaque, req.Quiet)
}

func (l *L1L2BatchOrca) Append(req common.SetRequest) error {
	//log.Println("append", string(req.Key))

	// Ordering of append and prepend operations won't matter much unless
	// there's a concurrent set that interleaves. In the case of a delete, the
	// append will fail to work the second time (in L1) and the delete will not
	// interfere. This append technically still succeeds if L1 doesn't work
	// and L2 succeeded because data is allowed to be in L2 and not L1.
	//
	// With a set, we can get data appended (or prepended) only in L1 if a set
	// completes both L2 and L1 between the L2 and L1 of this operation. This is
	// an accepted risk which can be solved by the locking wrapper if it
	// commonly happens.
	metrics.IncCounter(MetricCmdAppendL2)
	start := timer.Now()

	err := l.l2.Append(req)

	metrics.ObserveHist(HistAppendL2, timer.Since(start))

	if err != nil {
		// Appending in L2 did not succeed. Don't try in L1 since this means L2
		// may not have succeeded.
		if err == common.ErrItemNotStored {
			metrics.IncCounter(MetricCmdAppendNotStoredL2)
			metrics.IncCounter(MetricCmdAppendNotStored)
			return err
		}

		metrics.IncCounter(MetricCmdAppendErrorsL2)
		metrics.IncCounter(MetricCmdAppendErrors)
		return err
	}

	// L2 succeeded, so it's time to try L1. If L1 fails with a not found, we're
	// still good since L1 is allowed to not have the data when L2 does. If
	// there's an error, we need to fail because we're not in an unknown state
	// where L1 possibly doesn't have the append when L2 does. We don't recover
	// from this but instead fail the request and let the client retry.
	metrics.IncCounter(MetricCmdAppendL1)
	start = timer.Now()

	err = l.l1.Append(req)

	metrics.ObserveHist(HistAppendL1, timer.Since(start))

	if err != nil {
		// Not stored in L1 is still fine. There's a possibility that a
		// concurrent delete happened or that the data has just been pushed out
		// of L1. Append will not bring data back into L1 as it's not necessarily
		// going to be immediately read.
		if err == common.ErrItemNotStored || err == common.ErrKeyNotFound {
			metrics.IncCounter(MetricCmdAppendNotStoredL1)
			metrics.IncCounter(MetricCmdAppendStored)
			return l.res.Append(req.Opaque, req.Quiet)
		}

		metrics.IncCounter(MetricCmdAppendErrorsL1)
		metrics.IncCounter(MetricCmdAppendErrors)
		return err
	}

	metrics.IncCounter(MetricCmdAppendStoredL1)
	metrics.IncCounter(MetricCmdAppendStored)
	return l.res.Append(req.Opaque, req.Quiet)
}

func (l *L1L2BatchOrca) Prepend(req common.SetRequest) error {
	//log.Println("prepend", string(req.Key))

	metrics.IncCounter(MetricCmdPrependL2)
	start := timer.Now()

	err := l.l2.Prepend(req)

	metrics.ObserveHist(HistPrependL2, timer.Since(start))

	if err != nil {
		// Prepending in L2 did not succeed. Don't try in L1 since this means L2
		// may not have succeeded.
		if err == common.ErrItemNotStored {
			metrics.IncCounter(MetricCmdPrependNotStoredL2)
			metrics.IncCounter(MetricCmdPrependNotStored)
			return err
		}

		metrics.IncCounter(MetricCmdPrependErrorsL2)
		metrics.IncCounter(MetricCmdPrependErrors)
		return err
	}

	// L2 succeeded, so it's time to try L1. If L1 fails with a not found, we're
	// still good since L1 is allowed to not have the data when L2 does. If
	// there's an error, we need to fail because we're not in an unknown state
	// where L1 possibly doesn't have the Prepend when L2 does. We don't recover
	// from this but instead fail the request and let the client retry.
	metrics.IncCounter(MetricCmdPrependL1)
	start = timer.Now()

	err = l.l1.Prepend(req)

	metrics.ObserveHist(HistPrependL1, timer.Since(start))

	if err != nil {
		// Not stored in L1 is still fine. There's a possibility that a
		// concurrent delete happened or that the data has just been pushed out
		// of L1. Prepend will not bring data back into L1 as it's not necessarily
		// going to be immediately read.
		if err == common.ErrItemNotStored || err == common.ErrKeyNotFound {
			metrics.IncCounter(MetricCmdPrependNotStoredL1)
			metrics.IncCounter(MetricCmdPrependStored)
			return l.res.Prepend(req.Opaque, req.Quiet)
		}

		metrics.IncCounter(MetricCmdPrependErrorsL1)
		metrics.IncCounter(MetricCmdPrependErrors)
		return err
	}

	metrics.IncCounter(MetricCmdPrependStoredL1)
	metrics.IncCounter(MetricCmdPrependStored)
	return l.res.Prepend(req.Opaque, req.Quiet)
}

func (l *L1L2BatchOrca) Delete(req common.DeleteRequest) error {
	//log.Println("delete", string(req.Key))

	// Try L2 first
	metrics.IncCounter(MetricCmdDeleteL2)
	start := timer.Now()

	err := l.l2.Delete(req)

	metrics.ObserveHist(HistDeleteL2, timer.Since(start))

	if err != nil {
		// On a delete miss in L2 don't bother deleting in L1. There might be no
		// key at all, or another request may be deleting the same key. In that
		// case the other will finish up. Returning a key not found will trigger
		// error handling to send back an error response.
		if err == common.ErrKeyNotFound {
			metrics.IncCounter(MetricCmdDeleteMissesL2)
			metrics.IncCounter(MetricCmdDeleteMisses)
			return err
		}

		// If we fail to delete in L2, don't delete in L1. This can leave us in
		// an inconsistent state if the request succeeded in L2 but some
		// communication error caused the problem. In the typical deployment of
		// rend, the L1 and L2 caches are both on the same box with
		// communication happening over a unix domain socket. In this case, the
		// likelihood of this error path happening is very small.
		metrics.IncCounter(MetricCmdDeleteErrorsL2)
		metrics.IncCounter(MetricCmdDeleteErrors)
		return err
	}
	metrics.IncCounter(MetricCmdDeleteHitsL2)

	// Now delete in L1. This means we're temporarily inconsistent, but also
	// eliminated the interleaving where the data is deleted from L1, read from
	// L2, set in L1, then deleted in L2. By deleting from L2 first, if L1 goes
	// missing then no other request can undo part of this request.
	metrics.IncCounter(MetricCmdDeleteL1)
	start = timer.Now()

	err = l.l1.Delete(req)

	metrics.ObserveHist(HistDeleteL1, timer.Since(start))

	if err != nil {
		// Delete misses in L1 are fine. If we get here, that means the delete
		// in L2 hit. This isn't a miss per se since the overall effect is a
		// delete. Concurrent deletes might interleave to produce this, or the
		// data might have TTL'd out. Both cases are still fine.
		if err == common.ErrKeyNotFound {
			metrics.IncCounter(MetricCmdDeleteMissesL1)
			metrics.IncCounter(MetricCmdDeleteHits)
			// disregard the miss, don't return the error
			return l.res.Delete(req.Opaque)
		}
		metrics.IncCounter(MetricCmdDeleteErrorsL1)
		metrics.IncCounter(MetricCmdDeleteErrors)
		return err
	}

	metrics.IncCounter(MetricCmdDeleteHitsL1)
	metrics.IncCounter(MetricCmdDeleteHits)

	return l.res.Delete(req.Opaque)
}

func (l *L1L2BatchOrca) Touch(req common.TouchRequest) error {
	//log.Println("touch", string(req.Key))

	// Try L2 first
	metrics.IncCounter(MetricCmdTouchL2)
	start := timer.Now()

	err := l.l2.Touch(req)

	metrics.ObserveHist(HistTouchL2, timer.Since(start))

	if err != nil {
		// On a touch miss in L2 don't bother touch in L1. The data should be
		// TTL'd out within a second. This is yet another place where it's
		// possible to be inconsistent, but only for a short time. Any
		// concurrent requests will see the same behavior as this one. If the
		// touch misses here, any other request will see the same view.
		if err == common.ErrKeyNotFound {
			metrics.IncCounter(MetricCmdTouchMissesL2)
			metrics.IncCounter(MetricCmdTouchMisses)
			return err
		}

		// If we fail to touch in L2, don't touch in L1. If the touch succeeded
		// but for some reason the communication failed, then this is still OK
		// since L1 can TTL out while L2 still has the data. On the next get
		// request the data would still be retrievable, albeit more slowly.
		metrics.IncCounter(MetricCmdTouchErrorsL2)
		metrics.IncCounter(MetricCmdTouchErrors)
		return err
	}
	metrics.IncCounter(MetricCmdTouchHitsL2)

	// Try touching in L1 to touch hot data. I'd avoid doing anything in L1 here
	// but it can be a problem if someone decides to touch data down instead of
	// up to let it naturally expire. If I don't touch or delete in L1 then data
	// in L1 might live longer. Touching keeps hot data hot, while delete is
	// more disruptive.
	metrics.IncCounter(MetricCmdTouchTouchL1)
	start = timer.Now()

	err = l.l1.Touch(req)

	metrics.ObserveHist(HistTouchL1, timer.Since(start))

	if err != nil {
		if err == common.ErrKeyNotFound {
			// For a touch miss in L1, there's no problem.
			metrics.IncCounter(MetricCmdTouchTouchMissesL1)
		} else {
			metrics.IncCounter(MetricCmdTouchTouchErrorsL1)
			metrics.IncCounter(MetricCmdTouchErrors)
			return err
		}
	} else {
		metrics.IncCounter(MetricCmdTouchTouchHitsL1)
	}

	metrics.IncCounter(MetricCmdTouchHits)

	return l.res.Touch(req.Opaque)
}

func (l *L1L2BatchOrca) Get(req common.GetRequest) error {
	metrics.IncCounterBy(MetricCmdGetKeys, uint64(len(req.Keys)))
	//debugString := "get"
	//for _, k := range req.Keys {
	//	debugString += " "
	//	debugString += string(k)
	//}
	//println(debugString)

	metrics.IncCounter(MetricCmdGetL1)
	metrics.IncCounterBy(MetricCmdGetKeysL1, uint64(len(req.Keys)))
	start := timer.Now()

	resChan, errChan := l.l1.Get(req)

	var err error
	//var lastres common.GetResponse
	var l2keys [][]byte
	var l2opaques []uint32
	var l2quiets []bool

	// Read all the responses back from L1.
	// The contract is that the resChan will have GetResponse's for get hits and misses,
	// and the errChan will have any other errors, such as an out of memory error from
	// memcached. If any receive happens from errChan, there will be no more responses
	// from resChan.
	for {
		select {
		case res, ok := <-resChan:
			if !ok {
				resChan = nil
			} else {
				if res.Miss {
					metrics.IncCounter(MetricCmdGetMissesL1)
					l2keys = append(l2keys, res.Key)
					l2opaques = append(l2opaques, res.Opaque)
					l2quiets = append(l2quiets, res.Quiet)
				} else {
					metrics.IncCounter(MetricCmdGetHits)
					metrics.IncCounter(MetricCmdGetHitsL1)
					l.res.Get(res)
				}
			}

		case getErr, ok := <-errChan:
			if !ok {
				errChan = nil
			} else {
				metrics.IncCounter(MetricCmdGetErrors)
				metrics.IncCounter(MetricCmdGetErrorsL1)
				err = getErr
			}
		}

		if resChan == nil && errChan == nil {
			break
		}
	}

	// record metrics before going to L2
	metrics.ObserveHist(HistGetL1, timer.Since(start))

	// leave early on all hits
	if len(l2keys) == 0 {
		if err != nil {
			return err
		}
		return l.res.GetEnd(req.NoopOpaque, req.NoopEnd)
	}

	// Time for the same dance with L2
	req = common.GetRequest{
		Keys:       l2keys,
		NoopEnd:    req.NoopEnd,
		NoopOpaque: req.NoopOpaque,
		Opaques:    l2opaques,
		Quiet:      l2quiets,
	}

	metrics.IncCounter(MetricCmdGetL2)
	metrics.IncCounterBy(MetricCmdGetKeysL2, uint64(len(l2keys)))
	start = timer.Now()

	resChan, errChan = l.l2.Get(req)

	for {
		select {
		case res, ok := <-resChan:
			if !ok {
				resChan = nil
			} else {
				if res.Miss {
					metrics.IncCounter(MetricCmdGetMissesL2)
					// Missing L2 means a true miss
					metrics.IncCounter(MetricCmdGetMisses)
				} else {
					metrics.IncCounter(MetricCmdGetHitsL2)

					// For batch, don't set in l1. Typically batch users will read
					// data once and not again, so setting in L1 will not be valuable.
					// As well the data is typically just about to be replaced, making
					// it doubly useless.

					// overall operation is considered a hit
					metrics.IncCounter(MetricCmdGetHits)
				}

				getres := common.GetResponse{
					Key:    res.Key,
					Flags:  res.Flags,
					Data:   res.Data,
					Miss:   res.Miss,
					Opaque: res.Opaque,
					Quiet:  res.Quiet,
				}

				l.res.Get(getres)
			}

		case getErr, ok := <-errChan:
			if !ok {
				errChan = nil
			} else {
				metrics.IncCounter(MetricCmdGetErrors)
				metrics.IncCounter(MetricCmdGetEErrorsL2)
				err = getErr
			}
		}

		if resChan == nil && errChan == nil {
			break
		}
	}

	metrics.ObserveHist(HistGetL2, timer.Since(start))

	if err == nil {
		return l.res.GetEnd(req.NoopOpaque, req.NoopEnd)
	}

	return err
}

func (l *L1L2BatchOrca) GetE(req common.GetRequest) error {
	// The L1/L2 batch does not support getE, only L1Only does.
	log.Println("[WARN] Use of GetE in L1L2 Batch orchestrator")
	return common.ErrUnknownCmd
}

func (l *L1L2BatchOrca) Gat(req common.GATRequest) error {
	//log.Println("gat", string(req.Key))

	// Perform L2 for correctness, invalidate in L1 later
	metrics.IncCounter(MetricCmdGatL2)
	start := timer.Now()

	res, err := l.l2.GAT(req)

	metrics.ObserveHist(HistGatL2, timer.Since(start))

	// Errors here are genrally fatal to the connection, as something has gone
	// seriously wrong. Bail out early.
	// I should note that this is different than the other commands, where there
	// are some benevolent "errors" that include KeyNotFound or KeyExists. In
	// both Get and GAT the mini-internal-protocol is different because the Get
	// command uses a channel to send results back and an error channel to signal
	// some kind of fatal problem. The result signals non-fatal "errors"; in this
	// case it's ErrKeyNotFound --> res.Miss is true.
	if err != nil {
		metrics.IncCounter(MetricCmdGatErrorsL2)
		metrics.IncCounter(MetricCmdGatErrors)
		return err
	}

	if res.Miss {
		// We got a true miss here. L1 is assumed to be missing the data if L2
		// does not have it.
		metrics.IncCounter(MetricCmdGatMissesL2)
		metrics.IncCounter(MetricCmdGatMisses)

	} else {
		metrics.IncCounter(MetricCmdGatHitsL2)

		// Success finding and touching the data in L2, but still need to touch
		// in L1
		touchreq := common.TouchRequest{
			Key:    req.Key,
			Opaque: req.Opaque,
		}

		// Try touching in L1 to touch hot data. See touch impl for reasoning.
		metrics.IncCounter(MetricCmdGatTouchL1)
		start = timer.Now()

		err = l.l1.Touch(touchreq)

		metrics.ObserveHist(HistTouchL1, timer.Since(start))

		if err != nil {
			if err == common.ErrKeyNotFound {
				// For a touch miss in L1, there's no problem.
				metrics.IncCounter(MetricCmdGatTouchMissesL1)
			} else {
				metrics.IncCounter(MetricCmdGatTouchErrorsL1)
				metrics.IncCounter(MetricCmdGatErrors)
				return err
			}
		} else {
			metrics.IncCounter(MetricCmdGatTouchHitsL1)
		}

		// overall operation succeeded
		metrics.IncCounter(MetricCmdGatHits)
	}

	return l.res.GAT(res)
}

func (l *L1L2BatchOrca) Noop(req common.NoopRequest) error {
	return l.res.Noop(req.Opaque)
}

func (l *L1L2BatchOrca) Quit(req common.QuitRequest) error {
	return l.res.Quit(req.Opaque, req.Quiet)
}

func (l *L1L2BatchOrca) Version(req common.VersionRequest) error {
	return l.res.Version(req.Opaque)
}

func (l *L1L2BatchOrca) Unknown(req common.Request) error {
	return common.ErrUnknownCmd
}

func (l *L1L2BatchOrca) Error(req common.Request, reqType common.RequestType, err error) {
	var opaque uint32
	var quiet bool

	if req != nil {
		opaque = req.GetOpaque()
		quiet = req.IsQuiet()
	}

	l.res.Error(opaque, reqType, err, quiet)
}
