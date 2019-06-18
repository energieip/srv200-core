package service

import (
	"strings"

	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/common-components-go/pkg/tools"
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
	submac := strings.SplitN(driver.FullMac, ":", 4)
	mac := submac[len(submac)-1]
	proj.Mac = &mac
	proj.FullMac = &driver.FullMac

	//update project
	database.SaveProject(s.db, *proj)

	refModel := *proj.ModelName
	dType := tools.Model2Type(refModel)

	switch dType {
	case "LED":
		elt, _ := database.GetLedLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Led " + proj.Label + " in database")
			return
		}
		elt.FullMac = driver.FullMac
		elt.Mac = mac
		database.SaveLedLabelConfig(s.db, *elt)

		switchConf := sd.SwitchConfig{}
		switchConf.Mac = elt.SwitchMac
		switchConf.LedsSetup[elt.Mac] = *elt
		url := "/write/switch/" + elt.SwitchMac + "/update/settings"
		dump, _ := switchConf.ToJSON()
		s.server.SendCommand(url, dump)
	case "BLIND":
		elt, _ := database.GetBlindLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Blind " + proj.Label + " in database")
			return
		}
		elt.FullMac = driver.FullMac
		elt.Mac = mac
		database.SaveBlindLabelConfig(s.db, *elt)

		// send allow new driver configuration to the switch
		switchConf := sd.SwitchConfig{}
		switchConf.Mac = elt.SwitchMac
		switchConf.BlindsSetup[elt.Mac] = *elt
		url := "/write/switch/" + elt.SwitchMac + "/update/settings"
		dump, _ := switchConf.ToJSON()
		s.server.SendCommand(url, dump)
	case "HVAC":
		elt, _ := database.GetHvacLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Hvac " + proj.Label + " in database")
			return
		}
		elt.FullMac = driver.FullMac
		elt.Mac = mac
		database.SaveHvacLabelConfig(s.db, *elt)

		// send allow new driver configuration to the switch
		switchConf := sd.SwitchConfig{}
		switchConf.Mac = elt.SwitchMac
		switchConf.HvacsSetup[elt.Mac] = *elt
		url := "/write/switch/" + elt.SwitchMac + "/update/settings"
		dump, _ := switchConf.ToJSON()
		s.server.SendCommand(url, dump)
	case "SENSOR":
		elt, _ := database.GetSensorLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Hvac " + proj.Label + " in database")
			return
		}
		elt.FullMac = driver.FullMac
		elt.Mac = mac
		database.SaveSensorLabelConfig(s.db, *elt)

		// send allow new driver configuration to the switch
		switchConf := sd.SwitchConfig{}
		switchConf.Mac = elt.SwitchMac
		switchConf.SensorsSetup[elt.Mac] = *elt
		url := "/write/switch/" + elt.SwitchMac + "/update/settings"
		dump, _ := switchConf.ToJSON()
		s.server.SendCommand(url, dump)
	case "SWITCH":
		elt := database.GetSwitchLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find SWITCH " + proj.Label + " in database")
			return
		}
		elt.FullMac = driver.FullMac
		elt.Mac = mac
		database.SaveSwitchLabelConfig(s.db, *elt)

		switchConf := sd.SwitchConfig{}
		switchConf.Mac = elt.Mac
		switchConf.FullMac = elt.FullMac
		url := "/write/switch/" + elt.Mac + "/update/settings"
		dump, _ := switchConf.ToJSON()
		s.server.SendCommand(url, dump)
	}
}
