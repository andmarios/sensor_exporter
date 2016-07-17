# sensor_exporter #

`sensor_exporter` is a simple sensor exporter for Prometheus.

It is still work in progress.

## For end users

`sensor_exporter` currently supports Linux coretemp drivers exposed under `/sys`
and hddtemp in TCP/IP daemon mode. To run with both:

    go get github.com/andmarios/sensor_exporter
	sensor_exporter coretemp hddtemp,,localhost:7634

To list available sensors:

    sensor_exporter -list-sensors

## For developers

You can easily add your own sensor, please have a look at
`sensor_example/main.go`.  Your main task is to create a
`Scrape() (string, error)` which reads your sensor and returns a
[Prometheus compatible formatted string](https://prometheus.io/docs/instrumenting/exposition_formats/)
or an error. An error will lead to `sensor_exporter` stopping, so if you feel a
failed scrape shouldn't be catastrophic, log it and return an empty string
instead.

## License

GPLv3 or greater.
