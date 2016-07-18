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
Package sensor_hddtemp reads hard disk temperatures from a hddtemp instance
running in TCP/IP daemon mode.

It expects the hddtemp daemon's url as option:
    "hddtemp,,localhost:7634"

Note that whilst hddtemp won't wake up a disk to read its temperature, it
will prevent a disk from sleeping since every query resets the disk's timer.
It is advised to use if for SSD's mostly or always on hard disks.
*/
package sensor_hddtemp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andmarios/sensor_exporter/sensor"
)

var suggestedScrapeInterval = time.Duration(4800 * time.Millisecond)
var description = `Hddtemp reads and exposes disk temperatures from a hddtemp daemon. Its options
is the URL of the daemon. Example setup with default scrape interval:

  sensor_exporter hddtemp,,localhost:7634`
var timeOut = 3 * time.Second
var re = regexp.MustCompile(`(\|[^\|]*){3,3}\|[CF*]`)

type Sensor struct {
	Url  string
	Host string
}

func NewSensor(opts string) (sensor.Collector, error) {
	hostRe := regexp.MustCompile("^(.*):[^:]*$")
	host := hostRe.FindStringSubmatch(opts)
	s := Sensor{Url: opts, Host: host[1]}

	conn, err := net.DialTimeout("tcp", s.Url, timeOut)
	if err != nil {
		log.Printf("Adding hddtemp sensor at %s but could not connect to remote.\n", s.Url)
	} else {
		defer conn.Close()
	}
	return s, nil
}

var (
	device, model, degrees string
	temp                   float64
)

func (s Sensor) Scrape() (out string, e error) {
	conn, err := net.DialTimeout("tcp", s.Url, timeOut)
	if err != nil {
		sensor.Incident()
		log.Printf("Hddtemp @ %s, failed to connect: %s\n", s.Url, err.Error())
		return "", nil
	}
	defer conn.Close()

	reader, _ := bufio.NewReader(conn).ReadString('\n')
	// We get something like: |diskA|model|temp|degree|diskB|model|temp|degree
	// And the regexp below it breaks it to parts: |disk|model|temp|degree
	for _, v := range re.FindAllString(reader, -1) {
		v = strings.TrimLeft(v, "|") // Remove leading |
		v2 := strings.Split(v, "|")
		// Not a real need to assign these values but helps readability
		device = v2[0]
		model = v2[1]
		degrees = v2[3]
		if degrees != "*" { // We read a temperature
			temp, err = strconv.ParseFloat(v2[2], 64)
			if err != nil {
				sensor.Incident()
				log.Printf("Hddtemp: hddtemp daemon returned a funny string: %s\n", v)
				continue
			}
			if degrees == "F" {
				temp = (temp - 32) / 1.8 // Convert to Celsius
			}
			out += fmt.Sprintf("hdd_temperature_celsius{host=\"%s\",disk=\"%s\",model=\"%s\"} %.0f\n",
				s.Host, device, model, temp)
		}
	}

	return out, nil
}

func init() {
	var sensorsType, sensorsHelp []string
	sensorsType = append(sensorsType,
		[]string{"# TYPE hdd_temperature_celsius gauge"}...)
	sensorsHelp = append(sensorsHelp,
		[]string{"# HELP hdd_temperature_celsius Current temperature of the disk."}...)
	sensor.RegisterCollector("hddtemp", NewSensor, suggestedScrapeInterval,
		sensorsType, sensorsHelp, description)
}
