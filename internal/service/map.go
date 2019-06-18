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
		s.updateSwitchLabelCfg(sw)
	}

	for _, gr := range cfg.Groups {
		s.updateGroupCfg(gr)
	}

	for _, dr := range cfg.Leds {
		s.updateLedLabelSetup(dr)
	}

	for _, dr := range cfg.Sensors {
		s.updateSensorLabelSetup(dr)
	}

	for _, dr := range cfg.Hvacs {
		s.updateHvacLabelSetup(dr)
	}

	for _, dr := range cfg.Blinds {
		s.updateBlindLabelSetup(dr)
	}

	for _, proj := range cfg.Project {
		database.SaveProject(s.db, proj)
	}

	for _, md := range cfg.Models {
		database.SaveModel(s.db, md)
	}
}
