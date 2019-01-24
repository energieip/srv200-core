package service

import (
	"time"

	"github.com/energieip/srv200-coreservice-go/internal/core"

	"github.com/romana/rlog"
)

func (s *CoreService) prepareAPIConsumption(evtObj string, power int) {
	s.bufConsumption[evtObj] += power
}

func (s *CoreService) pushConsumptionEvent() {
	timerDump := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timerDump.C:
			conso := core.EventConsumption{}
			led, _ := s.bufConsumption[LedElt]
			conso.Leds = led
			blind, _ := s.bufConsumption[BlindElt]
			conso.Blinds = blind
			conso.Date = time.Now().Format(time.RFC3339)
			select {
			case s.eventsConsumptionAPI <- conso:
				rlog.Debug("Consumption API event Sent", s.bufConsumption)
			default:
				rlog.Debug("Consumption event Dropped", s.bufConsumption)
			}

			s.bufConsumption = make(map[string]int)
			s.bufConsumption[LedElt] = 0
			s.bufConsumption[BlindElt] = 0
		}
	}
}
