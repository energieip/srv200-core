package core

import "encoding/json"

//InstallDriver describe the link between the object in the building map and the configuration
type InstallDriver struct {
	Label  string `json:"label"`   //cable label
	Device string `json:"device"`  //device Type LED/HVAC/SENSOR/BLIND/SWITCH/NANOSENSE
	Mac    string `json:"fullMac"` //device Full Mac address
}

// ToJSON dump Project struct
func (p InstallDriver) ToJSON() (string, error) {
	inrec, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(inrec), err
}

//ToInstallDriver convert map interface to InstallDriver object
func ToInstallDriver(val interface{}) (*InstallDriver, error) {
	var p InstallDriver
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &p)
	return &p, err
}

//DriverDesc describe the driver configuration
type DriverDesc struct {
	Device string `json:"device"`  //device Type LED/HVAC/SENSOR/BLIND/SWITCH/NANOSENSE
	Mac    string `json:"fullMac"` //device Full Mac address
}

// ToJSON dump Project struct
func (p DriverDesc) ToJSON() (string, error) {
	inrec, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(inrec), err
}

//ToDriverDesc convert map interface to DriverDesc object
func ToDriverDesc(val interface{}) (*DriverDesc, error) {
	var p DriverDesc
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &p)
	return &p, err
}
