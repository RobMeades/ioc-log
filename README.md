# Introduction
This repo contains the logging server for the Internet of Chuffs, written in Golang.

# Installation
Grab the code and build it with:

`go get github.com/RobMeades/ioc-log`

The logging server relies on the files `log_enum_app.h` and `log_strings_app.c` matching those that were built into the ioc-client.  If you happen to be working with local versions then you should copy these over the installed files and run:

`go build github.com/RobMeades/ioc-log`

...to update the `ioc-log` executable before using it.

# Usage
To run the code, do something like:

`./ioc-log 1234 -o ~/chuffs/ioc-client-logs`

...where:

- `1234` is the port number that `ioc-log` should receive logging on,
- `~/chuffs/ioc-client-logs` is the (optional) directory to store log files.

Log files are named in the form:

`12.34.56.78_2017-12-17_15-35-01`

...where `12.34.56.78` is the IP address of the logging source, `2017-12-17` is the date that logging began and `15-35-01` is the time that logging began.  Only one source may be connected at any one time, new connections causing old ones to be dropped.

Two different log files are saved.  The raw received binary logging data is stored in a `.raw` file while the decoded (human readable) logging data is stored in a `.log` file.

A `.log` file will contain lines of the following form:

`2017-11-23_07-18-50_997.841:  BATTERY_VOLTAGE [127] 4153 (0x1039)`

...where `2017-11-23_07-18-50_997.841` is the date/time of the event  on the source device to 1000th of a millisecond accuracy, `BATTERY_VOLTAGE` is the name of the event that occurred, `[127]` is the enum representation of the decoded value `BATTERY_VOLTAGE`, and `4153 (0x1039)` is the parameter that was associated with the event by the logging source in decimal and hex.

To run the log file server in the background, do something like:

`nohup ./ioc-log 1234 -o ~/chuffs/ioc-client-logs > /dev/null &`

To get a nice list of log files with the most recent at the bottom, do something like:

`ls -l -t -r ~/chuffs/ioc-client-logs/*.log`

...where `~/chuffs/ioc-client-logs` is the path to your log file directory.

# Boot Setup
To run `ioc-log` at boot, create a file called something like `/lib/systemd/system/ioc-log.service` with contents something like:

```
[Unit]
Description=IoC log server
After=network-online.target

[Service]
ExecStart=/home/username/ioc-log 1234 -o /home/username/chuffs/ioc-client-logs
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
```
...where `username` is replaced by you user name on the system, etc.

Test this with:

`sudo systemctl start ioc-log`

...using `sudo systemctl status ioc-log` to check that it looks OK and then actually running an end-to-end link uploading logs from the [ioc-client](https://github.com/RobMeades/ioc-client).  If all looks good, set it to run at boot with:

`sudo systemctl enable ioc-log`

Reboot and check that it starts correctly; if it does not, check what happened with `sudo journalctl -b` and/or `sudo dmesg`.
