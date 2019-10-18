package service

import (
	"strings"

	ds "github.com/energieip/common-components-go/pkg/dsensor"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) updateGroupSensor(oldSensor ds.SensorSetup, sensor ds.SensorSetup) {
	if sensor.Group != nil {
		if oldSensor.Group != sensor.Group {
			if oldSensor.Group != nil {
				rlog.Info("Update old group", *oldSensor.Group)
				gr, _ := database.GetGroupConfig(s.db, *oldSensor.Group)
				if gr != nil {
					sensors := []string{}
					for _, v := range gr.Sensors {
						if v != sensor.Mac {
							sensors = append(sensors, v)
						}
					}
					gr.Sensors = sensors
					rlog.Info("Old group will be ", gr.Sensors)
					s.updateGroupCfg(gr)
				}
			}
			rlog.Info("Update new group", *sensor.Group)
			grNew, _ := database.GetGroupConfig(s.db, *sensor.Group)
			if grNew != nil {
				if !inArray(sensor.Mac, grNew.Sensors) {
					grNew.Sensors = append(grNew.Sensors, sensor.Mac)
					rlog.Info("new group will be", grNew.Sensors)
					s.updateGroupCfg(grNew)
				}
			}
		}
	}
}

func (s *CoreService) sendSwitchSensorSetup(elt ds.SensorSetup) {
	if elt.SwitchMac == "" {
		return
	}
	sw, _ := database.GetSwitchConfig(s.db, elt.SwitchMac)
	if sw != nil {
		ip := "0"
		if sw.IP != nil {
			ip = *sw.IP
		}
		dumpFreq := 1000
		if sw.DumpFrequency != nil {
			dumpFreq = *sw.DumpFrequency
		}
		url := "/write/switch/" + elt.SwitchMac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = elt.SwitchMac
		switchSetup.IP = ip
		switchSetup.DumpFrequency = dumpFreq
		switchSetup.SensorsSetup = make(map[string]ds.SensorSetup)
		switchSetup.SensorsSetup[elt.Mac] = elt
		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) updateSensorCfg(config interface{}) {
	cfg, _ := ds.ToSensorConf(config)

	var oldSensor *ds.SensorSetup
	var sensor *ds.SensorSetup
	if cfg.Mac != "" {
		oldSensor, _ = database.GetSensorConfig(s.db, cfg.Mac)
		if oldSensor == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}

		database.UpdateSensorConfig(s.db, *cfg)
		//Get corresponding switchMac
		sensor, _ = database.GetSensorConfig(s.db, cfg.Mac)
		if sensor == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}
	} else {
		if cfg.Label == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}
		oldSensor, _ = database.GetSensorLabelConfig(s.db, *cfg.Label)
		if oldSensor == nil {
			rlog.Error("Cannot find config for " + *cfg.Label)
			return
		}

		database.UpdateSensorLabelConfig(s.db, *cfg)
		//Get corresponding switchMac
		sensor, _ = database.GetSensorLabelConfig(s.db, *cfg.Label)
		if sensor == nil {
			rlog.Error("Cannot find config for " + *cfg.Label)
			return
		}
	}

	s.updateGroupSensor(*oldSensor, *sensor)
	if sensor.SwitchMac == "" {
		rlog.Info("No corresponding switch found for " + cfg.Mac)
		return
	}

	sw, _ := database.GetSwitchConfig(s.db, sensor.SwitchMac)
	if sw != nil {
		ip := "0"
		if sw.IP != nil {
			ip = *sw.IP
		}
		dumpFreq := 1000
		if sw.DumpFrequency != nil {
			dumpFreq = *sw.DumpFrequency
		}
		url := "/write/switch/" + sensor.SwitchMac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sensor.SwitchMac
		switchSetup.IP = ip
		switchSetup.DumpFrequency = dumpFreq
		switchSetup.SensorsConfig = make(map[string]ds.SensorConf)
		switchSetup.SensorsConfig[cfg.Mac] = *cfg
		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) updateSensorSetup(config interface{}) {
	byLbl := false
	cfg, _ := ds.ToSensorSetup(config)
	if cfg == nil {
		return
	}

	oldSensor, _ := database.GetSensorConfig(s.db, cfg.Mac)
	if oldSensor == nil && cfg.Label != nil {
		oldSensor, _ = database.GetSensorLabelConfig(s.db, *cfg.Label)
		if oldSensor != nil {
			//it means that the IFC has been uploaded but the MAC is unknown
			byLbl = true
		}
	}

	if oldSensor != nil {
		s.updateGroupSensor(*oldSensor, *cfg)
	}
	var sensor *ds.SensorSetup
	if byLbl {
		database.UpdateSensorLabelSetup(s.db, *cfg)
		//Get correspnding switchMac
		sensor, _ = database.GetSensorLabelConfig(s.db, *cfg.Label)
		if sensor == nil {
			rlog.Error("Cannot find config for " + *cfg.Label)
			return
		}

	} else {
		database.UpdateSensorSetup(s.db, *cfg)
		//Get correspnding switchMac
		sensor, _ = database.GetSensorConfig(s.db, cfg.Mac)
		if sensor == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}
	}

	if sensor.SwitchMac != "" {
		cfg.SwitchMac = sensor.SwitchMac
	}
	s.sendSwitchSensorSetup(*cfg)
}

func (s *CoreService) createSensorLabelSetup(config interface{}) {
	cfg, _ := ds.ToSensorSetup(config)
	if cfg == nil || cfg.Label == nil {
		rlog.Error("Cannot parse ")
		return
	}
	cfg.Mac = strings.ToUpper(cfg.Mac)
	if cfg.SwitchMac != "" {
		cfg.SwitchMac = strings.ToUpper(cfg.SwitchMac)
	}

	database.CreateSensorLabelSetup(s.db, *cfg)
}

func (s *CoreService) updateSensorLabelSetup(config interface{}) {
	cfg, _ := ds.ToSensorSetup(config)
	if cfg == nil || cfg.Label == nil {
		return
	}
	cfg.Mac = strings.ToUpper(cfg.Mac)
	if cfg.SwitchMac != "" {
		cfg.SwitchMac = strings.ToUpper(cfg.SwitchMac)
	}

	oldSensor, _ := database.GetSensorLabelConfig(s.db, *cfg.Label)
	if oldSensor != nil {
		s.updateGroupSensor(*oldSensor, *cfg)
	}

	database.UpdateSensorLabelSetup(s.db, *cfg)
	//Get correspnding switchMac
	sensor, _ := database.GetSensorLabelConfig(s.db, *cfg.Label)
	if sensor == nil {
		rlog.Error("Cannot find config for " + *cfg.Label)
		return
	}
	if sensor.SwitchMac != "" {
		cfg.SwitchMac = strings.ToUpper(sensor.SwitchMac)
	}
	s.sendSwitchSensorSetup(*cfg)
}
