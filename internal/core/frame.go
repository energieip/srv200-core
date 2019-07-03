package core

import "encoding/json"

//Frame represent device object in the building map
type Frame struct {
	Label        string `json:"label"`
	FriendlyName string `json:"friendlyName"`
	Cluster      int    `json:"cluster"`
}

// ToJSON dump Model struct
func (m Frame) ToJSON() (string, error) {
	inrec, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(inrec), err
}

//ToFrame convert map interface to Model object
func ToFrame(val interface{}) (*Frame, error) {
	var m Frame
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &m)
	return &m, err
}
