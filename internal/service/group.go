package service

import (
	"github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) isGroupRequiredUpdate(old gm.GroupStatus, new gm.GroupConfig) bool {
	for i, v := range old.Leds {
		if v != new.Leds[i] {
			return true
		}
	}
	if old.RuleBrightness != nil && new.RuleBrightness != nil {
		if *old.RuleBrightness != *new.RuleBrightness {
			return true
		}
	}
	if old.RuleBrightness != nil && new.RuleBrightness == nil {
		return true
	}
	if old.RuleBrightness == nil && new.RuleBrightness != nil {
		return true
	}

	if old.RulePresence != nil && new.RulePresence != nil {
		if *old.RulePresence != *new.RulePresence {
			return true
		}
	}
	if old.RulePresence != nil && new.RulePresence == nil {
		return true
	}
	if old.RulePresence == nil && new.RulePresence != nil {
		return true
	}
	for i, v := range old.Sensors {
		if v != new.Sensors[i] {
			return true
		}
	}
	for i, v := range old.Blinds {
		if v != new.Blinds[i] {
			return true
		}
	}
	if new.FriendlyName != nil {
		if old.FriendlyName != *new.FriendlyName {
			return true
		}
	}
	if new.SensorRule != nil {
		if old.SensorRule != *new.SensorRule {
			return true
		}
	}
	if new.SlopeStartAuto != nil {
		if old.SlopeStartAuto != *new.SlopeStartAuto {
			return true
		}
	}

	if new.SlopeStopAuto != nil {
		if old.SlopeStopAuto != *new.SlopeStopAuto {
			return true
		}
	}
	if new.SlopeStartManual != nil {
		if old.SlopeStartManual != *new.SlopeStartManual {
			return true
		}
	}

	if new.SlopeStopManual != nil {
		if old.SlopeStopManual != *new.SlopeStopManual {
			return true
		}
	}
	if new.CorrectionInterval != nil {
		if old.CorrectionInterval != *new.CorrectionInterval {
			return true
		}
	}
	if new.Watchdog != nil {
		if old.Watchdog != *new.Watchdog {
			return true
		}
	}
	return false
}

func (s *CoreService) updateGroupCfg(config interface{}) {
	cfg, _ := gm.ToGroupConfig(config)

	gr, _ := database.GetGroupConfig(s.db, cfg.Group)
	if gr != nil {
		database.UpdateGroupConfig(s.db, *cfg)
	} else {
		database.SaveGroupConfig(s.db, *cfg)

		for _, led := range cfg.Leds {
			light := dl.LedConf{
				Mac:   led,
				Group: &cfg.Group,
			}
			s.updateLedCfg(light)
		}
		for _, sensor := range cfg.Sensors {
			cell := ds.SensorConf{
				Mac:   sensor,
				Group: &cfg.Group,
			}
			s.updateSensorCfg(cell)
		}
		for _, blind := range cfg.Blinds {
			bl := dblind.BlindConf{
				Mac:   blind,
				Group: &cfg.Group,
			}
			s.updateBlindCfg(bl)
		}
	}

	for sw := range database.GetGroupSwitchs(s.db, cfg.Group) {
		url := "/write/switch/" + sw + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sw
		switchSetup.Groups = make(map[int]gm.GroupConfig)
		switchSetup.Groups[cfg.Group] = *cfg
		dump, _ := switchSetup.ToJSON()
		err := s.server.SendCommand(url, dump)
		if err != nil {
			rlog.Error("Cannot send update group config to " + sw + " on topic: " + url + " err:" + err.Error())
		} else {
			rlog.Info("Send update group config to " + sw + " on topic: " + url + " dump:" + dump)
		}
	}
}

func (s *CoreService) sendGroupCmd(cmd interface{}) {
	cmdGr, _ := core.ToGroupCmd(cmd)
	if cmdGr == nil {
		rlog.Error("Cannot parse cmd")
		return
	}
	for sw := range database.GetGroupSwitchs(s.db, cmdGr.Group) {
		url := "/write/switch/" + sw + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sw
		switchSetup.Groups = make(map[int]gm.GroupConfig)
		cfg := gm.GroupConfig{}
		cfg.Group = cmdGr.Group
		cfg.Auto = cmdGr.Auto
		cfg.SetpointLeds = cmdGr.SetpointLeds
		cfg.SetpointBlinds = cmdGr.SetpointBlinds
		cfg.SetpointSlatBlinds = cmdGr.SetpointSlats
		switchSetup.Groups[cmdGr.Group] = cfg
		dump, _ := switchSetup.ToJSON()
		err := s.server.SendCommand(url, dump)
		if err != nil {
			rlog.Error("Cannot group command to " + sw + " on topic: " + url + " err:" + err.Error())
		} else {
			rlog.Info("Send group command to " + sw + " on topic: " + url + " dump:" + dump)
		}
	}

}
