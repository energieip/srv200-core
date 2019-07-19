package database

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//GetDrivers return the project configuration
func GetDrivers(db Database) map[string]core.Driver {
	res := map[string]core.Driver{}

	projects := GetProjects(db)
	models := GetModels(db)

	for _, proj := range projects {
		mac := ""
		key := ""
		if proj.ModelName == nil {
			continue
		}
		model, ok := models[*proj.ModelName]
		if !ok {
			continue
		}
		if proj.Mac != nil {
			mac = *proj.Mac
		}
		if mac == "" {
			//use label : no driver installed
			key = proj.Label

		} else {
			key = *proj.Mac
		}
		res[key] = core.Driver{
			Mac:   mac,
			Label: proj.Label,
			Type:  model.DeviceType,
		}
	}

	leds := GetLedsStatus(db)
	for _, driv := range leds {
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    driv.Mac,
				Active: true,
				Type:   "led",
			}
		}
	}

	blinds := GetBlindsStatus(db)
	for _, driv := range blinds {
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    driv.Mac,
				Active: true,
				Type:   "blind",
			}
		}
	}

	sensors := GetBlindsStatus(db)
	for _, driv := range sensors {
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    driv.Mac,
				Active: true,
				Type:   "sensor",
			}
		}
	}

	hvacs := GetBlindsStatus(db)
	for _, driv := range hvacs {
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    driv.Mac,
				Active: true,
				Type:   "hvac",
			}
		}
	}

	wagos := GetWagosStatus(db)
	for _, driv := range wagos {
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    driv.Mac,
				Active: true,
				Type:   "wago",
			}
		}
	}

	switchs := GetSwitchsDump(db)
	for _, driv := range switchs {
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    driv.Mac,
				Active: true,
				Type:   "switch",
			}
		}
	}

	return res
}
