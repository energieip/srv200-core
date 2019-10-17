package service

import (
	"strconv"
	"time"

	gm "github.com/energieip/common-components-go/pkg/dgroup"
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
	if proj.Mac != nil {
		if *proj.Mac == driver.Mac {
			rlog.Info("Driver " + driver.Mac + " already associate to " + driver.Label)
			return
		}
	}
	proj.Mac = &driver.Mac
	t := time.Now()
	DateConv := t.Format(time.RFC3339)
	proj.CommissioningDate = &DateConv

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

		if elt.SwitchMac != "" {
			sw, _ := database.GetSwitchConfig(s.db, elt.SwitchMac)
			if sw != nil {
				ip := "0"
				if sw.IP != nil {
					ip = *sw.IP
				}
				switchConf := sd.SwitchConfig{}
				switchConf.Mac = elt.SwitchMac
				switchConf.IP = ip
				switchConf.LedsSetup = make(map[string]dl.LedSetup)
				switchConf.LedsSetup[elt.Mac] = *elt
				url := "/write/switch/" + elt.SwitchMac + "/update/settings"
				dump, _ := switchConf.ToJSON()
				s.server.SendCommand(url, dump)
			}
		}

		groupCfg, _ := database.GetGroupConfig(s.db, *elt.Group)
		newLeds := []string{}
		for _, led := range groupCfg.Leds {
			if led != elt.Mac {
				newLeds = append(newLeds, led)
			}
		}
		newLeds = append(newLeds, elt.Mac)
		groupCfg.Leds = newLeds

		if elt.FirstDay != nil && *elt.FirstDay == true {
			firstDay := []string{}
			for _, led := range groupCfg.FirstDay {
				if led != elt.Mac {
					firstDay = append(firstDay, led)
				}
			}
			firstDay = append(firstDay, elt.Mac)
			groupCfg.FirstDay = firstDay
		}

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
				url := "/write/switch/" + mac + "/update/settings"
				switchSetup := sd.SwitchConfig{}
				switchSetup.Mac = mac
				switchSetup.IP = ip
				switchSetup.Groups = make(map[int]gm.GroupConfig)
				switchSetup.Groups[groupCfg.Group] = *groupCfg
				dump, _ := switchSetup.ToJSON()
				s.server.SendCommand(url, dump)
			}
		}
	case pconst.BLIND:
		elt, _ := database.GetBlindLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Blind " + proj.Label + " in database")
			return
		}
		elt.Mac = driver.Mac
		database.SaveBlindLabelConfig(s.db, *elt)

		// send allow new driver configuration to the switch
		if elt.SwitchMac != "" {
			sw, _ := database.GetSwitchConfig(s.db, elt.SwitchMac)
			if sw != nil {
				ip := "0"
				if sw.IP != nil {
					ip = *sw.IP
				}
				switchConf := sd.SwitchConfig{}
				switchConf.Mac = elt.SwitchMac
				switchConf.IP = ip
				switchConf.BlindsSetup = make(map[string]db.BlindSetup)
				switchConf.BlindsSetup[elt.Mac] = *elt
				url := "/write/switch/" + elt.SwitchMac + "/update/settings"
				dump, _ := switchConf.ToJSON()
				s.server.SendCommand(url, dump)
			}
		}

		groupCfg, _ := database.GetGroupConfig(s.db, *elt.Group)
		newBlinds := []string{}
		for _, bld := range groupCfg.Blinds {
			if bld != elt.Mac {
				newBlinds = append(newBlinds, bld)
			}
		}
		newBlinds = append(newBlinds, elt.Mac)
		groupCfg.Blinds = newBlinds

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
				url := "/write/switch/" + mac + "/update/settings"
				switchSetup := sd.SwitchConfig{}
				switchSetup.Mac = mac
				switchSetup.IP = ip
				switchSetup.Groups = make(map[int]gm.GroupConfig)
				switchSetup.Groups[groupCfg.Group] = *groupCfg
				dump, _ := switchSetup.ToJSON()
				s.server.SendCommand(url, dump)
			}
		}

	case pconst.HVAC:
		elt, _ := database.GetHvacLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Hvac " + proj.Label + " in database")
			return
		}
		elt.Mac = driver.Mac
		database.SaveHvacLabelConfig(s.db, *elt)

		// send allow new driver configuration to the switch
		if elt.SwitchMac != "" {
			sw, _ := database.GetSwitchConfig(s.db, elt.SwitchMac)
			if sw != nil {
				ip := "0"
				if sw.IP != nil {
					ip = *sw.IP
				}
				switchConf := sd.SwitchConfig{}
				switchConf.Mac = elt.SwitchMac
				switchConf.IP = ip
				switchConf.HvacsSetup = make(map[string]dh.HvacSetup)
				switchConf.HvacsSetup[elt.Mac] = *elt
				url := "/write/switch/" + elt.SwitchMac + "/update/settings"
				dump, _ := switchConf.ToJSON()
				s.server.SendCommand(url, dump)
			}
		}

		groupCfg, _ := database.GetGroupConfig(s.db, *elt.Group)
		newHvacs := []string{}
		for _, hvac := range groupCfg.Hvacs {
			if hvac != elt.Mac {
				newHvacs = append(newHvacs, hvac)
			}
		}
		newHvacs = append(newHvacs, elt.Mac)
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
				url := "/write/switch/" + mac + "/update/settings"
				switchSetup := sd.SwitchConfig{}
				switchSetup.Mac = mac
				switchSetup.IP = ip
				switchSetup.Groups = make(map[int]gm.GroupConfig)
				switchSetup.Groups[groupCfg.Group] = *groupCfg
				dump, _ := switchSetup.ToJSON()
				s.server.SendCommand(url, dump)
			}
		}
	case pconst.SENSOR:
		elt, _ := database.GetSensorLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Sensor " + proj.Label + " in database")
			return
		}
		elt.Mac = driver.Mac
		database.SaveSensorLabelConfig(s.db, *elt)

		// send allow new driver configuration to the switch
		if elt.SwitchMac != "" {
			sw, _ := database.GetSwitchConfig(s.db, elt.SwitchMac)
			if sw != nil {
				ip := "0"
				if sw.IP != nil {
					ip = *sw.IP
				}
				switchConf := sd.SwitchConfig{}
				switchConf.Mac = elt.SwitchMac
				switchConf.IP = ip
				switchConf.SensorsSetup = make(map[string]ds.SensorSetup)
				switchConf.SensorsSetup[elt.Mac] = *elt
				url := "/write/switch/" + elt.SwitchMac + "/update/settings"
				dump, _ := switchConf.ToJSON()
				s.server.SendCommand(url, dump)
			}
		}

		groupCfg, _ := database.GetGroupConfig(s.db, *elt.Group)
		newSensors := []string{}
		for _, sens := range groupCfg.Sensors {
			if sens != elt.Mac {
				newSensors = append(newSensors, sens)
			}
		}
		newSensors = append(newSensors, elt.Mac)
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
				url := "/write/switch/" + mac + "/update/settings"
				switchSetup := sd.SwitchConfig{}
				switchSetup.Mac = mac
				switchSetup.IP = ip
				switchSetup.Groups = make(map[int]gm.GroupConfig)
				switchSetup.Groups[groupCfg.Group] = *groupCfg
				dump, _ := switchSetup.ToJSON()
				s.server.SendCommand(url, dump)
			}
		}

	case pconst.WAGO:
		elt, _ := database.GetWagoLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find Wago " + proj.Label + " in database")
			return
		}
		elt.Mac = driver.Mac
		database.SaveWagoLabelConfig(s.db, *elt)
		nanos := database.GetNanoSwitchSetup(s.db, elt.Cluster)
		for _, nano := range nanos {
			nano.Mac = elt.Mac + "." + strconv.Itoa(nano.ModbusOffset)
			database.SaveNanoLabelConfig(s.db, nano)

			projNano, _ := database.GetProject(s.db, nano.Label)
			if projNano != nil {
				projNano.Mac = &nano.Mac

				//update project
				database.SaveProject(s.db, *projNano)
			}

			groupCfg, _ := database.GetGroupConfig(s.db, nano.Group)
			newNanos := []string{}
			for _, n := range groupCfg.Nanosenses {
				if n != nano.Mac {
					newNanos = append(newNanos, n)
				}
			}
			newNanos = append(newNanos, nano.Mac)
			groupCfg.Nanosenses = newNanos

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
					url := "/write/switch/" + mac + "/update/settings"
					switchSetup := sd.SwitchConfig{}
					switchSetup.Mac = mac
					switchSetup.IP = ip
					switchSetup.Groups = make(map[int]gm.GroupConfig)
					switchSetup.Groups[groupCfg.Group] = *groupCfg
					dump, _ := switchSetup.ToJSON()
					s.server.SendCommand(url, dump)
				}
			}
		}
		s.sendSwitchWagoSetup(*elt)

	case pconst.SWITCH:
		elt := database.GetSwitchLabelConfig(s.db, proj.Label)
		if elt == nil {
			rlog.Error("Cannot find SWITCH " + proj.Label + " in database")
			return
		}
		elt.Mac = &driver.Mac
		database.SaveSwitchLabelConfig(s.db, *elt)
		//configuration will be sent with the next hello. It will avoid duplicate code.
	}
}
