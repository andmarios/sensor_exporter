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

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/andmarios/sensor_exporter/sensor"
	_ "github.com/andmarios/sensor_exporter/sensor_coretemp"
	_ "github.com/andmarios/sensor_exporter/sensor_example"
	_ "github.com/andmarios/sensor_exporter/sensor_hddtemp"
)

type Scraper struct {
	Collector sensor.Collector
	Interval  time.Duration
	Type      string
	Value     string
	Mutex     *sync.RWMutex
}

var scrapers []*Scraper
var supportTexts = make(map[string]bool)

var (
	defaultInterval = time.Duration(4800) * time.Millisecond
)

var (
	port        = flag.String("p", "9091", "port to listen on")
	listSensors = flag.Bool("list-sensors", false, "list available sensors")
)

func main() {
	flag.Parse()

	if *listSensors {
		for k, v := range sensor.AvailableCollectors {
			fmt.Printf("SENSOR %s\nDefault scrape interval: %s\n", k, v.DefaultInterval)
			fmt.Printf("%s\n\n", v.Description)
		}
		return
	}
	for k, _ := range sensor.AvailableCollectors {
		log.Printf("Found sensor type %s\n", k)
	}

	for _, v := range flag.Args() {
		scraper, err := processArg(v)
		if err != nil {
			log.Fatalf("Could not add “%s”. Err: %s\n", v, err)
		}
		scrapers = append(scrapers, scraper)
	}

	log.Println("Initializing sensors")
	for _, v := range scrapers {
		startSensor(v)
	}

	log.Printf("Initialization succesful. Listening on :%s\n", *port)
	http.HandleFunc("/metrics", metricsHandler)
	http.ListenAndServe(":"+*port, nil)

}

func startSensor(s *Scraper) {
	go func() {
		tick := time.Tick(s.Interval)
		for {
			select {
			case <-tick:
				value, err := s.Collector.Scrape()
				if err != nil {
					log.Printf("Could not scrape %s. Err: %s\n", s.Type, err)
					continue
				}
				s.Mutex.Lock()
				s.Value = value
				s.Mutex.Unlock()
			}
		}
	}()
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	for k, _ := range supportTexts {
		fmt.Fprintln(w, k)
	}
	for _, v := range scrapers {
		v.Mutex.RLock()
		fmt.Fprintln(w, v.Value)
		v.Mutex.RUnlock()
	}
}

func processArg(arg string) (*Scraper, error) {
	conf := strings.SplitN(arg, ",", 3)
	//var scraper sensor.Scraper
	var interval time.Duration
	var opts string
	var err error

	switch len(conf) {
	case 3: // Set opts
		opts = conf[2]
		fallthrough
	case 2: // Set interval if given
		interval, err = time.ParseDuration(conf[1])
		if err != nil {
			log.Printf("Could not understand scrape interval: %s. Using default.\n", conf[1])
			interval = 0
		}
		fallthrough
	case 1: // Check sensor and if needed default intervals
		if _, exists := sensor.AvailableCollectors[conf[0]]; !exists {
			return nil, errors.New("Sensor " + conf[0] + " not found")
		}
		if interval == 0 { // Try to assign scraper's suggested interval
			interval = sensor.AvailableCollectors[conf[0]].DefaultInterval
		}
		if interval == 0 { // Assign our interval if all else failed
			interval = defaultInterval
		}
		// Add sensors TYPE and HELP texts if needed to our supportTexts list
		for k, _ := range sensor.AvailableCollectors[conf[0]].Type {
			supportTexts[sensor.AvailableCollectors[conf[0]].Type[k]] = true
			supportTexts[sensor.AvailableCollectors[conf[0]].Help[k]] = true
		}

	default:
		return nil, errors.New("Could not create sensor")
	}

	log.Printf("Adding scraper for sensor %s with interval %s and opts: %s\n", conf[0], interval, opts)

	collector, err := sensor.AvailableCollectors[conf[0]].New(opts)
	if err != nil {
		return nil, errors.New("Could not init sensor: " + err.Error())
	}
	value, err := collector.Scrape()
	if err != nil {
		return nil, errors.New("Could not perform first scrape: " + err.Error())
	}
	scraper := &Scraper{collector, interval, conf[0], value, &sync.RWMutex{}}
	return scraper, nil
}
