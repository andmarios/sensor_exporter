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
Package sensor_log implements incident reporting for sensor_exporter. Users
may use it to easily detect abnormal behaviour.
*/
package sensor_log

import (
	"errors"
	"fmt"
	"time"

	"github.com/andmarios/sensor_exporter/sensor"
)

var suggestedScrapeInterval = time.Duration(3 * time.Second)
var description = `Log is a sensor that exposes sensor_exporter incidents. This is a metric about
serious issues that an administrator should investigate, such as scrapes failing
or taking too long. To use it with the suggested scrape interval:

  sensor_exporter log`

type Sensor struct {
}

var initialized = false

func NewSensor(opts string) (sensor.Collector, error) {
	if initialized == true {
		return nil, errors.New("Only one log sensor may be set.")
	}
	initialized = true
	s := Sensor{}
	return s, nil
}

func (s Sensor) Scrape() (out string, e error) {
	out += fmt.Sprintf("sensor_exporter_incidents %d\n", sensor.GetIncident())
	return out, nil
}

func init() {
	var sensorsType, sensorsHelp []string
	sensorsType = append(sensorsType,
		[]string{"# TYPE sensor_exporter_incidents counter"}...)
	sensorsHelp = append(sensorsHelp,
		[]string{"# HELP sensor_exporter_incidents Counter of serious incidents for sensor_exporter that an admin should investigate."}...)
	sensor.RegisterCollector("log", NewSensor, suggestedScrapeInterval,
		sensorsType, sensorsHelp, description)
}
