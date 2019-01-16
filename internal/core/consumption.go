package core

//EventConsumption
type EventConsumption struct {
	Leds   int    `json:"leds"`
	Blinds int    `json:"blinds"`
	Hvac   int    `json:"hvacs"`
	Date   string `json:"date"`
}
