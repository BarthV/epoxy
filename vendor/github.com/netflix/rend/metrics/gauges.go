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

import (
	"math"
	"sync/atomic"
)

const maxNumGauges = 1024

var (
	curIntGaugeID = new(uint32)
	intgnames     = make([]string, maxNumGauges)
	intgauges     = make([]uint64, maxNumGauges)
	intgtags      = make([]Tags, maxNumGauges)

	curFloatGaugeID = new(uint32)
	floatgnames     = make([]string, maxNumGauges)
	floatgauges     = make([]uint64, maxNumGauges)
	floatgtags      = make([]Tags, maxNumGauges)
)

// AddIntGauge registers an integer-based gauge and returns an ID that can be
// used to update it.
// There is a maximum of 1024 gauges, after which adding a new one will panic
func AddIntGauge(name string, tgs Tags) uint32 {
	id := atomic.AddUint32(curIntGaugeID, 1) - 1

	if id >= maxNumGauges {
		panic("Too many int gauges")
	}

	intgnames[id] = name

	tgs = copyTags(tgs)
	tgs[TagMetricType] = MetricTypeGauge
	tgs[TagDataType] = DataTypeUint64
	intgtags[id] = tgs

	return id
}

// AddFloatGauge registers a float-based gauge and returns an ID that can be
// used to access it.
// There is a maximum of 1024 gauges, after which adding a new one will panic
func AddFloatGauge(name string, tgs Tags) uint32 {
	id := atomic.AddUint32(curFloatGaugeID, 1) - 1

	if id >= maxNumGauges {
		panic("Too many float gauges")
	}

	floatgnames[id] = name

	tgs = copyTags(tgs)
	tgs[TagMetricType] = MetricTypeGauge
	tgs[TagDataType] = DataTypeFloat64
	floatgtags[id] = tgs

	return id
}

// SetIntGauge sets a gauge by the ID returned from AddIntGauge to the value given.
func SetIntGauge(id uint32, value uint64) {
	atomic.StoreUint64(&intgauges[id], value)
}

// SetFloatGauge sets a gauge by the ID returned from AddFloatGauge to the value given.
func SetFloatGauge(id uint32, value float64) {
	// The float64 value needs to be converted into an int64 here because
	// there is no atomic store for float values. This is a literal
	// reinterpretation of the same exact bits.
	v2 := math.Float64bits(value)
	atomic.StoreUint64(&floatgauges[id], v2)
}

func getAllGauges() ([]IntMetric, []FloatMetric) {
	numIDs := int(atomic.LoadUint32(curIntGaugeID))
	retint := make([]IntMetric, numIDs)

	for i := 0; i < numIDs; i++ {
		retint[i] = IntMetric{
			Name: intgnames[i],
			Val:  atomic.LoadUint64(&intgauges[i]),
			Tgs:  intgtags[i],
		}
	}

	numIDs = int(atomic.LoadUint32(curFloatGaugeID))
	retfloat := make([]FloatMetric, numIDs)

	for i := 0; i < numIDs; i++ {
		// The int64 bit pattern of the float value needs to be converted back
		// into a float64 here. This is a literal reinterpretation of the same
		// exact bits.
		intval := atomic.LoadUint64(&floatgauges[i])

		retfloat[i] = FloatMetric{
			Name: floatgnames[i],
			Val:  math.Float64frombits(intval),
			Tgs:  floatgtags[i],
		}
	}

	return retint, retfloat
}
