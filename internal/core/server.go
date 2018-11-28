package core

import (
	"encoding/json"

	"github.com/energieip/common-group-go/pkg/groupmodel"
	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
)

//SwitchSetup content
type SwitchSetup struct {
	Mac string `json:"mac"`
}

//ServerConfig server configuration
type ServerConfig struct {
	Switchs map[string]SwitchSetup              `json:"switchs"`
	Leds    map[string]driverled.LedSetup       `json:"leds"`
	Sensors map[string]driversensor.SensorSetup `json:"sensors"`
	Groups  map[int]groupmodel.GroupConfig      `json:"groups"`
}

// ToJSON dump server config struct
func (server ServerConfig) ToJSON() (string, error) {
	inrec, err := json.Marshal(server)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

// ToJSON dump switch struct
func (sw SwitchSetup) ToJSON() (string, error) {
	inrec, err := json.Marshal(sw)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToSwitchSetup convert map interface to Led object
func ToSwitchSetup(val interface{}) (*SwitchSetup, error) {
	var sw SwitchSetup
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &sw)
	return &sw, err
}
