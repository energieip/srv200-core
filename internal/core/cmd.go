package core

import "encoding/json"

type LedCmd struct {
	Mac      string `json:"mac"`
	Auto     bool   `json:"auto"`
	Setpoint int    `json:"setpoint"`
}

type GroupCmd struct {
	Group       int  `json:"group"`
	Auto        bool `json:"auto"`
	LedSetpoint int  `json:"ledSetPoint"`
}

// ToJSON dump LedCmd struct
func (m LedCmd) ToJSON() (string, error) {
	inrec, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToLedCmd convert map interface to ToLedCmd object
func ToLedCmd(val interface{}) (*LedCmd, error) {
	var m LedCmd
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &m)
	return &m, err
}

// ToJSON dump GroupCmd struct
func (m GroupCmd) ToJSON() (string, error) {
	inrec, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToGroupCmd convert map interface to ToGroupCmd object
func ToGroupCmd(val interface{}) (*GroupCmd, error) {
	var m GroupCmd
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &m)
	return &m, err
}
