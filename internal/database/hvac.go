package database

import (
	"github.com/energieip/common-components-go/pkg/dhvac"
	"github.com/energieip/common-components-go/pkg/pconst"
)

//SaveHvacConfig dump hvac config in database
func SaveHvacConfig(db Database, cfg dhvac.HvacSetup) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Mac
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbHvacs, criteria)
}

//SaveHvacLabelConfig dump hvac config in database
func SaveHvacLabelConfig(db Database, cfg dhvac.HvacSetup) error {
	criteria := make(map[string]interface{})
	if cfg.Label == nil {
		return NewError("Device " + cfg.Mac + "not found")
	}
	criteria["Label"] = *cfg.Label
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbHvacs, criteria)
}

//UpdateHvacConfig update hvac config in database
func UpdateHvacConfig(db Database, cfg dhvac.HvacConf) error {
	setup, dbID := GetHvacConfig(db, cfg.Mac)
	if setup == nil || dbID == "" {
		return NewError("Device " + cfg.Mac + " not found")
	}

	new := dhvac.UpdateConfig(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbHvacs, dbID, &new)
}

//UpdateHvacLabelConfig update hvac config in database
func UpdateHvacLabelConfig(db Database, cfg dhvac.HvacConf) error {
	if cfg.Label == nil {
		return NewError("Unknow label")
	}
	setup, dbID := GetHvacLabelConfig(db, *cfg.Label)
	if setup == nil || dbID == "" {
		return NewError("Device " + *cfg.Label + " not found")
	}

	new := dhvac.UpdateConfig(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbHvacs, dbID, &new)
}

//UpdateHvacLabelSetup update hvac config in database
func UpdateHvacLabelSetup(db Database, cfg dhvac.HvacSetup) error {
	if cfg.Label == nil {
		return NewError("Device label not found")
	}
	setup, dbID := GetHvacLabelConfig(db, *cfg.Label)
	if setup == nil || dbID == "" {
		cfg = dhvac.FillDefaultValue(cfg)
		return SaveHvacLabelConfig(db, cfg)
	}
	new := dhvac.UpdateSetup(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbHvacs, dbID, &new)
}

//UpdateHvacSetup update hvac config in database
func UpdateHvacSetup(db Database, cfg dhvac.HvacSetup) error {
	setup, dbID := GetHvacConfig(db, cfg.Mac)
	if setup == nil || dbID == "" {
		cfg = dhvac.FillDefaultValue(cfg)
		return SaveHvacConfig(db, cfg)
	}
	new := dhvac.UpdateSetup(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbHvacs, dbID, &new)
}

//SwitchHvacConfig update hvac config in database
func SwitchHvacConfig(db Database, old, oldFull, new, newFull string) error {
	setup, dbID := GetHvacConfig(db, old)
	if setup == nil || dbID == "" {
		return NewError("Device " + old + " not found")
	}
	setup.FullMac = &newFull
	setup.Mac = new
	return db.UpdateRecord(pconst.DbConfig, pconst.TbHvacs, dbID, setup)
}

//RemoveHvacConfig remove hvac config in database
func RemoveHvacConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbConfig, pconst.TbHvacs, criteria)
}

//RemoveHvacStatus remove hvac status in database
func RemoveHvacStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbStatus, pconst.TbHvacs, criteria)
}

//GetHvacSwitchStatus get cluster Config list
func GetHvacSwitchStatus(db Database, swMac string) map[string]dhvac.Hvac {
	res := map[string]dhvac.Hvac{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(pconst.DbStatus, pconst.TbHvacs, criteria)
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
	stored, err := db.GetRecords(pconst.DbStatus, pconst.TbHvacs, criteria)
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
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbHvacs, criteria)
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
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbHvacs, criteria)
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
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbHvacs)
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

//GetHvacsConfigByLabel return the hvac config list
func GetHvacsConfigByLabel(db Database) map[string]dhvac.HvacSetup {
	drivers := map[string]dhvac.HvacSetup{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbHvacs)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dhvac.ToHvacSetup(s)
		if err != nil || driver == nil || driver.Label == nil {
			continue
		}
		drivers[*driver.Label] = *driver
	}
	return drivers
}

//SaveHvacStatus dump hvac status in database
func SaveHvacStatus(db Database, status dhvac.Hvac) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	return SaveOnUpdateObject(db, status, pconst.DbStatus, pconst.TbHvacs, criteria)
}

//GetHvacsStatus return the hvac status list
func GetHvacsStatus(db Database) map[string]dhvac.Hvac {
	drivers := map[string]dhvac.Hvac{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbHvacs)
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

//GetHvacsStatusByLabel return the hvac status list
func GetHvacsStatusByLabel(db Database) map[string]dhvac.Hvac {
	drivers := map[string]dhvac.Hvac{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbHvacs)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dhvac.ToHvac(s)
		if err != nil || driver == nil || driver.Label == nil {
			continue
		}
		drivers[*driver.Label] = *driver
	}
	return drivers
}

//GetHvacStatus return the hvac status
func GetHvacStatus(db Database, mac string) *dhvac.Hvac {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(pconst.DbStatus, pconst.TbHvacs, criteria)
	if err != nil || stored == nil {
		return nil
	}
	driver, err := dhvac.ToHvac(stored)
	if err != nil {
		return nil
	}
	return driver
}
