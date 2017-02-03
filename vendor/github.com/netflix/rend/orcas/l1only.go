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
	"github.com/netflix/rend/common"
	"github.com/netflix/rend/handlers"
	"github.com/netflix/rend/metrics"
	"github.com/netflix/rend/timer"
)

type L1OnlyOrca struct {
	l1  handlers.Handler
	res common.Responder
}

func L1Only(l1, l2 handlers.Handler, res common.Responder) Orca {
	return &L1OnlyOrca{
		l1:  l1,
		res: res,
	}
}

func (l *L1OnlyOrca) Set(req common.SetRequest) error {
	//log.Println("set", string(req.Key))

	metrics.IncCounter(MetricCmdSetL1)
	start := timer.Now()

	err := l.l1.Set(req)

	metrics.ObserveHist(HistSetL1, timer.Since(start))

	if err == nil {
		metrics.IncCounter(MetricCmdSetSuccessL1)
		metrics.IncCounter(MetricCmdSetSuccess)

		err = l.res.Set(req.Opaque, req.Quiet)

	} else {
		metrics.IncCounter(MetricCmdSetErrorsL1)
		metrics.IncCounter(MetricCmdSetErrors)
	}

	return err
}

func (l *L1OnlyOrca) Add(req common.SetRequest) error {
	//log.Println("add", string(req.Key))

	metrics.IncCounter(MetricCmdAddL1)
	start := timer.Now()

	err := l.l1.Add(req)

	metrics.ObserveHist(HistAddL1, timer.Since(start))

	if err == nil {
		metrics.IncCounter(MetricCmdAddStoredL1)
		metrics.IncCounter(MetricCmdAddStored)

		err = l.res.Add(req.Opaque, req.Quiet)

	} else if err == common.ErrKeyExists {
		metrics.IncCounter(MetricCmdAddNotStoredL1)
		metrics.IncCounter(MetricCmdAddNotStored)
	} else {
		metrics.IncCounter(MetricCmdAddErrorsL1)
		metrics.IncCounter(MetricCmdAddErrors)
	}

	return err
}

func (l *L1OnlyOrca) Replace(req common.SetRequest) error {
	//log.Println("replace", string(req.Key))

	metrics.IncCounter(MetricCmdReplaceL1)
	start := timer.Now()

	err := l.l1.Replace(req)

	metrics.ObserveHist(HistReplaceL1, timer.Since(start))

	if err == nil {
		metrics.IncCounter(MetricCmdReplaceStoredL1)
		metrics.IncCounter(MetricCmdReplaceStored)

		err = l.res.Replace(req.Opaque, req.Quiet)

	} else if err == common.ErrKeyNotFound {
		metrics.IncCounter(MetricCmdReplaceNotStoredL1)
		metrics.IncCounter(MetricCmdReplaceNotStored)
	} else {
		metrics.IncCounter(MetricCmdReplaceErrorsL1)
		metrics.IncCounter(MetricCmdReplaceErrors)
	}

	return err
}

func (l *L1OnlyOrca) Append(req common.SetRequest) error {
	//log.Println("append", string(req.Key))

	metrics.IncCounter(MetricCmdAppendL1)
	start := timer.Now()

	err := l.l1.Append(req)

	metrics.ObserveHist(HistAppendL1, timer.Since(start))

	if err == nil {
		metrics.IncCounter(MetricCmdAppendStoredL1)
		metrics.IncCounter(MetricCmdAppendStored)

		err = l.res.Append(req.Opaque, req.Quiet)

	} else if err == common.ErrKeyNotFound {
		metrics.IncCounter(MetricCmdAppendNotStoredL1)
		metrics.IncCounter(MetricCmdAppendNotStored)
	} else {
		metrics.IncCounter(MetricCmdAppendErrorsL1)
		metrics.IncCounter(MetricCmdAppendErrors)
	}

	return err
}

func (l *L1OnlyOrca) Prepend(req common.SetRequest) error {
	//log.Println("prepend", string(req.Key))

	metrics.IncCounter(MetricCmdPrependL1)
	start := timer.Now()

	err := l.l1.Prepend(req)

	metrics.ObserveHist(HistPrependL1, timer.Since(start))

	if err == nil {
		metrics.IncCounter(MetricCmdPrependStoredL1)
		metrics.IncCounter(MetricCmdPrependStored)

		err = l.res.Prepend(req.Opaque, req.Quiet)

	} else if err == common.ErrKeyNotFound {
		metrics.IncCounter(MetricCmdPrependNotStoredL1)
		metrics.IncCounter(MetricCmdPrependNotStored)
	} else {
		metrics.IncCounter(MetricCmdPrependErrorsL1)
		metrics.IncCounter(MetricCmdPrependErrors)
	}

	return err
}

func (l *L1OnlyOrca) Delete(req common.DeleteRequest) error {
	//log.Println("delete", string(req.Key))

	metrics.IncCounter(MetricCmdDeleteL1)
	start := timer.Now()

	err := l.l1.Delete(req)

	metrics.ObserveHist(HistDeleteL1, timer.Since(start))

	if err == nil {
		metrics.IncCounter(MetricCmdDeleteHits)
		metrics.IncCounter(MetricCmdDeleteHitsL1)

		l.res.Delete(req.Opaque)

	} else if err == common.ErrKeyNotFound {
		metrics.IncCounter(MetricCmdDeleteMissesL1)
		metrics.IncCounter(MetricCmdDeleteMisses)
	} else {
		metrics.IncCounter(MetricCmdDeleteErrorsL1)
		metrics.IncCounter(MetricCmdDeleteErrors)
	}

	return err
}

func (l *L1OnlyOrca) Touch(req common.TouchRequest) error {
	//log.Println("touch", string(req.Key))

	metrics.IncCounter(MetricCmdTouchL1)
	start := timer.Now()

	err := l.l1.Touch(req)

	metrics.ObserveHist(HistTouchL1, timer.Since(start))

	if err == nil {
		metrics.IncCounter(MetricCmdTouchHitsL1)
		metrics.IncCounter(MetricCmdTouchHits)

		l.res.Touch(req.Opaque)

	} else if err == common.ErrKeyNotFound {
		metrics.IncCounter(MetricCmdTouchMissesL1)
		metrics.IncCounter(MetricCmdTouchMisses)
	} else {
		metrics.IncCounter(MetricCmdTouchMissesL1)
		metrics.IncCounter(MetricCmdTouchMisses)
	}

	return err
}

func (l *L1OnlyOrca) Get(req common.GetRequest) error {
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

	// Read all the responses back from l.l1.
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
					metrics.IncCounter(MetricCmdGetMisses)
				} else {
					metrics.IncCounter(MetricCmdGetHits)
					metrics.IncCounter(MetricCmdGetHitsL1)
				}
				l.res.Get(res)
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

	metrics.ObserveHist(HistGetL1, timer.Since(start))

	if err == nil {
		l.res.GetEnd(req.NoopOpaque, req.NoopEnd)
	}

	return err
}

func (l *L1OnlyOrca) GetE(req common.GetRequest) error {
	// For an L1 only orchestrator, this will fail if the backend is memcached.
	// It should be talking to another rend-based server, such as the L2 for the
	// EVCache server project.
	metrics.IncCounterBy(MetricCmdGetEKeys, uint64(len(req.Keys)))
	//debugString := "gete"
	//for _, k := range req.Keys {
	//	debugString += " "
	//	debugString += string(k)
	//}
	//println(debugString)

	metrics.IncCounter(MetricCmdGetEL1)
	metrics.IncCounterBy(MetricCmdGetEKeysL1, uint64(len(req.Keys)))
	start := timer.Now()

	resChan, errChan := l.l1.GetE(req)

	var err error

	// Read all the responses back from l.l1.
	// The contract is that the resChan will have GetEResponse's for get hits and misses,
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
					metrics.IncCounter(MetricCmdGetEMissesL1)
					metrics.IncCounter(MetricCmdGetEMisses)
				} else {
					metrics.IncCounter(MetricCmdGetEHits)
					metrics.IncCounter(MetricCmdGetEHitsL1)
				}
				l.res.GetE(res)
			}

		case getErr, ok := <-errChan:
			if !ok {
				errChan = nil
			} else {
				metrics.IncCounter(MetricCmdGetEErrors)
				metrics.IncCounter(MetricCmdGetEErrorsL1)
				err = getErr
			}
		}

		if resChan == nil && errChan == nil {
			break
		}
	}

	metrics.ObserveHist(HistGetEL1, timer.Since(start))

	if err == nil {
		l.res.GetEnd(req.NoopOpaque, req.NoopEnd)
	}

	return err
}

func (l *L1OnlyOrca) Gat(req common.GATRequest) error {
	//log.Println("gat", string(req.Key))

	metrics.IncCounter(MetricCmdGatL1)
	start := timer.Now()

	res, err := l.l1.GAT(req)

	metrics.ObserveHist(HistGatL1, timer.Since(start))

	if err == nil {
		if res.Miss {
			metrics.IncCounter(MetricCmdGatMissesL1)
			// TODO: Account for L2
			metrics.IncCounter(MetricCmdGatMisses)
		} else {
			metrics.IncCounter(MetricCmdGatHits)
			metrics.IncCounter(MetricCmdGatHitsL1)
		}
		l.res.GAT(res)
		// There is no GetEnd call required here since this is only ever
		// done in the binary protocol, where there's no END marker.
		// Calling l.res.GetEnd was a no-op here and is just useless.
		//l.res.GetEnd(0, false)
	} else {
		metrics.IncCounter(MetricCmdGatErrors)
		metrics.IncCounter(MetricCmdGatErrorsL1)
	}

	return err
}

func (l *L1OnlyOrca) Noop(req common.NoopRequest) error {
	return l.res.Noop(req.Opaque)
}

func (l *L1OnlyOrca) Quit(req common.QuitRequest) error {
	return l.res.Quit(req.Opaque, req.Quiet)
}

func (l *L1OnlyOrca) Version(req common.VersionRequest) error {
	return l.res.Version(req.Opaque)
}

func (l *L1OnlyOrca) Unknown(req common.Request) error {
	return common.ErrUnknownCmd
}

func (l *L1OnlyOrca) Error(req common.Request, reqType common.RequestType, err error) {
	var opaque uint32
	var quiet bool

	if req != nil {
		opaque = req.GetOpaque()
		quiet = req.IsQuiet()
	}

	l.res.Error(opaque, reqType, err, quiet)
}
