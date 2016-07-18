# sensor_exporter #

`sensor_exporter` is a simple sensor exporter for Prometheus.

It is still work in progress.

## For end users

`sensor_exporter` currently supports Linux coretemp drivers exposed under
`/sys`, hddtemp in TCP/IP daemon mode and NUT's upsd daemon. To run with
coretemp and hddtemp:

    go get github.com/andmarios/sensor_exporter
	sensor_exporter coretemp hddtemp,,localhost:7634

To list available sensors:

    sensor_exporter -list-sensors

![grafana screenshot](https://raw.githubusercontent.com/andmarios/sensor_exporter/master/grafana.png)

## For developers

You can easily add your own sensor, please have a look at
`sensor_example/main.go`.  Your main task is to create a
`Scrape() (string, error)` which reads your sensor and returns a
[Prometheus compatible formatted string](https://prometheus.io/docs/instrumenting/exposition_formats/)
or an error. An error will lead to `sensor_exporter` stopping, so if you feel a
failed scrape shouldn't be catastrophic, log it and return an empty string
instead.

## Motivation

I wanted to expose my CPU's temperatures to prometheus and grafana. The basic
coretemp reading work existed on my [i7tt](https://github.com/andmarios/i7tt),
so I only needed to convert it for prometheus.

I started with the excellent prometheus' golang_client lib, alas it is targeted
for application instrumentation, thus it exposed 37 metrics for the exporter.
My metrics were only 3 or 5, so it seemed like a waste to store 40 - 42 metrics
just for this.

I started simple, by exposing only the coretemp. Then I added the hddtemp and
then thought there are many sensors I would like to expose at some time, thus
I build this small framework.

## License

GPLv3 or greater.
