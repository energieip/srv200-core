package core

import "encoding/json"

type Driver struct {
	Mac    string `json:"mac"`
	Label  string `json:"label"`
	Type   string `json:"driverType"`
	Active bool   `json:"active"`
}

// ToJSON dump Driver struct
func (m Driver) ToJSON() (string, error) {
	inrec, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(inrec), err
}

//ToDriver convert map interface to Driver object
func ToDriver(val interface{}) (*Driver, error) {
	var m Driver
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &m)
	return &m, err
}
