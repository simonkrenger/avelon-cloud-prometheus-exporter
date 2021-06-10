package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/simonkrenger/avelon-cloud-prometheus-exporter/avelon"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	fetchInterval = 60 * time.Minute

	observedDevicesList []avelonDeviceMetrics
)

type avelonDeviceMetrics struct {
	Device *avelon.Device

	SingalStrengthGauge prometheus.Gauge
	BatteryLevelGauge   prometheus.Gauge
	AltitudeGauge       prometheus.Gauge

	TemperatureGauge prometheus.Gauge
	HumidityGauge    prometheus.Gauge
	PressureGauge    prometheus.Gauge
}

func init() {

	log.Printf("init() started")

	d := os.Getenv("DEVICE_LIST")
	if d == "" {
		log.Fatal("FATAL: DEVICE_LIST not specified, aborting...")
		os.Exit(1)
	}

	for _, code := range strings.Split(d, ",") {

		log.Printf("Setting up device %s...", code)
		device := new(avelon.Device)
		device.ActivationCode = code

		deviceMetrics := avelonDeviceMetrics{
			Device: device,
			SingalStrengthGauge: prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        "avelon_device_signal_strength",
				Help:        "Current device signal strength",
				ConstLabels: prometheus.Labels{"activationcode": device.ActivationCode},
			}),
			BatteryLevelGauge: prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        "avelon_device_battery_level_percent",
				Help:        "Current device battery level",
				ConstLabels: prometheus.Labels{"activationcode": device.ActivationCode},
			}),
			AltitudeGauge: prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        "avelon_device_altitude_msl",
				Help:        "Current device altitude",
				ConstLabels: prometheus.Labels{"activationcode": device.ActivationCode},
			}),

			TemperatureGauge: prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        "avelon_record_last_temperature_celsius",
				Help:        "Latest temperature measurement",
				ConstLabels: prometheus.Labels{"activationcode": device.ActivationCode},
			}),
			HumidityGauge: prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        "avelon_record_last_humidity_percent",
				Help:        "Latest air humidity measurement",
				ConstLabels: prometheus.Labels{"activationcode": device.ActivationCode},
			}),
			PressureGauge: prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        "avelon_record_last_pressure_hpa",
				Help:        "Latest air pressure measurement",
				ConstLabels: prometheus.Labels{"activationcode": device.ActivationCode},
			}),
		}
		observedDevicesList = append(observedDevicesList, deviceMetrics)

		prometheus.MustRegister(deviceMetrics.SingalStrengthGauge)
		prometheus.MustRegister(deviceMetrics.BatteryLevelGauge)
		prometheus.MustRegister(deviceMetrics.AltitudeGauge)

		prometheus.MustRegister(deviceMetrics.TemperatureGauge)
		prometheus.MustRegister(deviceMetrics.HumidityGauge)
		prometheus.MustRegister(deviceMetrics.PressureGauge)
	}
}

func main() {

	http.Handle("/metrics", promhttp.Handler())

	go fetchAvelonData()

	log.Printf("Serving metrics at ':8080/metrics'")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func fetchAvelonData() {

	for {

		log.Printf("Fetch started.")

		for _, dev := range observedDevicesList {
			info, err := dev.Device.FetchDeviceInfo()
			if err != nil {
				log.Printf("Fetching device information failed: '%v'", err)
				continue
			}
			dev.BatteryLevelGauge.Set(info.BatteryLevel)
			dev.SingalStrengthGauge.Set(float64(info.SignalStrength))
			dev.AltitudeGauge.Set(info.Altitude)

			records, err := dev.Device.FetchRecords()
			if err != nil {
				log.Printf("Fetching device records failed: '%v'", err)
				continue
			}

			dev.TemperatureGauge.Set(findLatestValueForType(*records, *info, "TEMPERATURE"))
			dev.HumidityGauge.Set(findLatestValueForType(*records, *info, "HUMIDITY"))
			dev.PressureGauge.Set(findLatestValueForType(*records, *info, "ATMOSPHERIC_PRESSURE_AVERAGE"))
		}

		log.Printf("Fetch finished.")
		time.Sleep(fetchInterval)
	}
}

func findLatestValueForType(records []avelon.AvelonRecordsResponse, info avelon.AvelonDeviceResponse, t string) (float64) {

	var latestTime int64
	var latestValue float64

	id := 0
	for _, dataPoints := range info.DataPoints {
		if dataPoints.IotType == t {
			id = dataPoints.ID
		}
	}
	if id == 0 {
		log.Printf("Warning: DatapointID for %s not found, returning 0.", t)
		return 0
	}

	for _, dataPoints := range records {
		if dataPoints.DataPointID == id {
			for _, m := range dataPoints.Values {
				if m.T > latestTime {
					latestTime = m.T
					latestValue = m.V
				}
			}
		}
	}

	return latestValue
}
