package service

import (
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	dl "github.com/energieip/common-components-go/pkg/dled"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) sendSwitchLedSetup(led dl.LedSetup) {
	if led.SwitchMac == "" {
		return
	}

	url := "/write/switch/" + led.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = led.SwitchMac
	switchSetup.LedsSetup = make(map[string]dl.LedSetup)

	switchSetup.LedsSetup[led.Mac] = led

	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) updateDriverGroup(grID int) {
	gr, _ := database.GetGroupConfig(s.db, grID)
	if gr == nil {
		return
	}

	for sw := range database.GetGroupSwitchs(s.db, grID) {
		url := "/write/switch/" + sw + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sw
		switchSetup.Groups = make(map[int]gm.GroupConfig)
		switchSetup.Groups[gr.Group] = *gr
		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) updateLedCfg(config interface{}) {
	cfg, _ := dl.ToLedConf(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	oldLed, _ := database.GetLedConfig(s.db, cfg.Mac)
	if oldLed == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	database.UpdateLedConfig(s.db, *cfg)
	//Get corresponding switchMac
	led, _ := database.GetLedConfig(s.db, cfg.Mac)
	if led == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	if led.Group != nil {
		if oldLed.Group != led.Group {
			if oldLed.Group != nil {
				rlog.Info("Update old group", *oldLed.Group)
				gr, _ := database.GetGroupConfig(s.db, *oldLed.Group)
				if gr != nil {
					leds := []string{}
					for _, v := range gr.Leds {
						if v != led.Mac {
							leds = append(leds, v)
						}
					}
					gr.Leds = leds
					firstDay := []string{}
					for _, v := range gr.FirstDay {
						if v != led.Mac {
							firstDay = append(firstDay, v)
						}
					}
					gr.FirstDay = firstDay
					rlog.Info("Old group will be ", gr.Leds)
					s.updateGroupCfg(gr)
					//	s.updateDriverGroup(gr.Group)
				}
			}
			rlog.Info("Update new group", *led.Group)
			grNew, _ := database.GetGroupConfig(s.db, *led.Group)
			if grNew != nil {
				grNew.Leds = append(grNew.Leds, cfg.Mac)
				rlog.Info("new group will be", grNew.Leds)
				s.updateGroupCfg(grNew)
				//	s.updateDriverGroup(grNew.Group)
			}
		}
	}
	url := "/write/switch/" + led.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = led.SwitchMac
	switchSetup.LedsConfig = make(map[string]dl.LedConf)

	switchSetup.LedsConfig[cfg.Mac] = *cfg

	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) updateLedSetup(config interface{}) {
	cfg, _ := dl.ToLedSetup(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	oldLed, _ := database.GetLedConfig(s.db, cfg.Mac)
	if oldLed != nil {
		if cfg.Group != nil {
			if oldLed.Group != cfg.Group {
				if oldLed.Group != nil {
					rlog.Info("Update old group", *oldLed.Group)
					gr, _ := database.GetGroupConfig(s.db, *oldLed.Group)
					if gr != nil {
						leds := []string{}
						for _, v := range gr.Leds {
							if v != cfg.Mac {
								leds = append(leds, v)
							}
						}
						gr.Leds = leds
						firstDay := []string{}
						for _, v := range gr.FirstDay {
							if v != cfg.Mac {
								firstDay = append(firstDay, v)
							}
						}
						gr.FirstDay = firstDay
						rlog.Info("Old group will be ", gr.Leds)
						s.updateGroupCfg(gr)
					}
				}
				rlog.Info("Update new group", *cfg.Group)
				grNew, _ := database.GetGroupConfig(s.db, *cfg.Group)
				if grNew != nil {
					grNew.Leds = append(grNew.Leds, cfg.Mac)
					rlog.Info("new group will be", grNew.Leds)
					s.updateGroupCfg(grNew)
				}
			}
		}
	}

	database.UpdateLedSetup(s.db, *cfg)
	//Get corresponding switchMac
	led, _ := database.GetLedConfig(s.db, cfg.Mac)
	if led == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}
	rlog.Info("Led configuration " + led.Mac + " saved")
	s.sendSwitchLedSetup(*cfg)
}

func (s *CoreService) updateLedLabelSetup(config interface{}) {
	cfg, _ := dl.ToLedSetup(config)
	if cfg == nil || cfg.Label == nil {
		rlog.Error("Cannot parse ")
		return
	}

	oldLed, _ := database.GetLedLabelConfig(s.db, *cfg.Label)
	if oldLed != nil {
		if cfg.Group != nil {
			if oldLed.Group != cfg.Group {
				if oldLed.Group != nil {
					rlog.Info("Update old group", *oldLed.Group)
					gr, _ := database.GetGroupConfig(s.db, *oldLed.Group)
					if gr != nil {
						leds := []string{}
						for _, v := range gr.Leds {
							if v != cfg.Mac {
								leds = append(leds, v)
							}
						}
						gr.Leds = leds
						firstDay := []string{}
						for _, v := range gr.FirstDay {
							if v != cfg.Mac {
								firstDay = append(firstDay, v)
							}
						}
						gr.FirstDay = firstDay
						rlog.Info("Old group will be ", gr.Leds)
						s.updateGroupCfg(gr)
					}
				}
				rlog.Info("Update new group", *cfg.Group)
				grNew, _ := database.GetGroupConfig(s.db, *cfg.Group)
				if grNew != nil {
					grNew.Leds = append(grNew.Leds, cfg.Mac)
					rlog.Info("new group will be", grNew.Leds)
					s.updateGroupCfg(grNew)
				}
			}
		}
	}

	database.UpdateLedLabelSetup(s.db, *cfg)
	//Get corresponding switchMac
	led, _ := database.GetLedConfig(s.db, cfg.Mac)
	if led == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}
	rlog.Info("Led configuration " + *cfg.Label + " saved")
	s.sendSwitchLedSetup(*cfg)
}

func (s *CoreService) sendLedCmd(cmd interface{}) {
	cmdLed, _ := core.ToLedCmd(cmd)
	if cmdLed == nil {
		rlog.Error("Cannot parse cmd")
		return
	}
	//Get correspnding switchMac
	led, _ := database.GetLedConfig(s.db, cmdLed.Mac)
	if led == nil {
		rlog.Error("Cannot find config for " + cmdLed.Mac)
		return
	}
	url := "/write/switch/" + led.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = led.SwitchMac
	switchSetup.LedsConfig = make(map[string]dl.LedConf)

	auto := cmdLed.Auto
	setpoint := cmdLed.Setpoint

	ledCfg := dl.LedConf{
		Mac:            led.Mac,
		Auto:           &auto,
		SetpointManual: &setpoint,
	}
	switchSetup.LedsConfig[led.Mac] = ledCfg

	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}
