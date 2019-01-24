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
