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
		if proj.FullMac != nil {
			mac = *proj.FullMac
		} else {
			if proj.Mac != nil {
				//case old driver
				mac = "00:00:00:" + *proj.Mac
			}
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
		mac := ""
		if driv.FullMac != nil {
			mac = *driv.FullMac
		} else {
			//case old driver
			mac = "00:00:00:" + driv.Mac
		}
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    mac,
				Active: true,
				Type:   "led",
			}
		}
	}

	blinds := GetBlindsStatus(db)
	for _, driv := range blinds {
		mac := ""
		if driv.FullMac != nil {
			mac = *driv.FullMac
		} else {
			//case old driver
			mac = "00:00:00:" + driv.Mac
		}
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    mac,
				Active: true,
				Type:   "blind",
			}
		}
	}

	sensors := GetBlindsStatus(db)
	for _, driv := range sensors {
		mac := ""
		if driv.FullMac != nil {
			mac = *driv.FullMac
		} else {
			//case old driver
			mac = "00:00:00:" + driv.Mac
		}
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    mac,
				Active: true,
				Type:   "sensor",
			}
		}
	}

	hvacs := GetBlindsStatus(db)
	for _, driv := range hvacs {
		mac := ""
		if driv.FullMac != nil {
			mac = *driv.FullMac
		} else {
			//case old driver
			mac = "00:00:00:" + driv.Mac
		}
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    mac,
				Active: true,
				Type:   "hvac",
			}
		}
	}

	wagos := GetWagosStatus(db)
	for _, driv := range wagos {
		mac := ""
		if driv.FullMac != nil {
			mac = *driv.FullMac
		} else {
			//case old driver
			mac = "00:00:00:" + driv.Mac
		}
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    mac,
				Active: true,
				Type:   "wago",
			}
		}
	}

	switchs := GetSwitchsDump(db)
	for _, driv := range switchs {
		mac := ""
		if driv.FullMac != "" {
			mac = driv.FullMac
		} else {
			//case old driver
			mac = "00:00:00:" + driv.Mac
		}
		d, ok := res[driv.Mac]
		if ok {
			d.Active = true
			res[driv.Mac] = d
		} else {
			res[driv.Mac] = core.Driver{
				Mac:    mac,
				Active: true,
				Type:   "switch",
			}
		}
	}

	return res
}
