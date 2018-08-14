package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/danesparza/tplink"
	"github.com/hashicorp/logutils"
	influxdb "github.com/influxdata/influxdb/client/v2"
)

var (
	//	Flags
	monitorIP      = flag.String("ip", "192.168.1.1", "TPLink HS110 ip address")
	influxURL      = flag.String("influxurl", "", "InfluxDB url - Ex: http://yourserver:8086")
	influxDatabase = flag.String("influxdb", "sensors", "Influx database to log to")
	loglevel       = flag.String("loglevel", "INFO", "Set the console log level")
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

		log.Printf("[DEBUG] Result: %+v\n", meter)
	}
}
