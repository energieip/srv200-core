package core

import "encoding/json"

//Model represent device object in the building map
type Model struct {
	Name           string `json:"name"`
	DeviceType     string `json:"deviceType"`
	Vendor         string `json:"vendor"`
	URL            string `json:"url"`
	ProductionYear string `json:"productionYear"`
}

// ToJSON dump Model struct
func (m Model) ToJSON() (string, error) {
	inrec, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToModel convert map interface to Model object
func ToModel(val interface{}) (*Model, error) {
	var m Model
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &m)
	return &m, err
}
