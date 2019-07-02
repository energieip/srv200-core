package core

import "encoding/json"

//InstallDriver describe the link between the object in the building map and the configuration
type InstallDriver struct {
	Label   string `json:"label"`   //cable label
	Device  string `json:"device"`  //device Type LED/HVAC/SENSOR/BLIND/SWITCH
	FullMac string `json:"fullMac"` //device Full Mac address
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
