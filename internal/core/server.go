package core

import (
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	pkg "github.com/energieip/common-components-go/pkg/service"
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
