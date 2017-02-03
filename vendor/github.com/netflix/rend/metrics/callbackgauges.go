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

package metrics

import "sync/atomic"

// IntGaugeCallback defines a function that the metrics package can use to
// retrieve an integer gauge value
type IntGaugeCallback func() uint64

// FloatGaugeCallback defines a function that the metrics package can use to
// retrieve an floating point gauge value
type FloatGaugeCallback func() float64

const maxNumCallbacks = 10240

var (
	curIntCbID   = new(uint32)
	intcbnames   = make([]string, maxNumCallbacks)
	intcallbacks = make([]IntGaugeCallback, maxNumCallbacks)
	intcbtags    = make([]Tags, maxNumCallbacks)

	curFloatCbID   = new(uint32)
	floatcbnames   = make([]string, maxNumCallbacks)
	floatcallbacks = make([]FloatGaugeCallback, maxNumCallbacks)
	floatcbtags    = make([]Tags, maxNumCallbacks)
)

// RegisterIntGaugeCallback registers a gauge callback which will be called every
// time metrics are requested.
// There is a maximum of 10240 int callbacks, after which adding a new one will panic.
func RegisterIntGaugeCallback(name string, tgs Tags, cb IntGaugeCallback) {
	id := atomic.AddUint32(curIntCbID, 1) - 1

	if id >= maxNumCallbacks {
		panic("Too many callbacks")
	}

	intcallbacks[id] = cb
	intcbnames[id] = name

	tgs = copyTags(tgs)
	tgs[TagMetricType] = MetricTypeGauge
	intcbtags[id] = tgs
}

// RegisterFloatGaugeCallback registers a gauge callback which will be called every
// time metrics are requested.
// There is a maximum of 10240 float callbacks, after which adding a new one will panic.
func RegisterFloatGaugeCallback(name string, tgs Tags, cb FloatGaugeCallback) {
	id := atomic.AddUint32(curFloatCbID, 1) - 1

	if id >= maxNumCallbacks {
		panic("Too many callbacks")
	}

	floatcallbacks[id] = cb
	floatcbnames[id] = name

	tgs = copyTags(tgs)
	tgs[TagMetricType] = MetricTypeGauge
	floatcbtags[id] = tgs
}

func getAllCallbackGauges() ([]IntMetric, []FloatMetric) {
	numIDs := int(atomic.LoadUint32(curIntCbID))
	retint := make([]IntMetric, numIDs)

	for i := 0; i < numIDs; i++ {
		retint[i] = IntMetric{
			Name: intcbnames[i],
			Val:  intcallbacks[i](),
			Tgs:  intcbtags[i],
		}
	}

	numIDs = int(atomic.LoadUint32(curFloatCbID))
	retfloat := make([]FloatMetric, numIDs)

	for i := 0; i < numIDs; i++ {
		retfloat[i] = FloatMetric{
			Name: floatcbnames[i],
			Val:  floatcallbacks[i](),
			Tgs:  floatcbtags[i],
		}
	}

	return retint, retfloat
}
