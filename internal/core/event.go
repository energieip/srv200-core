package core

import (
	"encoding/json"

	"github.com/energieip/common-components-go/pkg/dhvac"
	"github.com/energieip/common-components-go/pkg/dnanosense"
	"github.com/energieip/common-components-go/pkg/dwago"

	"github.com/energieip/common-components-go/pkg/dblind"
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

type EventBlind struct {
	Blind dblind.Blind `json:"blind"`
	Label string       `json:"label"`
}

type EventHvac struct {
	Hvac  dhvac.Hvac `json:"hvac"`
	Label string     `json:"label"`
}

type EventWago struct {
	Wago  dwago.Wago `json:"wago"`
	Label string     `json:"label"`
}

type EventNano struct {
	Nano  dnanosense.Nanosense `json:"nanosense"`
	Label string               `json:"label"`
}

//EventStatus
type EventStatus struct {
	Leds    []EventLed       `json:"leds"`
	Blinds  []EventBlind     `json:"blinds"`
	Sensors []EventSensor    `json:"sensors"`
	Hvacs   []EventHvac      `json:"hvacs"`
	Wagos   []EventWago      `json:"wagos"`
	Nanos   []EventNano      `json:"nanosenses"`
	Groups  []gm.GroupStatus `json:"groups"`
}

//ToEventStatus convert map interface to EventStatus object
func ToEventStatus(val interface{}) (*EventStatus, error) {
	var p EventStatus
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &p)
	return &p, err
}
