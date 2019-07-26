package service

import (
	"strings"

	dnano "github.com/energieip/common-components-go/pkg/dnanosense"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) sendSwitchNanoSetup(driver dnano.NanosenseSetup) {
	for sw := range database.GetCluster(s.db, driver.Cluster) {
		if sw == "" {
			continue
		}
		url := "/write/switch/" + sw + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sw
		switchSetup.NanosSetup = make(map[string]dnano.NanosenseSetup)
		switchSetup.NanosSetup[driver.Label] = driver

		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) updateNanoCfg(config interface{}) {
	cfg, _ := dnano.ToNanosenseConf(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	database.UpdateNanoConfig(s.db, *cfg)
	//Get corresponding switchMac
	nano, _ := database.GetNanoConfig(s.db, cfg.Label)
	if nano == nil {
		rlog.Error("Cannot find config for " + cfg.Label)
		return
	}

	s.sendSwitchNanoSetup(*nano)
}

func (s *CoreService) updateNanoSetup(config interface{}) {
	cfg, _ := dnano.ToNanosenseSetup(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}
	database.UpdateNanoSetup(s.db, *cfg)
	s.sendSwitchNanoSetup(*cfg)
}

func (s *CoreService) updateNanoLabelSetup(config interface{}) {
	cfg, _ := dnano.ToNanosenseSetup(config)
	if cfg == nil || cfg.Label == "" {
		rlog.Error("Cannot parse ")
		return
	}
	cfg.Mac = strings.ToUpper(cfg.Mac)

	old, _ := database.GetNanoLabelConfig(s.db, cfg.Label)
	if old != nil {
		s.updateGroupNano(*old, *cfg)
	}

	database.UpdateNanoLabelSetup(s.db, *cfg)
	//Get corresponding switchMac
	new, _ := database.GetNanoLabelConfig(s.db, cfg.Label)
	if new == nil {
		rlog.Error("Cannot find config for " + cfg.Label)
		return
	}
	rlog.Info("Nano configuration " + cfg.Label + " saved")
	s.sendSwitchNanoSetup(*cfg)
}

func (s *CoreService) updateGroupNano(old dnano.NanosenseSetup, cfg dnano.NanosenseSetup) {
	if old.Group != cfg.Group {
		rlog.Info("Update old group", old.Group)
		gr, _ := database.GetGroupConfig(s.db, old.Group)
		if gr != nil {
			nanos := []string{}
			for _, v := range gr.Nanosenses {
				if v != cfg.Mac {
					nanos = append(nanos, v)
				}
			}
			gr.Nanosenses = nanos
			rlog.Info("Old group will be ", gr.Nanosenses)
			s.updateGroupCfg(gr)
		}
		rlog.Info("Update new group", cfg.Group)
		grNew, _ := database.GetGroupConfig(s.db, cfg.Group)
		if grNew != nil && cfg.Mac != "" {
			if inArray(cfg.Mac, grNew.Nanosenses) {
				return
			}
			grNew.Nanosenses = append(grNew.Nanosenses, cfg.Mac)
			rlog.Info("new group will be", grNew.Leds)
			s.updateGroupCfg(grNew)
		}
	}
}
