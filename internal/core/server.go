package core

import (
	"encoding/json"

	"github.com/energieip/common-group-go/pkg/groupmodel"
	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
	pkg "github.com/energieip/common-service-go/pkg/service"
)

//SwitchSetup content
type SwitchSetup struct {
	Mac           string `json:"mac"`
	FriendlyName  string `json:"friendlyName"`
	IP            string `json:"ip"`
	Cluster       int    `json:"cluster"`
	DumpFrequency int    `json:"dumpFrequency"`
	IsConfigured  *bool  `json:"isConfigured"`
}

//SwitchConfig content
type SwitchConfig struct {
	Mac           string `json:"mac"`
	FriendlyName  string `json:"friendlyName"`
	IP            string `json:"ip"`
	Cluster       int    `json:"cluster"`
	IsConfigured  *bool  `json:"isConfigured"`
	DumpFrequency *int   `json:"dumpFrequency"`
}

//SwitchCmd content
type SwitchCmd struct {
	FriendlyName *string `json:"friendlyName"`
	IP           *string `json:"ip"`
	Cluster      *int    `json:"cluster"`
	IsConfigured *bool   `json:"isConfigured"`
}

//ServiceDump content
type ServiceDump struct {
	pkg.ServiceStatus
	SwitchMac string `json:"switchMac"`
}

//ServerConfig server configuration
type ServerConfig struct {
	Switchs  map[string]SwitchSetup              `json:"switchs"`
	Leds     map[string]driverled.LedSetup       `json:"leds"`
	Sensors  map[string]driversensor.SensorSetup `json:"sensors"`
	Groups   map[int]groupmodel.GroupConfig      `json:"groups"`
	Services map[string]pkg.Service              `json:"services"`
	Models   map[string]Model                    `json:"models"`
	Projects map[string]Project                  `json:"projects"`
}

//ServerCmd server configuration
type ServerCmd struct {
	Switchs map[string]SwitchCmd               `json:"switchs"`
	Leds    map[string]driverled.LedConf       `json:"leds"`
	Sensors map[string]driversensor.SensorConf `json:"sensors"`
	Groups  map[int]groupmodel.GroupConfig     `json:"groups"`
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

//ToSwitchSetup convert map interface to SwitchSetup object
func ToSwitchSetup(val interface{}) (*SwitchSetup, error) {
	var sw SwitchSetup
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &sw)
	return &sw, err
}

//ToSwitchConfig convert map interface to SwitchConfig object
func ToSwitchConfig(val interface{}) (*SwitchConfig, error) {
	var sw SwitchConfig
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &sw)
	return &sw, err
}
