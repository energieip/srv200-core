package core

import (
	"encoding/json"

	db "github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/dhvac"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
)

//MapInfo
type MapInfo struct {
	Groups  map[string]gm.GroupConfig  `json:"groups"`
	Leds    map[string]dl.LedSetup     `json:"leds"`
	Sensors map[string]ds.SensorSetup  `json:"sensors"`
	Hvacs   map[string]dhvac.HvacSetup `json:"hvacs"`
	Blinds  map[string]db.BlindSetup   `json:"blinds"`
	Models  map[string]Model           `json:"models"`
	Project map[string]Project         `json:"projects"`
	Switchs map[string]SwitchConfig    `json:"switchs"`
}

// ToJSON dump MapInfo struct
func (p MapInfo) ToJSON() (string, error) {
	inrec, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToMapInfo convert map interface to Project object
func ToMapInfo(val interface{}) (*MapInfo, error) {
	var p MapInfo
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &p)
	return &p, err
}
