package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/danesparza/tplink"
	"github.com/hashicorp/logutils"
	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/montanaflynn/stats"
)

var (
	//	Flags
	monitorIP      = flag.String("ip", "192.168.1.1", "TPLink HS110 ip address")
	influxURL      = flag.String("influxurl", "", "InfluxDB url - Ex: http://yourserver:8086")
	influxDatabase = flag.String("influxdb", "sensors", "Influx database to log to")
	loglevel       = flag.String("loglevel", "INFO", "Set the console log level")

	//	Variables
	maxPoints = 300 // 5 minute moving average
)

func main() {

	//	Parse the command line for flags:
	flag.Parse()

	//	Gather hostname
	hostname, _ := os.Hostname()

	//	Set the log level from flag
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(*loglevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	//	Keep track of the values over time
	//	-- initialize the slice size to be maxPoints
	tcurrent := make([]float64, maxPoints)
	tvolts := make([]float64, maxPoints)
	tpower := make([]float64, maxPoints)

	//	Emit settings:
	log.Printf("Loglevel: %s", *loglevel)
	log.Printf("[INFO] Using IP address: %s\n", *monitorIP)
	log.Printf("[INFO] Logging from hostname: %s\n", hostname)
	log.Printf("[INFO] Using influx url: %s\n", *influxURL)
	log.Printf("[INFO] Using influx db: %s\n", *influxDatabase)

	//	Spin up a connection to influx
	c, _ := influxdb.NewHTTPClient(influxdb.HTTPConfig{Addr: *influxURL})

	//	Spin up a connection to the tplink device
	plug := tplink.NewHS110(*monitorIP)

	log.Printf("[INFO] Collecting data and logging...\n")
	for range time.Tick(time.Second * 1) {
		meter, err := plug.Meter()
		if err != nil {
			log.Fatalf("[ERROR] failed: %s\n", err)
		}

		//	Track the measurements
		tcurrent = append(tcurrent, meter.Current)
		tvolts = append(tvolts, meter.Voltage)
		tpower = append(tpower, meter.Power)

		//	Calculate standard deviation
		tcurrentmean, err := stats.Mean(tcurrent)
		if err != nil {
			log.Fatal(err)
		}

		tvoltsmean, err := stats.Mean(tvolts)
		if err != nil {
			log.Fatal(err)
		}

		tpowermean, err := stats.Mean(tpower)
		if err != nil {
			log.Fatal(err)
		}

		//	Keep a rolling collection of data...
		//	If we already have maxPoints items
		//	remove the first item:
		if len(tcurrent) > maxPoints {
			tcurrent = tcurrent[1:]
		}

		if len(tvolts) > maxPoints {
			tvolts = tvolts[1:]
		}

		if len(tpower) > maxPoints {
			tpower = tpower[1:]
		}

		if *influxURL != "" {
			// Create a new point batch
			bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
				Database:  *influxDatabase,
				Precision: "s",
			})
			if err != nil {
				log.Fatalf("[ERROR] Creating batch point failed: %s", err)
			}

			// Create a point and add to batch
			tags := map[string]string{
				"host":     hostname,
				"deviceip": *monitorIP,
			}
			fields := map[string]interface{}{
				"current": meter.Current,
				"voltage": meter.Voltage,
				"power":   meter.Power,
				"total":   meter.Total,
				"cmean":   tcurrentmean,
				"vmean":   tvoltsmean,
				"pmean":   tpowermean,
			}

			pt, err := influxdb.NewPoint("tplink-HS110", tags, fields, time.Now())
			if err != nil {
				log.Fatal(err)
			}
			bp.AddPoint(pt)

			// Write the batch
			if err := c.Write(bp); err != nil {
				log.Printf("[WARN] Problem writing to InfluxDB server: %v", err)
			}
		}

		log.Printf("[DEBUG]\n Direct reading: %+v\n Means: current: %v volts: %v power: %v\n\n", meter, tcurrentmean, tvoltsmean, tpowermean)
	}
}
