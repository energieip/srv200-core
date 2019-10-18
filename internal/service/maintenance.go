package service

import (
	"strconv"

	"github.com/energieip/common-components-go/pkg/dnanosense"
	"github.com/energieip/common-components-go/pkg/dwago"
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

	project := database.GetProjectByMac(s.db, replace.OldFullMac)
	if project == nil {
		rlog.Error("Unkown old driver")
		return
	}

	project.Mac = &replace.NewFullMac

	//update driver tables
	if project.ModelName != nil {
		refModel := *project.ModelName
		dType := tools.Model2Type(refModel)
		switch dType {
		case pconst.LED:
			oldDriver, _ := database.GetLedConfig(s.db, replace.OldFullMac)
			if oldDriver == nil {
				rlog.Error("Cannot find Led " + replace.OldFullMac + " in database")
				return
			}

			err := database.SwitchLedConfig(s.db, replace.OldFullMac, *project.Mac)
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
				if led != replace.OldFullMac {
					newLeds = append(newLeds, led)
				}
			}

			firstDay := []string{}
			for _, led := range groupCfg.FirstDay {
				if led != replace.OldFullMac {
					firstDay = append(firstDay, led)
				} else {
					firstDay = append(firstDay, replace.NewFullMac)
				}
			}
			newLeds = append(newLeds, replace.NewFullMac)
			groupCfg.Leds = newLeds
			groupCfg.FirstDay = firstDay

			database.UpdateGroupConfig(s.db, *groupCfg)
			newSwitch := database.GetGroupSwitchs(s.db, groupCfg.Group)
			for mac := range newSwitch {
				if mac == "" {
					continue
				}
				sw, _ := database.GetSwitchConfig(s.db, mac)
				if sw != nil {
					ip := "0"
					if sw.IP != nil {
						ip = *sw.IP
					}
					dumpFreq := 1000
					if sw.DumpFrequency != nil {
						dumpFreq = *sw.DumpFrequency
					}
					url := "/write/switch/" + mac + "/update/settings"
					switchSetup := sd.SwitchConfig{}
					switchSetup.Mac = mac
					switchSetup.IP = ip
					switchSetup.DumpFrequency = dumpFreq
					switchSetup.Groups = make(map[int]gm.GroupConfig)
					switchSetup.Groups[groupCfg.Group] = *groupCfg
					dump, _ := switchSetup.ToJSON()
					s.server.SendCommand(url, dump)
				}
			}

		case pconst.BLIND:
			oldDriver, _ := database.GetBlindConfig(s.db, replace.OldFullMac)
			if oldDriver == nil {
				rlog.Error("Cannot find Blind " + replace.OldFullMac + " in database")
				return
			}

			err := database.SwitchBlindConfig(s.db, replace.OldFullMac, *project.Mac)
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
				if blind != replace.OldFullMac {
					newBlinds = append(newBlinds, blind)
				}
			}
			newBlinds = append(newBlinds, replace.NewFullMac)
			groupCfg.Blinds = newBlinds

			database.UpdateGroupConfig(s.db, *groupCfg)
			newSwitch := database.GetGroupSwitchs(s.db, groupCfg.Group)
			for mac := range newSwitch {
				sw, _ := database.GetSwitchConfig(s.db, mac)
				if sw != nil {
					ip := "0"
					if sw.IP != nil {
						ip = *sw.IP
					}
					dumpFreq := 1000
					if sw.DumpFrequency != nil {
						dumpFreq = *sw.DumpFrequency
					}
					url := "/write/switch/" + mac + "/update/settings"
					switchSetup := sd.SwitchConfig{}
					switchSetup.Mac = mac
					switchSetup.IP = ip
					switchSetup.DumpFrequency = dumpFreq
					switchSetup.Groups = make(map[int]gm.GroupConfig)
					switchSetup.Groups[groupCfg.Group] = *groupCfg
					dump, _ := switchSetup.ToJSON()
					s.server.SendCommand(url, dump)
				}
			}
		case pconst.HVAC:
			oldDriver, _ := database.GetHvacConfig(s.db, replace.OldFullMac)
			if oldDriver == nil {
				rlog.Error("Cannot find Hvac " + replace.OldFullMac + " in database")
				return
			}

			err := database.SwitchHvacConfig(s.db, replace.OldFullMac, *project.Mac)
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
				if hvac != replace.OldFullMac {
					newHvacs = append(newHvacs, hvac)
				}
			}
			newHvacs = append(newHvacs, replace.NewFullMac)
			groupCfg.Hvacs = newHvacs

			database.UpdateGroupConfig(s.db, *groupCfg)
			newSwitch := database.GetGroupSwitchs(s.db, groupCfg.Group)
			for mac := range newSwitch {
				if mac == "" {
					continue
				}
				sw, _ := database.GetSwitchConfig(s.db, mac)
				if sw != nil {
					ip := "0"
					if sw.IP != nil {
						ip = *sw.IP
					}
					dumpFreq := 1000
					if sw.DumpFrequency != nil {
						dumpFreq = *sw.DumpFrequency
					}
					url := "/write/switch/" + mac + "/update/settings"
					switchSetup := sd.SwitchConfig{}
					switchSetup.Mac = mac
					switchSetup.IP = ip
					switchSetup.DumpFrequency = dumpFreq
					switchSetup.Groups = make(map[int]gm.GroupConfig)
					switchSetup.Groups[groupCfg.Group] = *groupCfg
					dump, _ := switchSetup.ToJSON()
					s.server.SendCommand(url, dump)
				}
			}
		case pconst.SENSOR:
			oldDriver, _ := database.GetSensorConfig(s.db, replace.OldFullMac)
			if oldDriver == nil {
				rlog.Error("Cannot find Sensor " + replace.OldFullMac + " in database")
				return
			}

			err := database.SwitchSensorConfig(s.db, replace.OldFullMac, *project.Mac)
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
				if sensor != replace.OldFullMac {
					newSensors = append(newSensors, sensor)
				}
			}
			newSensors = append(newSensors, replace.NewFullMac)
			groupCfg.Sensors = newSensors

			database.UpdateGroupConfig(s.db, *groupCfg)
			newSwitch := database.GetGroupSwitchs(s.db, groupCfg.Group)
			for mac := range newSwitch {
				if mac == "" {
					continue
				}
				sw, _ := database.GetSwitchConfig(s.db, mac)
				if sw != nil {
					ip := "0"
					if sw.IP != nil {
						ip = *sw.IP
					}
					dumpFreq := 1000
					if sw.DumpFrequency != nil {
						dumpFreq = *sw.DumpFrequency
					}
					url := "/write/switch/" + mac + "/update/settings"
					switchSetup := sd.SwitchConfig{}
					switchSetup.Mac = mac
					switchSetup.IP = ip
					switchSetup.DumpFrequency = dumpFreq
					switchSetup.Groups = make(map[int]gm.GroupConfig)
					switchSetup.Groups[groupCfg.Group] = *groupCfg
					dump, _ := switchSetup.ToJSON()
					s.server.SendCommand(url, dump)
				}
			}
		case pconst.WAGO:
			oldDriver, _ := database.GetWagoConfig(s.db, replace.OldFullMac)
			if oldDriver == nil {
				rlog.Error("Cannot find Wago " + replace.OldFullMac + " in database")
				return
			}

			err := database.SwitchWagoConfig(s.db, replace.OldFullMac, *project.Mac)
			if err != nil {
				rlog.Error("Cannot update Wago database", err)
				return
			}

			newDriver, _ := database.GetWagoConfig(s.db, *project.Mac)
			if newDriver == nil {
				rlog.Error("Cannot find Wago " + *project.Mac + " in database")
				return
			}

			var oldNanos []dnanosense.NanosenseSetup
			nanos := database.GetNanoSwitchSetup(s.db, newDriver.Cluster)
			for _, nano := range nanos {
				oldNanos = append(oldNanos, nano)
				projectNano := database.GetProjectByMac(s.db, nano.Mac)
				nano.Mac = *project.Mac + "." + strconv.Itoa(nano.ModbusID)
				database.SaveNanoLabelConfig(s.db, nano)
				if projectNano != nil {
					projectNano.Mac = &nano.Mac
					database.SaveProject(s.db, *projectNano)
				}
			}

			for sw := range database.GetCluster(s.db, oldDriver.Cluster) {
				if sw == "" {
					continue
				}
				switchSetup := sd.SwitchConfig{}
				switchSetup.Mac = sw
				switchSetup.WagosSetup = make(map[string]dwago.WagoSetup)
				switchSetup.WagosSetup[oldDriver.Mac] = *oldDriver
				switchSetup.NanosSetup = make(map[string]dnanosense.NanosenseSetup)
				for _, nano := range oldNanos {
					switchSetup.NanosSetup[nano.Mac] = nano
				}
				s.sendSwitchRemoveConfig(switchSetup)
			}
			for mac := range database.GetCluster(s.db, newDriver.Cluster) {
				if mac == "" {
					continue
				}
				sw, _ := database.GetSwitchConfig(s.db, mac)
				if sw != nil {
					ip := "0"
					if sw.IP != nil {
						ip = *sw.IP
					}
					dumpFreq := 1000
					if sw.DumpFrequency != nil {
						dumpFreq = *sw.DumpFrequency
					}
					switchSetupNew := sd.SwitchConfig{}
					switchSetupNew.Mac = mac
					switchSetupNew.IP = ip
					switchSetupNew.DumpFrequency = dumpFreq
					switchSetupNew.WagosSetup = make(map[string]dwago.WagoSetup)
					switchSetupNew.WagosSetup[*project.Mac] = *newDriver
					switchSetupNew.NanosSetup = make(map[string]dnanosense.NanosenseSetup)
					for _, nano := range nanos {
						switchSetupNew.NanosSetup[nano.Mac] = nano
					}
					url := "/write/switch/" + mac + "/update/settings"
					dump, _ := switchSetupNew.ToJSON()
					s.server.SendCommand(url, dump)
				}
			}

		case pconst.SWITCH:
			device, _ := database.GetSwitchConfig(s.db, replace.OldFullMac)
			if device == nil {
				rlog.Error("Cannot find Switch " + replace.OldFullMac + " in database")
				return
			}

			err := database.ReplaceSwitchConfig(s.db, replace.OldFullMac, *project.Mac)
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
