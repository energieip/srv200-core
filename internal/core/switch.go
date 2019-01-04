package core

import (
	"encoding/json"

	sdevice "github.com/energieip/common-switch-go/pkg/deviceswitch"
)

type SwitchDump struct {
	sdevice.Switch
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
