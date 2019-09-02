package service

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
)

func (s *CoreService) updateMapInfo(config interface{}) {
	cfg, _ := core.ToMapInfo(config)
	if cfg == nil {
		s.uploadValue = "failure"
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

	for _, dr := range cfg.Wagos {
		s.updateWagoLabelSetup(dr)
	}

	for _, dr := range cfg.Nanosenses {
		s.updateNanoLabelSetup(dr)
	}

	for _, proj := range cfg.Project {
		proj.CommissioningDate = nil
		database.SaveProject(s.db, proj)
	}

	for _, md := range cfg.Models {
		database.SaveModel(s.db, md)
	}

	for _, md := range cfg.Frames {
		database.SaveFrame(s.db, md)
	}
	s.uploadValue = "success"
}
