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

package sensor_coretemp

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andmarios/sensor_exporter/sensor"
)

var suggestedScrapeInterval = time.Duration(4800 * time.Millisecond)
var description = `Coretemp is sensor that reads CPU temperature from the coretemp driver on
Linux. It uses the files that the coretemp driver exposes under /sys. It does
not take any options. To use it with the suggested scrape period:

  sensor_exporter coretemp`

var initialized = false

type Sensor struct {
}

func NewSensor(opts string) (sensor.Collector, error) {
	if initialized == true {
		return nil, errors.New("Coretemp sensor may only be used once per instance.")
	}

	err := detectCoreTempSensors()
	if err != nil {
		return nil, errors.New("Coretemp could not initialize sensors: " + err.Error())
	}

	numOfInputs = len(cpuTempFiles)
	if numOfInputs == 0 {
		return nil, errors.New("Coretemp could not find any sensors.")
	}

	s := Sensor{}
	return s, nil
}

func (s Sensor) Scrape() (out string, e error) {
	for k, file := range cpuTempFiles {
		// Read from sysfs
		dat, err := ioutil.ReadFile(file)
		if err != nil {
			return "", errors.New("Coretemp could not scrape: " + err.Error())
		}
		valueString := strings.TrimSuffix(string(dat), "\n")
		value, err := strconv.ParseFloat(string(valueString), 64)
		if err != nil {
			return "", errors.New("Coretemp could not scrape: " + err.Error())
		}
		value = value / 1000
		// Write value
		out += fmt.Sprintf("cpu_temperature_celsius{sensor=\"%s\"} %.1f\n", cpuLabel[k], value)
	}

	return out, nil
}

func init() {
	var sensorsType, sensorsHelp []string
	sensorsType = append(sensorsType,
		[]string{"# TYPE cpu_temperature_celsius gauge"}...)
	sensorsHelp = append(sensorsHelp,
		[]string{"# HELP cpu_temperature_celsius Current temperature of the CPU."}...)
	sensor.RegisterCollector("coretemp", NewSensor, suggestedScrapeInterval,
		sensorsType, sensorsHelp, description)
}

// Here are stored the filenames of the sysfs files we use.
var (
	cpuTempFiles  []string
	cpuLabelFiles []string
)

// Here are stored the contents of the files described above.
var (
	cpuLabel    []string
	numOfInputs int
)

// detect_sensors tries to find sysfs files created from coretemp driver
// that contain the info we seek. It then reads once the contents of files
// that do not change over time: sensor labels
func detectCoreTempSensors() error {
	// Each regexp matches a sysfs file we seek.
	inputs, _ := regexp.Compile("coretemp.*temp([0-9]+)_input")
	labels, _ := regexp.Compile("coretemp.*temp([0-9]+)_label")

	// Check populates our filename arrays with matches.
	matchSensorFiles := func(path string, f os.FileInfo, err error) error {
		if inputs.MatchString(path) {
			cpuTempFiles = append(cpuTempFiles, path)
		} else if labels.MatchString(path) {
			cpuLabelFiles = append(cpuLabelFiles, path)
		}
		return nil
	}

	_ = filepath.Walk("/sys/devices/platform/", matchSensorFiles)

	// Read temperature labels from /sys
	for _, file := range cpuLabelFiles {
		dat, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		value := strings.TrimSuffix(string(dat), "\n")
		cpuLabel = append(cpuLabel, value)
	}
	return nil
}
