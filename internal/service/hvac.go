package service

import (
	"strings"

	"github.com/energieip/common-components-go/pkg/dhvac"
	"github.com/energieip/common-components-go/pkg/dserver"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) updateGroupHvac(oldHvac dhvac.HvacSetup, hvac dhvac.HvacSetup) {
	if hvac.Group != nil {
		if oldHvac.Group != hvac.Group {
			if oldHvac.Group != nil {
				rlog.Info("Update old group", *oldHvac.Group)
				gr, _ := database.GetGroupConfig(s.db, *oldHvac.Group)
				if gr != nil {
					hvacs := []string{}
					for _, v := range gr.Hvacs {
						if v != hvac.Mac {
							hvacs = append(hvacs, v)
						}
					}
					gr.Hvacs = hvacs
					rlog.Info("Old group will be ", gr.Hvacs)
					s.updateGroupCfg(gr)
				}
			}
			rlog.Info("Update new group", *hvac.Group)
			grNew, _ := database.GetGroupConfig(s.db, *hvac.Group)
			if grNew != nil && hvac.Mac != "" && !inArray(hvac.Mac, grNew.Hvacs) {
				grNew.Hvacs = append(grNew.Hvacs, hvac.Mac)
				rlog.Info("new group will be", grNew.Hvacs)
				s.updateGroupCfg(grNew)
			}
		}
	}
}

func (s *CoreService) sendSwitchHvacSetup(elt dhvac.HvacSetup) {
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
		switchSetup.HvacsSetup = make(map[string]dhvac.HvacSetup)
		switchSetup.HvacsSetup[elt.Mac] = elt
		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) updateHvacCfg(config interface{}) {
	cfg, _ := dhvac.ToHvacConf(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	var oldHvac *dhvac.HvacSetup
	var hvac *dhvac.HvacSetup
	if cfg.Mac != "" {
		oldHvac, _ = database.GetHvacConfig(s.db, cfg.Mac)
		if oldHvac == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}

		database.UpdateHvacConfig(s.db, *cfg)
		//Get corresponding switchMac
		hvac, _ = database.GetHvacConfig(s.db, cfg.Mac)
		if hvac == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}
	} else {
		if cfg.Label == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}
		oldHvac, _ = database.GetHvacLabelConfig(s.db, *cfg.Label)
		if oldHvac == nil {
			rlog.Error("Cannot find config for " + *cfg.Label)
			return
		}

		database.UpdateHvacLabelConfig(s.db, *cfg)
		//Get corresponding switchMac
		hvac, _ = database.GetHvacLabelConfig(s.db, *cfg.Label)
		if hvac == nil {
			rlog.Error("Cannot find config for " + *cfg.Label)
			return
		}
	}
	s.updateGroupHvac(*oldHvac, *hvac)

	if hvac.SwitchMac == "" {
		rlog.Info("No corresponding switch found for " + cfg.Mac)
		return
	}

	sw, _ := database.GetSwitchConfig(s.db, hvac.SwitchMac)
	if sw != nil {
		ip := "0"
		if sw.IP != nil {
			ip = *sw.IP
		}
		dumpFreq := 1000
		if sw.DumpFrequency != nil {
			dumpFreq = *sw.DumpFrequency
		}
		url := "/write/switch/" + hvac.SwitchMac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = hvac.SwitchMac
		switchSetup.IP = ip
		switchSetup.DumpFrequency = dumpFreq
		switchSetup.HvacsConfig = make(map[string]dhvac.HvacConf)
		switchSetup.HvacsConfig[cfg.Mac] = *cfg
		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) sendHvacCmd(cmdHvac interface{}) {
	cmd, _ := dserver.ToHvacCmd(cmdHvac)
	if cmd == nil {
		rlog.Error("Cannot parse cmd")
		return
	}
	//Get correspnding switchMac
	driver, _ := database.GetHvacConfig(s.db, cmd.Mac)
	if driver == nil {
		rlog.Error("Cannot find config for " + cmd.Mac)
		return
	}
	if driver.SwitchMac == "" {
		rlog.Error("No corresponding switch found for " + cmd.Mac)
		return
	}
	sw, _ := database.GetSwitchConfig(s.db, driver.SwitchMac)
	if sw != nil {
		ip := "0"
		if sw.IP != nil {
			ip = *sw.IP
		}
		dumpFreq := 1000
		if sw.DumpFrequency != nil {
			dumpFreq = *sw.DumpFrequency
		}
		url := "/write/switch/" + driver.SwitchMac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = driver.SwitchMac
		switchSetup.IP = ip
		switchSetup.DumpFrequency = dumpFreq
		switchSetup.HvacsConfig = make(map[string]dhvac.HvacConf)
		cfg := dhvac.HvacConf{
			Mac:   cmd.Mac,
			Shift: &cmd.ShiftTemp,
		}
		switchSetup.HvacsConfig[cmd.Mac] = cfg
		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) updateHvacLabelSetup(config interface{}) {
	cfg, _ := dhvac.ToHvacSetup(config)
	if cfg == nil || cfg.Label == nil {
		rlog.Error("Cannot parse ")
		return
	}
	cfg.Mac = strings.ToUpper(cfg.Mac)
	if cfg.SwitchMac != "" {
		cfg.SwitchMac = strings.ToUpper(cfg.SwitchMac)
	}
	oldHvac, _ := database.GetHvacLabelConfig(s.db, *cfg.Label)
	if oldHvac != nil {
		s.updateGroupHvac(*oldHvac, *cfg)
	}

	database.UpdateHvacLabelSetup(s.db, *cfg)

	//Get corresponding switchMac
	hvac, _ := database.GetHvacLabelConfig(s.db, *cfg.Label)
	if hvac == nil {
		rlog.Error("Cannot find config for " + *cfg.Label)
		return
	}
	if hvac.SwitchMac == "" {
		rlog.Info("No corresponding switch found for " + *cfg.Label)
		return
	}
	s.sendSwitchHvacSetup(*hvac)
}

func (s *CoreService) updateHvacSetup(config interface{}) {
	byLbl := false
	cfg, _ := dhvac.ToHvacSetup(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	oldHvac, _ := database.GetHvacConfig(s.db, cfg.Mac)
	if oldHvac == nil && cfg.Label != nil {
		oldHvac, _ = database.GetHvacLabelConfig(s.db, *cfg.Label)
		if oldHvac != nil {
			//it means that the IFC has been uploaded but the MAC is unknown
			byLbl = true
		}
	}
	if oldHvac != nil {
		s.updateGroupHvac(*oldHvac, *cfg)
	}

	var hvac *dhvac.HvacSetup
	if byLbl {
		database.UpdateHvacLabelSetup(s.db, *cfg)

		//Get corresponding switchMac
		hvac, _ = database.GetHvacLabelConfig(s.db, *cfg.Label)
		if hvac == nil {
			rlog.Error("Cannot find config for " + *cfg.Label)
			return
		}
	} else {
		database.UpdateHvacSetup(s.db, *cfg)

		//Get corresponding switchMac
		hvac, _ = database.GetHvacConfig(s.db, cfg.Mac)
		if hvac == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}
	}

	s.sendSwitchHvacSetup(*hvac)
}
