//
// Copyright 2016 Marios Andreopoulos
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package sensor

import (
	"sync/atomic"
	"time"
)

// A Collector Every sensor must implement this interface. When called the sensor must read
// data from its source and return a prometheus compatible values string.
type Collector interface {
	Scrape() (string, error)
}

// A CollectorEntry contains information about a Collector:
// - the function that creates a new Collector
// - the suggested scrape interval for this Collector
// - a list of Prometheus TYPE and HELP strings for the Collector
//   see <https://prometheus.io/docs/instrumenting/exposition_formats/>
// - a description of the Collector. It is a good idea to document its opts
//   here too.
type CollectorEntry struct {
	New             func(string) (Collector, error)
	DefaultInterval time.Duration
	Type            []string
	Help            []string
	Description     string
}

// The list of available collectors
var AvailableCollectors = make(map[string]CollectorEntry)

var incidents uint64 = 0

// RegisterCollector shoukd be called at the init function of each sensor
// package to register itself to sensor_exporter. It is like golang's
// image and image/jpg, image/gif relation.
func RegisterCollector(name string, f func(string) (Collector, error),
	suggestedInterval time.Duration, sensorsType, sensorsHelp []string, description string) {
	AvailableCollectors[name] = CollectorEntry{
		New:             f,
		DefaultInterval: suggestedInterval,
		Type:            sensorsType,
		Help:            sensorsHelp,
		Description:     description,
	}
}

// Incident increases atomically the number of incidents for the program. This
// can be used by the sensors and the main package to export statistics about
// serious but not fatal errors (e.g a scrape taking too long or failing). If
// the user loads the sensor "log", then these statistics will be exported as
// a counter value.
func Incident() {
	atomic.AddUint64(&incidents, 1)
}

// GetIncident gets atomically the number of incidents for the program. Although
// exposed, its primary target is the log sensor.
func GetIncident() uint64 {
	return atomic.LoadUint64(&incidents)
}
