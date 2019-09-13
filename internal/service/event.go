package service

import (
	"github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/dhvac"
	dl "github.com/energieip/common-components-go/pkg/dled"
	"github.com/energieip/common-components-go/pkg/dnanosense"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	"github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/common-components-go/pkg/dwago"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	cmap "github.com/orcaman/concurrent-map"
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
	HvacElt   = "hvac"
	WagoElt   = "wago"
	NanoElt   = "nano"
	SwitchElt = "switch"
)

func (s *CoreService) prepareAPIEvent(evtType, evtObj string, event interface{}) {
	bufAPI := cmap.New()
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
		bufAPI.Set(evtType, core.EventStatus{
			Leds:    []core.EventLed{},
			Sensors: []core.EventSensor{},
			Groups:  []gm.GroupStatus{},
			Blinds:  []core.EventBlind{},
			Hvacs:   []core.EventHvac{},
			Wagos:   []core.EventWago{},
			Switchs: []core.EventSwitch{},
			Nanos:   []core.EventNano{},
		})

		value, _ := bufAPI.Get(evtType)
		val, _ := core.ToEventStatus(value)
		val.Sensors = append(val.Sensors, evt)
		bufAPI.Set(evtType, val)

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

		bufAPI.Set(evtType, core.EventStatus{
			Leds:    []core.EventLed{},
			Sensors: []core.EventSensor{},
			Groups:  []gm.GroupStatus{},
			Blinds:  []core.EventBlind{},
			Hvacs:   []core.EventHvac{},
			Wagos:   []core.EventWago{},
			Nanos:   []core.EventNano{},
			Switchs: []core.EventSwitch{},
		})

		value, _ := bufAPI.Get(evtType)
		val, _ := core.ToEventStatus(value)
		val.Leds = append(val.Leds, evt)
		bufAPI.Set(evtType, val)

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

		bufAPI.Set(evtType, core.EventStatus{
			Leds:    []core.EventLed{},
			Sensors: []core.EventSensor{},
			Groups:  []gm.GroupStatus{},
			Blinds:  []core.EventBlind{},
			Hvacs:   []core.EventHvac{},
			Wagos:   []core.EventWago{},
			Nanos:   []core.EventNano{},
			Switchs: []core.EventSwitch{},
		})

		value, _ := bufAPI.Get(evtType)
		val, _ := core.ToEventStatus(value)
		val.Blinds = append(val.Blinds, evt)
		bufAPI.Set(evtType, val)

	case HvacElt:
		hvac, err := dhvac.ToHvac(event)
		if err != nil || hvac == nil {
			return
		}
		label := ""
		project := database.GetProjectByMac(s.db, hvac.Mac)
		if project != nil {
			label = project.Label
		}
		evt := core.EventHvac{
			Hvac:  *hvac,
			Label: label,
		}
		bufAPI.Set(evtType, core.EventStatus{
			Leds:    []core.EventLed{},
			Sensors: []core.EventSensor{},
			Groups:  []gm.GroupStatus{},
			Blinds:  []core.EventBlind{},
			Hvacs:   []core.EventHvac{},
			Wagos:   []core.EventWago{},
			Nanos:   []core.EventNano{},
			Switchs: []core.EventSwitch{},
		})

		value, _ := bufAPI.Get(evtType)
		val, _ := core.ToEventStatus(value)
		val.Hvacs = append(val.Hvacs, evt)
		bufAPI.Set(evtType, val)

	case WagoElt:
		wago, err := dwago.ToWago(event)
		if err != nil || wago == nil {
			return
		}
		label := ""
		project := database.GetProjectByMac(s.db, wago.Mac)
		if project != nil {
			label = project.Label
		}
		evt := core.EventWago{
			Wago:  *wago,
			Label: label,
		}

		bufAPI.Set(evtType, core.EventStatus{
			Leds:    []core.EventLed{},
			Sensors: []core.EventSensor{},
			Groups:  []gm.GroupStatus{},
			Blinds:  []core.EventBlind{},
			Hvacs:   []core.EventHvac{},
			Wagos:   []core.EventWago{},
			Nanos:   []core.EventNano{},
			Switchs: []core.EventSwitch{},
		})
		value, _ := bufAPI.Get(evtType)
		val, _ := core.ToEventStatus(value)
		val.Wagos = append(val.Wagos, evt)
		bufAPI.Set(evtType, val)

	case NanoElt:
		nano, err := dnanosense.ToNanosense(event)
		if err != nil || nano == nil {
			return
		}
		label := ""
		project := database.GetProjectByMac(s.db, nano.Mac)
		if project != nil {
			label = project.Label
		}
		evt := core.EventNano{
			Nano:  *nano,
			Label: label,
		}

		bufAPI.Set(evtType, core.EventStatus{
			Leds:    []core.EventLed{},
			Sensors: []core.EventSensor{},
			Groups:  []gm.GroupStatus{},
			Blinds:  []core.EventBlind{},
			Hvacs:   []core.EventHvac{},
			Wagos:   []core.EventWago{},
			Nanos:   []core.EventNano{},
			Switchs: []core.EventSwitch{},
		})

		value, _ := bufAPI.Get(evtType)
		val, _ := core.ToEventStatus(value)
		val.Nanos = append(val.Nanos, evt)
		bufAPI.Set(evtType, val)

	case GroupElt:
		group, err := gm.ToGroupStatus(event)
		if err != nil || group == nil {
			return
		}
		bufAPI.Set(evtType, core.EventStatus{
			Leds:    []core.EventLed{},
			Sensors: []core.EventSensor{},
			Groups:  []gm.GroupStatus{},
			Blinds:  []core.EventBlind{},
			Hvacs:   []core.EventHvac{},
			Wagos:   []core.EventWago{},
			Nanos:   []core.EventNano{},
			Switchs: []core.EventSwitch{},
		})

		value, _ := bufAPI.Get(evtType)
		val, _ := core.ToEventStatus(value)
		val.Groups = append(val.Groups, *group)
		bufAPI.Set(evtType, val)

	case SwitchElt:
		sw, err := dswitch.ToSwitchStatus(event)
		if err != nil || sw == nil {
			return
		}
		label := ""
		project := database.GetProjectByMac(s.db, sw.Mac)
		if project != nil {
			label = project.Label
		}
		evt := core.EventSwitch{
			Switch: *sw,
			Label:  label,
		}
		bufAPI.Set(evtType, core.EventStatus{
			Leds:    []core.EventLed{},
			Sensors: []core.EventSensor{},
			Groups:  []gm.GroupStatus{},
			Blinds:  []core.EventBlind{},
			Hvacs:   []core.EventHvac{},
			Wagos:   []core.EventWago{},
			Nanos:   []core.EventNano{},
			Switchs: []core.EventSwitch{},
		})

		value, _ := bufAPI.Get(evtType)
		val, _ := core.ToEventStatus(value)
		val.Switchs = append(val.Switchs, evt)
		bufAPI.Set(evtType, val)
	}
	if len(bufAPI) != 0 {
		select {
		case s.eventsAPI <- bufAPI.Items():
			rlog.Debug("API event Sent", bufAPI)
		default:
			rlog.Debug("API event Dropped", bufAPI)
		}
	}
}
