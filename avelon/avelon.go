package avelon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	avelonurl string = "https://iot.avelon.cloud/api/v1/360/no-auth/devices/"
)

// AvelonDeviceDatapoint is the description of available datapoints within an AvelonDeviceResponse
type AvelonDeviceDatapoint struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
	Unit     int    `json:"unit"`
	IotType  string `json:"iotType"`
}

// AvelonDeviceResponse is the response from the /api/v1/360/no-auth/devices/ endpoint
type AvelonDeviceResponse struct {
	ID               int                     `json:"id"`
	Name             string                  `json:"name"`
	Serialnumber     string                  `json:"serialnumber"`
	ConnectionState  string                  `json:"connectionState"`
	RegistrationDate int64                   `json:"registrationDate"` // Unix Timestamp
	LastConnection   int64                   `json:"lastConnection"`   // Unix Timestamp
	BatteryLevel     float64                 `json:"batteryLevel"`
	Altitude         float64                 `json:"altitude"`
	AltitudeMode     int                     `json:"altitudeMode"`
	SignalStrength   int                     `json:"signalStrength"`
	ClientID         int                     `json:"clientId"`
	DeviceType       string                  `json:"deviceType"`
	Activated        bool                    `json:"activated"`
	ActivationCode   string                  `json:"activationCode"`
	DataPoints       []AvelonDeviceDatapoint `json:"dataPoints"`
	SelfManaged      bool                    `json:"selfManaged"`
}

// AvelonRecordsResponse is the response from the /api/v1/360/no-auth/devices/<id>/records endpoint
type AvelonRecordsResponse struct {
	DataPointID int `json:"dataPointId"`
	Values      []struct {
		T int64   `json:"t"`
		V float64 `json:"v"`
	} `json:"values"`
}

// Device represents a device
type Device struct {
	ActivationCode string
}

// FetchDeviceInfo queries the device endpoint
func (d *Device) FetchDeviceInfo() (*AvelonDeviceResponse, error) {
	if d.ActivationCode == "" {
		return nil, fmt.Errorf("ActivationCode not set")
	}

	deviceURL := avelonurl + d.ActivationCode
	resp, err := http.Get(deviceURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to GET " + deviceURL + ": " + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read body for " + deviceURL + ": " + err.Error())
	}

	var dev AvelonDeviceResponse
	err = json.Unmarshal(body, &dev)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal failed for response '" + string(body) + "': " + err.Error())
	}
	return &dev, nil
}

// FetchRecords queries the records endpoint
func (d *Device) FetchRecords() (*[]AvelonRecordsResponse, error) {
	if d.ActivationCode == "" {
		return nil, fmt.Errorf("ActivationCode not set")
	}

	recordsURL := avelonurl + d.ActivationCode + "/records"
	resp, err := http.Get(recordsURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to GET for " + recordsURL + ": " + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read body for " + recordsURL + ": " + err.Error())
	}

	var rec []AvelonRecordsResponse
	err = json.Unmarshal(body, &rec)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal failed for response '" + string(body) + "': " + err.Error())
	}
	return &rec, nil
}
