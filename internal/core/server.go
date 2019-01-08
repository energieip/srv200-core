package core

import (
	gm "github.com/energieip/common-group-go/pkg/groupmodel"
	dl "github.com/energieip/common-led-go/pkg/driverled"
	ds "github.com/energieip/common-sensor-go/pkg/driversensor"
	pkg "github.com/energieip/common-service-go/pkg/service"
)

//ServerConfig server configuration
type ServerConfig struct {
	Switchs  map[string]SwitchConfig   `json:"switchs"`
	Leds     map[string]dl.LedSetup    `json:"leds"`
	Sensors  map[string]ds.SensorSetup `json:"sensors"`
	Groups   map[int]gm.GroupConfig    `json:"groups"`
	Services map[string]pkg.Service    `json:"services"`
	Models   map[string]Model          `json:"models"`
	Projects map[string]Project        `json:"projects"`
}

//ServerCmd server configuration
type ServerCmd struct {
	Switchs map[string]SwitchCmd     `json:"switchs"`
	Leds    map[string]dl.LedConf    `json:"leds"`
	Sensors map[string]ds.SensorConf `json:"sensors"`
	Groups  map[int]gm.GroupConfig   `json:"groups"`
}
