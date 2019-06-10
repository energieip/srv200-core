package service

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
)

func (s *CoreService) updateMapInfo(config interface{}) {
	cfg, _ := core.ToMapInfo(config)
	if cfg == nil {
		return
	}

	for _, sw := range cfg.Switchs {
		s.updateSwitchCfg(sw)
	}

	for _, gr := range cfg.Groups {
		s.updateGroupCfg(gr)
	}

	for _, dr := range cfg.Leds {
		s.updateLedSetup(dr)
	}

	for _, dr := range cfg.Sensors {
		s.updateSensorSetup(dr)
	}

	for _, dr := range cfg.Hvacs {
		s.updateHvacSetup(dr)
	}

	for _, dr := range cfg.Blinds {
		s.updateBlindSetup(dr)
	}

	for _, proj := range cfg.Project {
		database.SaveProject(s.db, proj)
	}

	for _, md := range cfg.Models {
		database.SaveModel(s.db, md)
	}
}
