package core

import "encoding/json"

type ReplaceDriver struct {
	OldFullMac string `json:"oldFullMac"`
	NewFullMac string `json:"newFullMac"`
}

// ToJSON dump Project struct
func (p ReplaceDriver) ToJSON() (string, error) {
	inrec, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(inrec), err
}

//ToReplaceDriver convert map interface to ReplaceDriver object
func ToReplaceDriver(val interface{}) (*ReplaceDriver, error) {
	var p ReplaceDriver
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &p)
	return &p, err
}
