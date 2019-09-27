package service

import (
	"strings"

	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/common-components-go/pkg/dwago"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) sendSwitchWagoSetup(wago dwago.WagoSetup) {

	for mac := range database.GetCluster(s.db, wago.Cluster) {
		if mac == "" {
			continue
		}
		sw, _ := database.GetSwitchConfig(s.db, mac)
		if sw != nil {
			ip := "0"
			if sw.IP != nil {
				ip = *sw.IP
			}
			dumpFreq := 1000
			if sw.DumpFrequency != nil {
				dumpFreq = *sw.DumpFrequency
			}
			url := "/write/switch/" + mac + "/update/settings"
			switchSetup := sd.SwitchConfig{}
			switchSetup.Mac = mac
			switchSetup.IP = ip
			switchSetup.DumpFrequency = dumpFreq
			switchSetup.WagosSetup = make(map[string]dwago.WagoSetup)
			switchSetup.WagosSetup[wago.Mac] = wago
			dump, _ := switchSetup.ToJSON()
			s.server.SendCommand(url, dump)
		}
	}
}

func (s *CoreService) updateWagoCfg(config interface{}) {
	cfg, _ := dwago.ToWagoConf(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	var wago *dwago.WagoSetup
	if cfg.Mac != "" {
		database.UpdateWagoConfig(s.db, *cfg)
		//Get corresponding switchMac
		wago, _ := database.GetWagoConfig(s.db, cfg.Mac)
		if wago == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}
	} else {
		if cfg.Label == nil {
			rlog.Error("Cannot find config for " + cfg.Mac)
			return
		}
		database.UpdateWagoLabelConfig(s.db, *cfg)
		//Get corresponding switchMac
		wago, _ = database.GetWagoLabelConfig(s.db, *cfg.Label)
		if wago == nil {
			rlog.Error("Cannot find config for " + *cfg.Label)
			return
		}
	}

	s.sendSwitchWagoSetup(*wago)
}

func (s *CoreService) updateWagoSetup(config interface{}) {
	byLbl := false
	cfg, _ := dwago.ToWagoSetup(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	oldWago, _ := database.GetWagoConfig(s.db, cfg.Mac)
	if oldWago == nil && cfg.Label != nil {
		oldWago, _ = database.GetWagoLabelConfig(s.db, *cfg.Label)
		if oldWago != nil {
			//it means that the IFC has been uploaded but the MAC is unknown
			byLbl = true
		}
	}

	if byLbl {
		database.UpdateWagoLabelSetup(s.db, *cfg)
	} else {
		database.UpdateWagoSetup(s.db, *cfg)
	}
	s.sendSwitchWagoSetup(*cfg)
}

func (s *CoreService) updateWagoLabelSetup(config interface{}) {
	cfg, _ := dwago.ToWagoSetup(config)
	if cfg == nil || cfg.Label == nil {
		rlog.Error("Cannot parse ")
		return
	}
	cfg.Mac = strings.ToUpper(cfg.Mac)

	database.UpdateWagoLabelSetup(s.db, *cfg)
	s.sendSwitchWagoSetup(*cfg)
}
