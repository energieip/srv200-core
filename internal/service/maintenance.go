package service

import (
	"strings"

	"github.com/energieip/common-components-go/pkg/pconst"

	db "github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	dh "github.com/energieip/common-components-go/pkg/dhvac"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/common-components-go/pkg/tools"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) replaceDriver(driver interface{}) {
	replace, _ := core.ToReplaceDriver(driver)
	if replace == nil {
		rlog.Error("Cannot parse replace driver")
		return
	}

	oldSubmac := strings.SplitN(replace.OldFullMac, ":", 4)
	oldMac := oldSubmac[len(oldSubmac)-1]
	newSubmac := strings.SplitN(replace.NewFullMac, ":", 4)
	newMac := newSubmac[len(newSubmac)-1]

	project := database.GetProjectByFullMac(s.db, replace.OldFullMac)
	if project == nil {
		project = database.GetProjectByMac(s.db, oldMac)
	}

	if project == nil {
		rlog.Error("Unkown old driver")
		return
	}

	project.FullMac = &replace.NewFullMac
	project.Mac = &newMac

	//update driver tables
	if project.ModelName != nil {
		refModel := *project.ModelName
		dType := tools.Model2Type(refModel)
		switch dType {
		case pconst.LED:
			oldDriver, _ := database.GetLedConfig(s.db, oldMac)
			if oldDriver == nil {
				rlog.Error("Cannot find Led " + oldMac + " in database")
				return
			}

			err := database.SwitchLedConfig(s.db, oldMac, replace.OldFullMac, *project.Mac, *project.FullMac)
			if err != nil {
				rlog.Error("Cannot update Led database", err)
				return
			}

			//send remove reset old driver configuration to the switch
			switchConf := sd.SwitchConfig{}
			switchConf.Mac = oldDriver.SwitchMac
			switchConf.LedsSetup = make(map[string]dl.LedSetup)
			switchConf.LedsSetup[oldDriver.Mac] = *oldDriver
			s.sendSwitchRemoveConfig(switchConf)

			// update group configuration
			// send update to all switch where this group is running
			groupCfg, _ := database.GetGroupConfig(s.db, *oldDriver.Group)
			newLeds := []string{}
			for _, led := range groupCfg.Leds {
				if led != oldMac {
					newLeds = append(newLeds, led)
				}
			}
			groupCfg.Leds = newLeds

			database.UpdateGroupConfig(s.db, *groupCfg)
			newSwitch := database.GetGroupSwitchs(s.db, groupCfg.Group)
			for sw := range newSwitch {
				url := "/write/switch/" + sw + "/update/settings"
				switchSetup := sd.SwitchConfig{}
				switchSetup.Mac = sw
				switchSetup.Groups = make(map[int]gm.GroupConfig)
				switchSetup.Groups[groupCfg.Group] = *groupCfg
				dump, _ := switchSetup.ToJSON()
				s.server.SendCommand(url, dump)
			}

		case pconst.BLIND:
			oldDriver, _ := database.GetBlindConfig(s.db, oldMac)
			if oldDriver == nil {
				rlog.Error("Cannot find Blind " + oldMac + " in database")
				return
			}

			err := database.SwitchBlindConfig(s.db, oldMac, replace.OldFullMac, *project.Mac, *project.FullMac)
			if err != nil {
				rlog.Error("Cannot update Blind database", err)
				return
			}

			//send remove reset old driver configuration to the switch
			switchConf := sd.SwitchConfig{}
			switchConf.Mac = oldDriver.SwitchMac
			switchConf.BlindsSetup = make(map[string]db.BlindSetup)
			switchConf.BlindsSetup[oldDriver.Mac] = *oldDriver
			s.sendSwitchRemoveConfig(switchConf)

			// update group configuration
			// send update to all switch where this group is running
			groupCfg, _ := database.GetGroupConfig(s.db, *oldDriver.Group)
			newBlinds := []string{}
			for _, blind := range groupCfg.Blinds {
				if blind != oldMac {
					newBlinds = append(newBlinds, blind)
				}
			}
			groupCfg.Blinds = newBlinds

			database.UpdateGroupConfig(s.db, *groupCfg)
			newSwitch := database.GetGroupSwitchs(s.db, groupCfg.Group)
			for sw := range newSwitch {
				url := "/write/switch/" + sw + "/update/settings"
				switchSetup := sd.SwitchConfig{}
				switchSetup.Mac = sw
				switchSetup.Groups = make(map[int]gm.GroupConfig)
				switchSetup.Groups[groupCfg.Group] = *groupCfg
				dump, _ := switchSetup.ToJSON()
				s.server.SendCommand(url, dump)
			}
		case pconst.HVAC:
			oldDriver, _ := database.GetHvacConfig(s.db, oldMac)
			if oldDriver == nil {
				rlog.Error("Cannot find Hvac " + oldMac + " in database")
				return
			}

			err := database.SwitchHvacConfig(s.db, oldMac, replace.OldFullMac, *project.Mac, *project.FullMac)
			if err != nil {
				rlog.Error("Cannot update Hvac database", err)
				return
			}

			//send remove reset old driver configuration to the switch
			switchConf := sd.SwitchConfig{}
			switchConf.Mac = oldDriver.SwitchMac
			switchConf.HvacsSetup = make(map[string]dh.HvacSetup)
			switchConf.HvacsSetup[oldDriver.Mac] = *oldDriver
			s.sendSwitchRemoveConfig(switchConf)

			// update group configuration
			// send update to all switch where this group is running
			groupCfg, _ := database.GetGroupConfig(s.db, *oldDriver.Group)
			newHvacs := []string{}
			for _, hvac := range groupCfg.Hvacs {
				if hvac != oldMac {
					newHvacs = append(newHvacs, hvac)
				}
			}
			groupCfg.Hvacs = newHvacs

			database.UpdateGroupConfig(s.db, *groupCfg)
			newSwitch := database.GetGroupSwitchs(s.db, groupCfg.Group)
			for sw := range newSwitch {
				url := "/write/switch/" + sw + "/update/settings"
				switchSetup := sd.SwitchConfig{}
				switchSetup.Mac = sw
				switchSetup.Groups = make(map[int]gm.GroupConfig)
				switchSetup.Groups[groupCfg.Group] = *groupCfg
				dump, _ := switchSetup.ToJSON()
				s.server.SendCommand(url, dump)
			}
		case pconst.SENSOR:
			oldDriver, _ := database.GetSensorConfig(s.db, oldMac)
			if oldDriver == nil {
				rlog.Error("Cannot find Sensor " + oldMac + " in database")
				return
			}

			err := database.SwitchSensorConfig(s.db, oldMac, replace.OldFullMac, *project.Mac, *project.FullMac)
			if err != nil {
				rlog.Error("Cannot update Sensor database", err)
				return
			}

			//send remove reset old driver configuration to the switch
			switchConf := sd.SwitchConfig{}
			switchConf.Mac = oldDriver.SwitchMac
			switchConf.SensorsSetup = make(map[string]ds.SensorSetup)
			switchConf.SensorsSetup[oldDriver.Mac] = *oldDriver
			s.sendSwitchRemoveConfig(switchConf)

			// update group configuration
			// send update to all switch where this group is running
			groupCfg, _ := database.GetGroupConfig(s.db, *oldDriver.Group)
			newSensors := []string{}
			for _, sensor := range groupCfg.Sensors {
				if sensor != oldMac {
					newSensors = append(newSensors, sensor)
				}
			}
			groupCfg.Sensors = newSensors

			database.UpdateGroupConfig(s.db, *groupCfg)
			newSwitch := database.GetGroupSwitchs(s.db, groupCfg.Group)
			for sw := range newSwitch {
				url := "/write/switch/" + sw + "/update/settings"
				switchSetup := sd.SwitchConfig{}
				switchSetup.Mac = sw
				switchSetup.Groups = make(map[int]gm.GroupConfig)
				switchSetup.Groups[groupCfg.Group] = *groupCfg
				dump, _ := switchSetup.ToJSON()
				s.server.SendCommand(url, dump)
			}
		case pconst.SWITCH:
			device, _ := database.GetSwitchConfig(s.db, oldMac)
			if device == nil {
				rlog.Error("Cannot find Switch " + oldMac + " in database")
				return
			}

			err := database.ReplaceSwitchConfig(s.db, oldMac, replace.OldFullMac, *project.Mac, *project.FullMac)
			if err != nil {
				rlog.Error("Cannot update Switch database", err)
				return
			}
		}
	}

	//update project
	err := database.SaveProject(s.db, *project)
	if err != nil {
		rlog.Error("Cannot saved new project configuration")
		return
	}
}
