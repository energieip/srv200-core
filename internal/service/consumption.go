package service

import (
	"time"

	"github.com/energieip/srv200-coreservice-go/internal/core"
	cmap "github.com/orcaman/concurrent-map"

	"github.com/romana/rlog"
)

func (s *CoreService) prepareAPIConsumption(evtObj string, power int) {
	old, _ := s.bufConsumption.Get(evtObj)
	new := old.(int) + power
	s.bufConsumption.Set(evtObj, new)
}

func (s *CoreService) pushConsumptionEvent() {
	s.bufConsumption.Set(LedElt, 0)
	s.bufConsumption.Set(BlindElt, 0)
	s.bufConsumption.Set(HvacElt, 0)
	timerDump := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timerDump.C:
			conso := core.EventConsumption{}
			led, _ := s.bufConsumption.Get(LedElt)
			conso.Leds = led.(int)
			blind, _ := s.bufConsumption.Get(BlindElt)
			conso.Blinds = blind.(int)
			hvac, _ := s.bufConsumption.Get(HvacElt)
			conso.Hvac = hvac.(int)
			conso.Date = time.Now().Format(time.RFC3339)
			select {
			case s.eventsConsumptionAPI <- conso:
				rlog.Debug("Consumption API event Sent", s.bufConsumption)
			default:
				rlog.Debug("Consumption event Dropped", s.bufConsumption)
			}
			s.bufConsumption = cmap.New()
			s.bufConsumption.Set(LedElt, 0)
			s.bufConsumption.Set(BlindElt, 0)
			s.bufConsumption.Set(HvacElt, 0)
		}
	}
}
