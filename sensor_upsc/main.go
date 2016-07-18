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
Package sensor_upsc implements a sensor that uses upsd to get information from
a UPS. It is pretty basic. To add a UPS start sensor_exporter like:

    sensor_exporter uspc,,UPS@HOST

For localhost, HOST may be ommited.

Currently only a few values are reported since I care only about my UPS.
If you are interested to support more values, sumbit a pull request. It is
an easy job, just add entries to upscVarFloat, sensorsType, sensorsHelp. ;)

You can consult the UPSC manual for available readings and their description:
http://networkupstools.org/docs/user-manual.chunked/apcs01.html

Of interest is also the network protocol:
http://networkupstools.org/docs/developer-guide.chunked/ar01s09.html#_command_reference
*/
package sensor_upsc

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andmarios/sensor_exporter/sensor"
)

var suggestedScrapeInterval = time.Duration(10 * time.Second)
var description = `Upsc is a sensor that uses the upsc program to get information from a UPS.
To use it with the suggested scrape interval (HOST may be ommitted for
localhost):

  sensor_exporter upsc,,UPS@HOST`
var timeOut = 10 * time.Second

type Sensor struct {
	Labels     string
	Host       string
	Ups        string
	Re         *regexp.Regexp
	BeginToken string
	EndToken   string
}

// Strings that are used to detect readings from upsd responses. If you add an
// entry to upscVarFloat and your UPS returns this value, the sensor will
// expose it. Please also add a TYPE and HELP entry.
var (
	upscVarFloat = map[string]string{
		"battery.charge":  "upsc_battery_charge",
		"battery.voltage": "upsc_battery_voltage",
		"input.frequency": "upsc_input_frequency",
		"input.voltage":   "upsc_input_voltage",
		"input.current":   "upsc_input_current",
		"output.voltage":  "upsc_output_voltage",
		"ups.load":        "upsc_ups_load",
		"ups.temperature": "upsc_ups_temperature",
	}
	sensorsType = []string{
		"# TYPE upsc_battery_charge gauge",
		"# TYPE upsc_battery_voltage gauge",
		"# TYPE upsc_input_frequency gauge",
		"# TYPE upsc_input_voltage gauge",
		"# TYPE upsc_input_current gauge",
		"# TYPE upsc_output_voltage gauge",
		"# TYPE upsc_ups_load gauge",
		"# TYPE upsc_ups_temperature gauge",
	}
	sensorsHelp = []string{
		"# HELP upsc_battery_charge gauge Battery charge (percent)",
		"# HELP upsc_battery_voltage Battery voltage (V)",
		"# HELP upsc_input_frequency Input line frequency (Hz)",
		"# HELP upsc_input_voltage Input voltage (V)",
		"# HELP upsc_input_current Input current (A)",
		"# HELP upsc_output_voltage Output voltage (V)",
		"# HELP upsc_ups_load Load on UPS (percent)",
		"# HELP upsc_ups_temperature UPS temperature (degrees C)",
	}
)

func NewSensor(opts string) (sensor.Collector, error) {
	conf := strings.Split(opts, `@`)
	var labels, host, ups string
	switch len(conf) {
	case 2:
		ups = conf[0]
		host = conf[1]
		hostParts := strings.Split(host, `:`) // Do not use port in label
		if len(hostParts) == 1 {              // set default port if needed
			host += ":3493"
		}
		labels = fmt.Sprintf("{ups=\"%s\",host=\"%s\"}", ups, hostParts[0])
	case 1:
		labels = fmt.Sprintf("{ups=\"%s\"}", conf[0])
		host = "localhost:3493"
	default:
		return nil, errors.New("Upsc, could not understand UPS URI. Empty or too many '@'?. Opts: " + opts)
	}
	// Output is like: VAR UPS ups.load "14"
	reString := "VAR " + ups + " ([a-zA-Z.]*) \"(.*)\""
	re, err := regexp.Compile(reString)
	if err != nil {
		return nil, errors.New("Upsc, could not compile regural expression: " + reString + ". Err: " + err.Error())
	}
	conn, err := net.DialTimeout("tcp", host, timeOut)
	if err != nil {
		log.Printf("Adding upsc sensor at %s but could not connect to remote.\n", host)
	} else {
		defer conn.Close()
	}
	s := Sensor{Labels: labels, Host: host, Ups: ups, Re: re,
		BeginToken: "BEGIN LIST VAR " + ups + "\n", EndToken: "END LIST VAR " + ups + "\n"}
	return s, nil
}

func (s Sensor) Scrape() (out string, e error) {
	conn, err := net.DialTimeout("tcp", s.Host, timeOut)
	if err != nil {
		sensor.Incident()
		log.Printf("Upsc %s@%s, failed to connect: %s\n", s.Ups, s.Host, err.Error())
		return "", nil
	}
	defer conn.Close()
	fmt.Fprintf(conn, "LIST VAR "+s.Ups+"\n")
	reader := bufio.NewReader(conn)

	res, err := reader.ReadString('\n')
	if err != nil {
		sensor.Incident()
		log.Printf("Upsc %s@%s, reading returned error: %s\n", s.Ups, s.Host, err.Error())
		return "", nil
	}
	if res == "ERR UNKNOWN-UPS" {
		sensor.Incident()
		log.Printf("Upsc %s@%s, upsd daemon said \"unknown ups\".\n", s.Ups, s.Host)
		return "", nil
	} else if res != s.BeginToken {
		sensor.Incident()
		log.Printf("Upsc %s@%s, upsd daemon returned unknown response: %s.\n", s.Ups, s.Host, res)
		return "", nil
	}

	var v []string
	for {
		res, err = reader.ReadString('\n')
		//		fmt.Println(res)
		if err != nil {
			sensor.Incident()
			log.Printf("Upsc %s@%s, connection error while reading: %s\n", s.Ups, s.Host, err.Error())
			return "", nil
		}
		v = s.Re.FindStringSubmatch(res)
		if len(v) == 3 {
			if value, exists := upscVarFloat[v[1]]; exists {
				reading, err := strconv.ParseFloat(v[2], 64)
				if err != nil {
					sensor.Incident()
					log.Printf("Upsc %s@%s, could not parse %s. Error: %s\n", s.Ups, s.Host, v[1], err.Error())
					break
				}
				out += fmt.Sprintf("%s%s %.2f\n", value, s.Labels, reading)
			}
		}

		if res == s.EndToken {
			break
		}
	}

	return out, nil
}

func init() {
	sensor.RegisterCollector("upsc", NewSensor, suggestedScrapeInterval,
		sensorsType, sensorsHelp, description)
}
