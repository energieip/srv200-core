package service

import (
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
)

func (s *CoreService) updateMapInfo(config interface{}) {
	cfg, _ := core.ToMapInfo(config)
	if cfg == nil {
		s.uploadValue = "failure"
		return
	}

	//association[label] = mac
	association := make(map[string]string)

	//clean old configuration except the project table (already associate qrcode)
	database.PrepareDB(s.db, true)

	for _, proj := range cfg.Project {
		proj.CommissioningDate = nil
		driver, _ := database.GetProject(s.db, proj.Label)
		if driver != nil {
			//case the project was already existing check if the device type is changed.
			if driver.ModelName != nil && proj.ModelName != nil {
				if *driver.ModelName != *proj.ModelName {
					//special case where the device has changed for a given cable so a new fresh install need to be done
					mac := ""
					proj.Mac = &mac
				} else {
					if driver.Mac != nil {
						association[proj.Label] = *driver.Mac
					}
				}
			}
		}
		database.SaveProject(s.db, proj)
	}

	for _, md := range cfg.Models {
		database.SaveModel(s.db, md)
	}

	for _, sw := range cfg.Switchs {
		if sw.Label != nil {
			mac, ok := association[*sw.Label]
			if ok {
				sw.Mac = &mac
			}
			s.createSwitchLabelCfg(sw)
		}
	}

	for _, md := range cfg.Frames {
		database.SaveFrame(s.db, md)
	}

	for _, gr := range cfg.Groups {
		s.createGroup(&gr)
	}

	for _, dr := range cfg.Leds {
		if dr.Label != nil {
			mac, ok := association[*dr.Label]
			if ok {
				dr.Mac = mac
			}
			s.createLedLabelSetup(dr)
		}
		//update group config
		if dr.Mac != "" {
			grNew, _ := database.GetGroupConfig(s.db, *dr.Group)
			if grNew != nil {
				if !inArray(dr.Mac, grNew.Leds) {
					grNew.Leds = append(grNew.Leds, dr.Mac)
				}
				firsts := []string{}
				for _, v := range grNew.FirstDay {
					if v != dr.Mac {
						firsts = append(firsts, v)
					}
				}
				if dr.FirstDay != nil && *dr.FirstDay == true {
					firsts = append(firsts, dr.Mac)
				}
				grNew.FirstDay = firsts
				database.UpdateGroupConfig(s.db, *grNew)
			}
		}
	}

	for _, dr := range cfg.Sensors {
		if dr.Label != nil {
			mac, ok := association[*dr.Label]
			if ok {
				dr.Mac = mac
			}
			s.createSensorLabelSetup(dr)
		}
		if dr.Mac != "" {
			grNew, _ := database.GetGroupConfig(s.db, *dr.Group)
			if grNew != nil {
				if !inArray(dr.Mac, grNew.Sensors) {
					grNew.Sensors = append(grNew.Sensors, dr.Mac)
					database.UpdateGroupConfig(s.db, *grNew)
				}
			}
		}
	}

	for _, dr := range cfg.Hvacs {
		if dr.Label != nil {
			mac, ok := association[*dr.Label]
			if ok {
				dr.Mac = mac
			}
			s.createHvacLabelSetup(dr)
		}
		if dr.Mac != "" {
			grNew, _ := database.GetGroupConfig(s.db, *dr.Group)
			if grNew != nil {
				if !inArray(dr.Mac, grNew.Hvacs) {
					grNew.Hvacs = append(grNew.Hvacs, dr.Mac)
					database.UpdateGroupConfig(s.db, *grNew)
				}
			}
		}
	}

	for _, dr := range cfg.Blinds {
		if dr.Label != nil {
			mac, ok := association[*dr.Label]
			if ok {
				dr.Mac = mac
			}
			s.createBlindLabelSetup(dr)
		}
		if dr.Mac != "" {
			grNew, _ := database.GetGroupConfig(s.db, *dr.Group)
			if grNew != nil {
				if !inArray(dr.Mac, grNew.Blinds) {
					grNew.Blinds = append(grNew.Blinds, dr.Mac)
					database.UpdateGroupConfig(s.db, *grNew)
				}
			}
		}
	}

	for _, dr := range cfg.Wagos {
		if dr.Label != nil {
			mac, ok := association[*dr.Label]
			if ok {
				dr.Mac = mac
			}
			s.createWagoLabelSetup(dr)
		}
	}

	for _, dr := range cfg.Nanosenses {
		if dr.Label != "" {
			mac, ok := association[dr.Label]
			if ok {
				dr.Mac = mac
			}
			s.createNanoLabelSetup(dr)
		}
		if dr.Mac != "" {
			grNew, _ := database.GetGroupConfig(s.db, dr.Group)
			if grNew != nil {
				if !inArray(dr.Mac, grNew.Nanosenses) {
					grNew.Nanosenses = append(grNew.Nanosenses, dr.Mac)
					database.UpdateGroupConfig(s.db, *grNew)
				}
			}
		}
	}

	//reset every switch to be sure that there is no cache with an old config
	switchs := database.GetSwitchsConfig(s.db)
	for sw := range switchs {
		isConfigured := false
		url := "/write/switch/" + sw + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sw
		switchSetup.IsConfigured = &isConfigured
		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}

	s.uploadValue = "success"
}
