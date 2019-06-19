package database

import (
	"github.com/energieip/common-components-go/pkg/dhvac"
)

//SaveHvacConfig dump hvac config in database
func SaveHvacConfig(db Database, cfg dhvac.HvacSetup) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Mac
	return SaveOnUpdateObject(db, cfg, ConfigDB, HvacsTable, criteria)
}

//SaveHvacLabelConfig dump hvac config in database
func SaveHvacLabelConfig(db Database, cfg dhvac.HvacSetup) error {
	criteria := make(map[string]interface{})
	if cfg.Label == nil {
		return NewError("Device " + cfg.Mac + "not found")
	}
	criteria["Label"] = *cfg.Label
	return SaveOnUpdateObject(db, cfg, ConfigDB, HvacsTable, criteria)
}

//UpdateHvacConfig update hvac config in database
func UpdateHvacConfig(db Database, cfg dhvac.HvacConf) error {
	setup, dbID := GetHvacConfig(db, cfg.Mac)
	if setup == nil || dbID == "" {
		return NewError("Device " + cfg.Mac + " not found")
	}

	if cfg.FriendlyName != nil {
		setup.FriendlyName = cfg.FriendlyName
	}

	if cfg.Group != nil {
		setup.Group = cfg.Group
	}

	if cfg.DumpFrequency != nil {
		setup.DumpFrequency = *cfg.DumpFrequency
	}

	if cfg.Label != nil {
		setup.Label = cfg.Label
	}

	return db.UpdateRecord(ConfigDB, HvacsTable, dbID, setup)
}

//UpdateHvacLabelSetup update hvac config in database
func UpdateHvacLabelSetup(db Database, cfg dhvac.HvacSetup) error {
	if cfg.Label == nil {
		return NewError("Device label not found")
	}
	setup, dbID := GetHvacLabelConfig(db, *cfg.Label)
	if setup == nil || dbID == "" {
		if cfg.FriendlyName == nil && cfg.Label != nil {
			name := *cfg.Label
			cfg.FriendlyName = &name
		}
		if cfg.Group != nil {
			group := 0
			setup.Group = &group
		}
		if cfg.DumpFrequency == 0 {
			setup.DumpFrequency = 1000
		}
		return SaveHvacLabelConfig(db, cfg)
	}

	if cfg.FriendlyName != nil {
		setup.FriendlyName = cfg.FriendlyName
	}

	if cfg.Group != nil {
		setup.Group = cfg.Group
	}

	if cfg.DumpFrequency != 0 {
		setup.DumpFrequency = cfg.DumpFrequency
	}

	if cfg.Label != nil {
		setup.Label = cfg.Label
	}

	return db.UpdateRecord(ConfigDB, HvacsTable, dbID, setup)
}

//UpdateHvacSetup update hvac config in database
func UpdateHvacSetup(db Database, cfg dhvac.HvacSetup) error {
	setup, dbID := GetHvacConfig(db, cfg.Mac)
	if setup == nil || dbID == "" {
		if cfg.FriendlyName == nil && cfg.Label != nil {
			name := *cfg.Label
			cfg.FriendlyName = &name
		}
		if cfg.Group != nil {
			group := 0
			setup.Group = &group
		}
		if cfg.DumpFrequency == 0 {
			setup.DumpFrequency = 1000
		}
		return SaveHvacConfig(db, cfg)
	}

	if cfg.FriendlyName != nil {
		setup.FriendlyName = cfg.FriendlyName
	}

	if cfg.Group != nil {
		setup.Group = cfg.Group
	}

	if cfg.DumpFrequency != 0 {
		setup.DumpFrequency = cfg.DumpFrequency
	}

	if cfg.Label != nil {
		setup.Label = cfg.Label
	}

	return db.UpdateRecord(ConfigDB, HvacsTable, dbID, setup)
}

//SwitchHvacConfig update hvac config in database
func SwitchHvacConfig(db Database, old, oldFull, new, newFull string) error {
	setup, dbID := GetHvacConfig(db, old)
	if setup == nil || dbID == "" {
		return NewError("Device " + old + " not found")
	}
	setup.FullMac = newFull
	setup.Mac = new
	return db.UpdateRecord(ConfigDB, HvacsTable, dbID, setup)
}

//RemoveHvacConfig remove hvac config in database
func RemoveHvacConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(ConfigDB, HvacsTable, criteria)
}

//RemoveHvacStatus remove hvac status in database
func RemoveHvacStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(StatusDB, HvacsTable, criteria)
}

//GetHvacSwitchStatus get cluster Config list
func GetHvacSwitchStatus(db Database, swMac string) map[string]dhvac.Hvac {
	res := map[string]dhvac.Hvac{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(StatusDB, HvacsTable, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := dhvac.ToHvac(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
}

//GetHvacSwitchSetup get hvac Config list
func GetHvacSwitchSetup(db Database, swMac string) map[string]dhvac.HvacSetup {
	res := map[string]dhvac.HvacSetup{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(StatusDB, HvacsTable, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := dhvac.ToHvacSetup(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
}

//GetHvacConfig return the sensor configuration
func GetHvacConfig(db Database, mac string) (*dhvac.HvacSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(ConfigDB, HvacsTable, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := dhvac.ToHvacSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//GetHvacLabelConfig return the sensor configuration
func GetHvacLabelConfig(db Database, label string) (*dhvac.HvacSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(ConfigDB, HvacsTable, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := dhvac.ToHvacSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//GetHvacsConfig return the hvac config list
func GetHvacsConfig(db Database) map[string]dhvac.HvacSetup {
	drivers := map[string]dhvac.HvacSetup{}
	stored, err := db.FetchAllRecords(ConfigDB, HvacsTable)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dhvac.ToHvacSetup(s)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Mac] = *driver
	}
	return drivers
}

//SaveHvacStatus dump hvac status in database
func SaveHvacStatus(db Database, status dhvac.Hvac) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	return SaveOnUpdateObject(db, status, StatusDB, HvacsTable, criteria)
}

//GetHvacsStatus return the hvac status list
func GetHvacsStatus(db Database) map[string]dhvac.Hvac {
	drivers := map[string]dhvac.Hvac{}
	stored, err := db.FetchAllRecords(StatusDB, HvacsTable)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dhvac.ToHvac(s)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Mac] = *driver
	}
	return drivers
}

//GetHvacStatus return the hvac status
func GetHvacStatus(db Database, mac string) *dhvac.Hvac {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(StatusDB, HvacsTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	driver, err := dhvac.ToHvac(stored)
	if err != nil {
		return nil
	}
	return driver
}
