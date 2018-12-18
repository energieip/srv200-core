package core

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
