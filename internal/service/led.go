package service

import (
	dl "github.com/energieip/common-components-go/pkg/dled"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

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
					for i, v := range gr.Leds {
						if v == led.Mac {
							gr.Leds = append(gr.Leds[:i], gr.Leds[i+1:]...)
							break
						}
					}
					rlog.Info("Old group will be ", gr.Leds)
					s.updateGroupCfg(gr)
				}
			}
			rlog.Info("Update new group", *led.Group)
			grNew, _ := database.GetGroupConfig(s.db, *led.Group)
			if grNew != nil {
				grNew.Leds = append(grNew.Leds, cfg.Mac)
				rlog.Info("new group will be", grNew.Leds)
				s.updateGroupCfg(grNew)
			}
		}
	}
	url := "/write/switch/" + led.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = led.SwitchMac
	switchSetup.LedsConfig = make(map[string]dl.LedConf)

	switchSetup.LedsConfig[cfg.Mac] = *cfg

	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send update config to " + led.SwitchMac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + led.SwitchMac + " on topic: " + url + " dump:" + dump)
	}
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
	rlog.Info("Ready to send ", ledCfg)
	rlog.Info("To switch", led.SwitchMac)
	switchSetup.LedsConfig[led.Mac] = ledCfg

	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send update config to " + led.SwitchMac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + led.SwitchMac + " on topic: " + url + " dump:" + dump)
	}
}
