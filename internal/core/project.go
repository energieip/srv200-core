package core

import "encoding/json"

//Project describe the link between the object in the building map and the configuration
type Project struct {
	Label     string  `json:"label"` //cable label
	ModelName *string `json:"modelName,omitempty"`
	Mac       *string `json:"mac,omitempty"`      //device Mac address (short version)
	FullMac   *string `json:"fullMac,omitempty"`  //device Full Mac address
	ModbusID  *int    `json:"modbusID,omitempty"` //modbusID
}

// ToJSON dump Project struct
func (p Project) ToJSON() (string, error) {
	inrec, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToProject convert map interface to Project object
func ToProject(val interface{}) (*Project, error) {
	var p Project
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &p)
	return &p, err
}
