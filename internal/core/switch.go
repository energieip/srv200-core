package core

import (
	"encoding/json"

	sd "github.com/energieip/common-components-go/pkg/dswitch"
)

//SwitchConfig content
type SwitchConfig struct {
	Mac           string `json:"mac"`
	FullMac       string `json:"fullMac"`
	FriendlyName  string `json:"friendlyName"`
	IP            string `json:"ip"`
	Cluster       int    `json:"cluster"`
	IsConfigured  *bool  `json:"isConfigured"`
	DumpFrequency *int   `json:"dumpFrequency"`
	Label         string `json:"label"`
}

//SwitchCmd content
type SwitchCmd struct {
	Mac          string `json:"mac"`
	FullMac      string `json:"fullMac"`
	IsConfigured *bool  `json:"isConfigured"`
}

type SwitchDump struct {
	sd.Switch
	ErrorCode *int `json:"errorCode"`
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
