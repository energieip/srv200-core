package core

import (
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
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
