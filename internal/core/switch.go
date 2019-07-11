package core

import (
	"encoding/json"

	sd "github.com/energieip/common-components-go/pkg/dswitch"
)

//SwitchConfig content
type SwitchConfig struct {
	Mac           *string `json:"mac,omitempty"`
	FullMac       *string `json:"fullMac,omitempty"`
	FriendlyName  *string `json:"friendlyName,omitempty"`
	IP            *string `json:"ip,omitempty"`
	Cluster       *int    `json:"cluster,omitempty"`
	Profil        string  `json:"profil"` // I/O card configuration none/pulse
	IsConfigured  *bool   `json:"isConfigured,omitempty"`
	DumpFrequency *int    `json:"dumpFrequency,omitempty"`
	Label         *string `json:"label,omitempty"`
}

//SwitchCmd content
type SwitchCmd struct {
	Mac          string `json:"mac"`
	FullMac      string `json:"fullMac"`
	IsConfigured *bool  `json:"isConfigured,omitempty"`
}

type SwitchDump struct {
	sd.Switch
	StatePuls1   int   `json:"statePuls1"`
	StatePuls2   int   `json:"statePuls2"`
	StatePuls3   int   `json:"statePuls3"`
	StatePuls4   int   `json:"statePuls4"`
	StatePuls5   int   `json:"statePuls5"`
	StateBaes    int   `json:"stateBaes"`
	LedsPower    int64 `json:"ledsPower"`
	BlindsPower  int64 `json:"blindsPower"`
	HvacsPower   int64 `json:"hvacsPower"`
	TotalPower   int64 `json:"totalPower"`
	HvacsEnergy  int64 `json:"hvacsEnergy"`
	LedsEnergy   int64 `json:"ledsEnergy"`
	BlindsEnergy int64 `json:"blindsEnergy"`
	TotalEnergy  int64 `json:"totalEnergy"`
	ErrorCode    *int  `json:"errorCode"`
}

//ToSwitchDump convert map interface to SwitchDump object
func ToSwitchDump(val interface{}) (*SwitchDump, error) {
	var sw SwitchDump
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
