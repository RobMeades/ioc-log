# Introduction
This repo contains the logging server for the Internet of Chuffs, written in Golang.

# Installation
Grab the code and build it with:

`go get github.com/u-blox/ioc-log`

The logging server relies on the files `log_enum.h` and `log_srings.c` matching those that were built into the ioc-client.  If you happen to be working with local versions then you should copy these over the installed files and run:

`go build github.com/u-blox/ioc-log`

...to update the `ioc-log` executable before using it.

# Usage
To run the code, do something like:

`./ioc-log 1234 -o log`

...where:

- `1234` is the port number that `ioc-log` should receive logging on,
- `log` is the (optional) directory to store log files.

Log files are named in the form:

`x12.34.56.78_2017-12-17_15-35-01.log`

...where `12.34.56.78` is the IP address of the logging source, `2017-12-17` is the date that logging began and `15-35-01` is the time that logging began.  If `x` is present at the start of the log file name then the time and date are those of the ioc-server, otherwise they are those of the ioc-client.  Only one source may be connected at any one time, new connections causing old ones to be dropped.

Each log file will contain lines of the following form:

`1970-01-01_00-00-00_000.000:   AN_EVENT 0 (0x0)`

...where `1970-01-01_00-00-00_000.000` is the date/time of the event  on the source device to 1000th of a millisecond accuracy, `AN_EVENT` is the name of the event that occurred, and `0 (0x0)` is the parameter that was associated with the event by the logging source in decimal and hex.
