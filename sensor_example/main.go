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

/*
Package sensor_example implements a simple sensor that can be used as a
template.

In general your sensor should have:

(a) a Sensor struct (can be empty) that implements the Scrape() function.
(b) a function with a signature like NewSensor() which creates a new sensor.
(c) use the init() function to register itself to the main package.
*/
package sensor_example

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/andmarios/sensor_exporter/sensor"
)

var suggestedScrapeInterval = time.Duration(1 * time.Second)
var description = `Example is an example sensor that returns a random number in [0.0, 1.0)
with every scrape. To use it with the suggested scrape interval:

  sensor_exporter example`

type Sensor struct {
	Id int
}

func NewSensor(opts string) (sensor.Collector, error) {
	_ = opts // This sensor does not have any option
	s := Sensor{Id: rand.Intn(100)}
	return s, nil
}

func (s Sensor) Scrape() (out string, e error) {
	value := rand.Float64()
	if value == 0 { // A serious incident that should be reported
		sensor.Incident()
		log.Println("Sensor example got a zero!")
	}
	out += fmt.Sprintf("sensor_sample_random{id=\"%d\"} %.1f\n", s.Id, value)
	return out, nil
}

func init() {
	var sensorsType, sensorsHelp []string
	sensorsType = append(sensorsType,
		[]string{"# TYPE sensor_sample_random gauge"}...)
	sensorsHelp = append(sensorsHelp,
		[]string{"# HELP sensor_sample_random A random number in [0.0, 1.0)"}...)
	sensor.RegisterCollector("example", NewSensor, suggestedScrapeInterval,
		sensorsType, sensorsHelp, description)
}
