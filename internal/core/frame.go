package core

import "encoding/json"

//Frame represent device object in the building map
type Frame struct {
	Label        string `json:"label"`
	FriendlyName string `json:"friendlyName"`
	Cluster      int    `json:"cluster"`
}

type FrameStatus struct {
	Label        string `json:"label"`
	FriendlyName string `json:"friendlyName"`
	Cluster      int    `json:"cluster"`
	StatePuls1   int    `json:"statePuls1"`
	StatePuls2   int    `json:"statePuls2"`
	StatePuls3   int    `json:"statePuls3"`
	StatePuls4   int    `json:"statePuls4"`
	StatePuls5   int    `json:"statePuls5"`
	StateBaes    int    `json:"stateBaes"`
	LedsPower    int64  `json:"ledsPower"`
	BlindsPower  int64  `json:"blindsPower"`
	HvacsPower   int64  `json:"hvacsPower"`
	TotalPower   int64  `json:"totalPower"`
	HvacsEnergy  int64  `json:"hvacsEnergy"`
	LedsEnergy   int64  `json:"ledsEnergy"`
	BlindsEnergy int64  `json:"blindsEnergy"`
	TotalEnergy  int64  `json:"totalEnergy"`
	Profil       string `json:"profil"`
	Error        int    `json:"error"`
}

// ToJSON dump Model struct
func (m Frame) ToJSON() (string, error) {
	inrec, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(inrec), err
}

//ToFrame convert map interface to Model object
func ToFrame(val interface{}) (*Frame, error) {
	var m Frame
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &m)
	return &m, err
}
