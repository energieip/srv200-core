package core

import (
	gm "github.com/energieip/common-group-go/pkg/groupmodel"
	dl "github.com/energieip/common-led-go/pkg/driverled"
	ds "github.com/energieip/common-sensor-go/pkg/driversensor"
)

type EventLed struct {
	Led   dl.Led `json:"led"`
	Label string `json:"label"`
}

type EventSensor struct {
	Sensor ds.Sensor `json:"sensor"`
	Label  string    `json:"label"`
}

//EventStatus
type EventStatus struct {
	Leds    []EventLed       `json:"leds"`
	Sensors []EventSensor    `json:"sensors"`
	Groups  []gm.GroupStatus `json:"groups"`
}
