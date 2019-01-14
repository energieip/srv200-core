package service

import (
	"time"

	"github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

const (
	EventRemove = "remove"
	EventUpdate = "update"
	EventAdd    = "add"

	LedElt    = "led"
	SensorElt = "sensor"
	BlindElt  = "blind"
	GroupElt  = "group"
)

func (s *CoreService) prepareAPIEvent(evtType, evtObj string, event interface{}) {
	switch evtObj {
	case SensorElt:
		sensor, err := ds.ToSensor(event)
		if err != nil && sensor == nil {
			return
		}

		label := ""
		project := database.GetProjectByMac(s.db, sensor.Mac)
		if project != nil {
			label = project.Label
		}
		evt := core.EventSensor{
			Sensor: *sensor,
			Label:  label,
		}
		_, ok := s.bufAPI[evtType]
		if !ok {
			s.bufAPI[evtType] = core.EventStatus{
				Leds:    []core.EventLed{},
				Sensors: []core.EventSensor{},
				Groups:  []gm.GroupStatus{},
				Blinds:  []core.EventBlind{},
			}
		}
		val, ok := s.bufAPI[evtType]
		val.Sensors = append(val.Sensors, evt)
		s.bufAPI[evtType] = val

	case LedElt:
		led, err := dl.ToLed(event)
		if err != nil || led == nil {
			return
		}
		label := ""
		project := database.GetProjectByMac(s.db, led.Mac)
		if project != nil {
			label = project.Label
		}
		evt := core.EventLed{
			Led:   *led,
			Label: label,
		}

		_, ok := s.bufAPI[evtType]
		if !ok {
			s.bufAPI[evtType] = core.EventStatus{
				Leds:    []core.EventLed{},
				Sensors: []core.EventSensor{},
				Groups:  []gm.GroupStatus{},
				Blinds:  []core.EventBlind{},
			}
		}
		val, ok := s.bufAPI[evtType]
		val.Leds = append(val.Leds, evt)
		s.bufAPI[evtType] = val

	case BlindElt:
		blind, err := dblind.ToBlind(event)
		if err != nil || blind == nil {
			return
		}
		label := ""
		project := database.GetProjectByMac(s.db, blind.Mac)
		if project != nil {
			label = project.Label
		}
		evt := core.EventBlind{
			Blind: *blind,
			Label: label,
		}

		_, ok := s.bufAPI[evtType]
		if !ok {
			s.bufAPI[evtType] = core.EventStatus{
				Leds:    []core.EventLed{},
				Sensors: []core.EventSensor{},
				Groups:  []gm.GroupStatus{},
				Blinds:  []core.EventBlind{},
			}
		}
		val, ok := s.bufAPI[evtType]
		val.Blinds = append(val.Blinds, evt)
		s.bufAPI[evtType] = val

	case GroupElt:
		group, err := gm.ToGroupStatus(event)
		if err != nil || group == nil {
			return
		}
		_, ok := s.bufAPI[evtType]
		if !ok {
			s.bufAPI[evtType] = core.EventStatus{
				Leds:    []core.EventLed{},
				Sensors: []core.EventSensor{},
				Groups:  []gm.GroupStatus{},
			}
		}
		val, ok := s.bufAPI[evtType]
		val.Groups = append(val.Groups, *group)
		s.bufAPI[evtType] = val
	}
}

func (s *CoreService) pushAPIEvent() {
	timerDump := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timerDump.C:
			if len(s.bufAPI) != 0 {
				select {
				case s.eventsAPI <- s.bufAPI:
					rlog.Debug("API event Sent", s.bufAPI)
				default:
					rlog.Debug("API event Dropped", s.bufAPI)
				}
			}
			s.bufAPI = make(map[string]core.EventStatus)
		}
	}
}
