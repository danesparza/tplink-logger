# tplink-logger [![CircleCI](https://circleci.com/gh/danesparza/tplink-logger.svg?style=shield)](https://circleci.com/gh/danesparza/tplink-logger)
InfluxDB logger for HS110 energy monitoring outlet

## Quick start
Grab the [latest release](https://github.com/danesparza/tplink-logger/releases/latest) for your platform and run it.  

## Command line parameters
```
tplink-logger.exe --help

Usage of tplink-logger:
  -influxdb string
        Influx database to log to (default "sensors")
  -influxurl string
        InfluxDB url - Ex: http://yourserver:8086
  -ip string
        TPLink HS110 ip address (default "192.168.1.1")
  -loglevel string
        Set the console log level (default "INFO")
```
