package service

import (
	"github.com/energieip/common-components-go/pkg/pconst"

	db "github.com/energieip/common-components-go/pkg/dblind"
	dh "github.com/energieip/common-components-go/pkg/dhvac"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) installDriver(dr interface{}) {
	driver, _ := core.ToInstallDriver(dr)
	if driver == nil {
		rlog.Error("Cannot parse replace driver")
		return
	}

	proj, _ := database.GetProject(s.db, driver.Label)
	if proj == nil {
		rlog.Error("Unknow label " + driver.Label)
		return
	}
	proj.Mac = &driver.Mac

	//update project
	database.SaveProject(s.db, *proj)

	dType := driver.Device
	switch dType {
	case pconst.LED:
		elt, _ := database.GetLedLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Led " + proj.Label + " in database")
			return
		}
		elt.Mac = driver.Mac
		database.SaveLedLabelConfig(s.db, *elt)

		switchConf := sd.SwitchConfig{}
		switchConf.Mac = elt.SwitchMac
		switchConf.LedsSetup = make(map[string]dl.LedSetup)
		switchConf.LedsSetup[elt.Mac] = *elt
		url := "/write/switch/" + elt.SwitchMac + "/update/settings"
		dump, _ := switchConf.ToJSON()
		s.server.SendCommand(url, dump)
	case pconst.BLIND:
		elt, _ := database.GetBlindLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Blind " + proj.Label + " in database")
			return
		}
		elt.Mac = driver.Mac
		database.SaveBlindLabelConfig(s.db, *elt)

		// send allow new driver configuration to the switch
		switchConf := sd.SwitchConfig{}
		switchConf.Mac = elt.SwitchMac
		switchConf.BlindsSetup = make(map[string]db.BlindSetup)
		switchConf.BlindsSetup[elt.Mac] = *elt
		url := "/write/switch/" + elt.SwitchMac + "/update/settings"
		dump, _ := switchConf.ToJSON()
		s.server.SendCommand(url, dump)
	case pconst.HVAC:
		elt, _ := database.GetHvacLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Hvac " + proj.Label + " in database")
			return
		}
		elt.Mac = driver.Mac
		database.SaveHvacLabelConfig(s.db, *elt)

		// send allow new driver configuration to the switch
		switchConf := sd.SwitchConfig{}
		switchConf.Mac = elt.SwitchMac
		switchConf.HvacsSetup = make(map[string]dh.HvacSetup)
		switchConf.HvacsSetup[elt.Mac] = *elt
		url := "/write/switch/" + elt.SwitchMac + "/update/settings"
		dump, _ := switchConf.ToJSON()
		s.server.SendCommand(url, dump)
	case pconst.SENSOR:
		elt, _ := database.GetSensorLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Sensor " + proj.Label + " in database")
			return
		}
		elt.Mac = driver.Mac
		database.SaveSensorLabelConfig(s.db, *elt)

		// send allow new driver configuration to the switch
		switchConf := sd.SwitchConfig{}
		switchConf.Mac = elt.SwitchMac
		switchConf.SensorsSetup = make(map[string]ds.SensorSetup)
		switchConf.SensorsSetup[elt.Mac] = *elt
		url := "/write/switch/" + elt.SwitchMac + "/update/settings"
		dump, _ := switchConf.ToJSON()
		s.server.SendCommand(url, dump)
	case pconst.SWITCH:
		elt := database.GetSwitchLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find SWITCH " + proj.Label + " in database")
			return
		}
		elt.Mac = &driver.Mac
		database.SaveSwitchLabelConfig(s.db, *elt)

		switchConf := sd.SwitchConfig{}
		switchConf.Mac = *elt.Mac
		url := "/write/switch/" + *elt.Mac + "/update/settings"
		dump, _ := switchConf.ToJSON()
		s.server.SendCommand(url, dump)
	}
}
