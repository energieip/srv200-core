package service

import (
	"github.com/energieip/common-components-go/pkg/dhvac"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) updateHvacCfg(config interface{}) {
	cfg, _ := dhvac.ToHvacConf(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	oldHvac, _ := database.GetHvacConfig(s.db, cfg.Mac)
	if oldHvac == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	database.UpdateHvacConfig(s.db, *cfg)
	//Get corresponding switchMac
	hvac, _ := database.GetHvacConfig(s.db, cfg.Mac)
	if hvac == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

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
					// s.updateDriverGroup(gr.Group)
				}
			}
			rlog.Info("Update new group", *hvac.Group)
			grNew, _ := database.GetGroupConfig(s.db, *hvac.Group)
			if grNew != nil {
				grNew.Hvacs = append(grNew.Hvacs, cfg.Mac)
				rlog.Info("new group will be", grNew.Hvacs)
				s.updateGroupCfg(grNew)
				// s.updateDriverGroup(grNew.Group)
			}
		}
	}
	url := "/write/switch/" + hvac.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = hvac.SwitchMac
	switchSetup.HvacsConfig = make(map[string]dhvac.HvacConf)

	switchSetup.HvacsConfig[cfg.Mac] = *cfg

	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) sendHvacCmd(cmdHvac interface{}) {
	cmd, _ := core.ToHvacCmd(cmdHvac)
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
	url := "/write/switch/" + driver.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = driver.SwitchMac
	switchSetup.HvacsConfig = make(map[string]dhvac.HvacConf)

	//TODO
	cfg := dhvac.HvacConf{
		Mac: cmd.Mac,
	}
	switchSetup.HvacsConfig[cmd.Mac] = cfg

	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) updateHvacLabelSetup(config interface{}) {
	cfg, _ := dhvac.ToHvacSetup(config)
	if cfg == nil || cfg.Label == nil {
		rlog.Error("Cannot parse ")
		return
	}

	oldHvac, _ := database.GetHvacLabelConfig(s.db, *cfg.Label)
	if oldHvac != nil {
		if cfg.Group != nil {
			if oldHvac.Group != cfg.Group {
				if oldHvac.Group != nil {
					rlog.Info("Update old group", *oldHvac.Group)
					gr, _ := database.GetGroupConfig(s.db, *oldHvac.Group)
					if gr != nil {
						hvacs := []string{}
						for _, v := range gr.Hvacs {
							if v != cfg.Mac {
								hvacs = append(hvacs, v)
							}
						}
						gr.Hvacs = hvacs
						rlog.Info("Old group will be ", gr.Hvacs)
						s.updateGroupCfg(gr)
					}
				}
				rlog.Info("Update new group", *cfg.Group)
				grNew, _ := database.GetGroupConfig(s.db, *cfg.Group)
				if grNew != nil {
					grNew.Hvacs = append(grNew.Hvacs, cfg.Mac)
					rlog.Info("new group will be", grNew.Hvacs)
					s.updateGroupCfg(grNew)
				}
			}
		}
	}

	database.UpdateHvacLabelSetup(s.db, *cfg)

	//Get corresponding switchMac
	hvac, _ := database.GetHvacConfig(s.db, cfg.Mac)
	if hvac == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	if hvac.SwitchMac == "" {
		return
	}

	url := "/write/switch/" + hvac.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = hvac.SwitchMac
	switchSetup.HvacsSetup = make(map[string]dhvac.HvacSetup)

	switchSetup.HvacsSetup[cfg.Mac] = *cfg

	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) updateHvacSetup(config interface{}) {
	cfg, _ := dhvac.ToHvacSetup(config)
	if cfg == nil || cfg.Label == nil {
		rlog.Error("Cannot parse ")
		return
	}

	oldHvac, _ := database.GetHvacConfig(s.db, cfg.Mac)
	if oldHvac != nil {
		if cfg.Group != nil {
			if oldHvac.Group != cfg.Group {
				if oldHvac.Group != nil {
					rlog.Info("Update old group", *oldHvac.Group)
					gr, _ := database.GetGroupConfig(s.db, *oldHvac.Group)
					if gr != nil {
						hvacs := []string{}
						for _, v := range gr.Hvacs {
							if v != cfg.Mac {
								hvacs = append(hvacs, v)
							}
						}
						gr.Hvacs = hvacs
						rlog.Info("Old group will be ", gr.Hvacs)
						s.updateGroupCfg(gr)
					}
				}
				rlog.Info("Update new group", *cfg.Group)
				grNew, _ := database.GetGroupConfig(s.db, *cfg.Group)
				if grNew != nil {
					grNew.Hvacs = append(grNew.Hvacs, cfg.Mac)
					rlog.Info("new group will be", grNew.Hvacs)
					s.updateGroupCfg(grNew)
				}
			}
		}
	}

	database.UpdateHvacSetup(s.db, *cfg)

	//Get corresponding switchMac
	hvac, _ := database.GetHvacConfig(s.db, cfg.Mac)
	if hvac == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	if hvac.SwitchMac == "" {
		return
	}

	url := "/write/switch/" + hvac.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = hvac.SwitchMac
	switchSetup.HvacsSetup = make(map[string]dhvac.HvacSetup)

	switchSetup.HvacsSetup[cfg.Mac] = *cfg

	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}
