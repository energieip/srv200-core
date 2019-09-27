package service

import (
	"strings"

	"github.com/energieip/common-components-go/pkg/dblind"
	"github.com/energieip/common-components-go/pkg/dserver"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) updateGroupBlind(oldBlind dblind.BlindSetup, blind dblind.BlindSetup) {
	if blind.Group != nil {
		if oldBlind.Group != blind.Group {
			if oldBlind.Group != nil {
				rlog.Info("Update old group", *oldBlind.Group)
				gr, _ := database.GetGroupConfig(s.db, *oldBlind.Group)
				if gr != nil {
					blinds := []string{}
					for _, v := range gr.Blinds {
						if v != blind.Mac {
							blinds = append(blinds, v)
						}
					}
					gr.Blinds = blinds
					rlog.Info("Old group will be ", gr.Blinds)
					s.updateGroupCfg(gr)
				}
			}
			rlog.Info("Update new group", *blind.Group)
			grNew, _ := database.GetGroupConfig(s.db, *blind.Group)
			if grNew != nil && blind.Mac != "" && !inArray(blind.Mac, grNew.Blinds) {
				grNew.Blinds = append(grNew.Blinds, blind.Mac)
				rlog.Info("new group will be", grNew.Blinds)
				s.updateGroupCfg(grNew)
			}
		}
	}
}

func (s *CoreService) sendSwitchBlindSetup(bld dblind.BlindSetup) {
	if bld.SwitchMac == "" {
		return
	}
	sw, _ := database.GetSwitchConfig(s.db, bld.SwitchMac)
	if sw != nil {
		ip := "0"
		if sw.IP != nil {
			ip = *sw.IP
		}
		url := "/write/switch/" + bld.SwitchMac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = bld.SwitchMac
		switchSetup.IP = ip
		switchSetup.BlindsSetup = make(map[string]dblind.BlindSetup)
		switchSetup.BlindsSetup[bld.Mac] = bld

		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) updateBlindCfg(config interface{}) {
	cfg, _ := dblind.ToBlindConf(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	var oldBlind *dblind.BlindSetup
	var blind *dblind.BlindSetup
	if cfg.Mac != "" {
		oldBlind, _ = database.GetBlindConfig(s.db, cfg.Mac)
		if oldBlind == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}

		database.UpdateBlindConfig(s.db, *cfg)
		//Get corresponding switchMac
		blind, _ = database.GetBlindConfig(s.db, cfg.Mac)
		if blind == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}
	} else {
		if cfg.Label == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}
		oldBlind, _ = database.GetBlindLabelConfig(s.db, *cfg.Label)
		if oldBlind == nil {
			rlog.Error("Cannot find config for " + *cfg.Label)
			return
		}

		database.UpdateBlindLabelConfig(s.db, *cfg)
		//Get corresponding switchMac
		blind, _ = database.GetBlindLabelConfig(s.db, *cfg.Label)
		if blind == nil {
			rlog.Error("Cannot find config for " + *cfg.Label)
			return
		}
	}
	s.updateGroupBlind(*oldBlind, *blind)

	if blind.SwitchMac == "" {
		rlog.Info("No corresponding switch present for " + cfg.Mac)
		return
	}

	sw, _ := database.GetSwitchConfig(s.db, blind.SwitchMac)
	if sw != nil {
		ip := "0"
		if sw.IP != nil {
			ip = *sw.IP
		}
		url := "/write/switch/" + blind.SwitchMac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = blind.SwitchMac
		switchSetup.IP = ip
		switchSetup.BlindsConfig = make(map[string]dblind.BlindConf)

		switchSetup.BlindsConfig[cfg.Mac] = *cfg

		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) sendBlindCmd(cmdBlind interface{}) {
	cmd, _ := dserver.ToBlindCmd(cmdBlind)
	if cmd == nil {
		rlog.Error("Cannot parse cmd")
		return
	}
	//Get correspnding switchMac
	driver, _ := database.GetBlindConfig(s.db, cmd.Mac)
	if driver == nil {
		rlog.Error("Cannot find config for " + cmd.Mac)
		return
	}
	if driver.SwitchMac == "" {
		rlog.Error("No corresponding switch present for " + cmd.Mac)
		return
	}
	sw, _ := database.GetSwitchConfig(s.db, driver.SwitchMac)
	if sw != nil {
		ip := "0"
		if sw.IP != nil {
			ip = *sw.IP
		}
		url := "/write/switch/" + driver.SwitchMac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = driver.SwitchMac
		switchSetup.IP = ip
		switchSetup.BlindsConfig = make(map[string]dblind.BlindConf)

		cfg := dblind.BlindConf{
			Mac:    cmd.Mac,
			Blind1: cmd.Blind1,
			Blind2: cmd.Blind2,
			Slat1:  cmd.Slat1,
			Slat2:  cmd.Slat2,
		}
		switchSetup.BlindsConfig[cmd.Mac] = cfg

		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) updateBlindSetup(config interface{}) {
	byLbl := false
	cfg, _ := dblind.ToBlindSetup(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	oldBlind, _ := database.GetBlindConfig(s.db, cfg.Mac)
	if oldBlind == nil && cfg.Label != nil {
		oldBlind, _ = database.GetBlindLabelConfig(s.db, *cfg.Label)
		if oldBlind != nil {
			//it means that the IFC has been uploaded but the MAC is unknown
			byLbl = true
		}
	}

	if oldBlind != nil {
		s.updateGroupBlind(*oldBlind, *cfg)
	}

	var blind *dblind.BlindSetup
	if byLbl {
		database.UpdateBlindLabelSetup(s.db, *cfg)
		//Get corresponding switchMac
		blind, _ = database.GetBlindLabelConfig(s.db, *cfg.Label)
		if blind == nil {
			rlog.Error("Cannot find config for " + *cfg.Label)
			return
		}
	} else {
		database.UpdateBlindSetup(s.db, *cfg)
		//Get corresponding switchMac
		blind, _ = database.GetBlindConfig(s.db, cfg.Mac)
		if blind == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}
	}
	if blind.SwitchMac != "" {
		cfg.SwitchMac = blind.SwitchMac
	}

	s.sendSwitchBlindSetup(*cfg)
}

func (s *CoreService) updateBlindLabelSetup(config interface{}) {
	cfg, _ := dblind.ToBlindSetup(config)
	if cfg == nil || cfg.Label == nil {
		rlog.Error("Cannot parse ")
		return
	}
	cfg.Mac = strings.ToUpper(cfg.Mac)
	cfg.SwitchMac = strings.ToUpper(cfg.SwitchMac)

	oldBlind, _ := database.GetBlindLabelConfig(s.db, *cfg.Label)
	if oldBlind != nil {
		s.updateGroupBlind(*oldBlind, *cfg)
	}
	database.UpdateBlindLabelSetup(s.db, *cfg)
	//Get corresponding switchMac
	blind, _ := database.GetBlindLabelConfig(s.db, *cfg.Label)
	if blind == nil {
		rlog.Error("Cannot find config for " + *cfg.Label)
		return
	}
	if blind.SwitchMac != "" {
		cfg.SwitchMac = blind.SwitchMac
	}
	s.sendSwitchBlindSetup(*cfg)
}
