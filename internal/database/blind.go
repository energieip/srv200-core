package database

import (
	"github.com/energieip/common-components-go/pkg/dblind"
	"github.com/energieip/common-components-go/pkg/pconst"
)

//SaveBlindConfig dump blind config in database
func SaveBlindConfig(db Database, cfg dblind.BlindSetup) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Mac
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbBlinds, criteria)
}

//SaveBlindLabelConfig dump blind config in database
func SaveBlindLabelConfig(db Database, cfg dblind.BlindSetup) error {
	criteria := make(map[string]interface{})
	if cfg.Label == nil {
		return NewError("Device " + cfg.Mac + "not found")
	}
	criteria["Label"] = *cfg.Label
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbBlinds, criteria)
}

//RemoveSwitchBlindStatus remove led status in database
func RemoveSwitchBlindStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = mac
	return db.DeleteRecord(pconst.DbStatus, pconst.TbBlinds, criteria)
}

//UpdateBlindConfig update blind config in database
func UpdateBlindConfig(db Database, cfg dblind.BlindConf) error {
	setup, dbID := GetBlindConfig(db, cfg.Mac)
	if setup == nil || dbID == "" {
		return NewError("Device " + cfg.Mac + " not found")
	}

	new := dblind.UpdateConfig(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbBlinds, dbID, &new)
}

//UpdateBlindLabelConfig update blind config in database
func UpdateBlindLabelConfig(db Database, cfg dblind.BlindConf) error {
	if cfg.Label == nil {
		return NewError("Unknow Device")
	}
	setup, dbID := GetBlindLabelConfig(db, *cfg.Label)
	if setup == nil || dbID == "" {
		return NewError("Device " + *cfg.Label + " not found")
	}

	new := dblind.UpdateConfig(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbBlinds, dbID, &new)
}

//UpdateBlindLabelSetup update blind config in database
func UpdateBlindLabelSetup(db Database, cfg dblind.BlindSetup) error {
	if cfg.Label == nil {
		return NewError("Device label not found")
	}
	setup, dbID := GetBlindLabelConfig(db, *cfg.Label)
	if setup == nil || dbID == "" {
		cfg = dblind.FillDefaultValue(cfg)
		return SaveBlindLabelConfig(db, cfg)
	}
	new := dblind.UpdateSetup(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbBlinds, dbID, &new)
}

//UpdateBlindSetup update blind config in database
func UpdateBlindSetup(db Database, cfg dblind.BlindSetup) error {
	setup, dbID := GetBlindConfig(db, cfg.Mac)
	if setup == nil || dbID == "" {
		cfg = dblind.FillDefaultValue(cfg)
		return SaveBlindConfig(db, cfg)
	}

	new := dblind.UpdateSetup(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbBlinds, dbID, &new)
}

//SwitchBlindConfig update blind config in database
func SwitchBlindConfig(db Database, old, new string) error {
	setup, dbID := GetBlindConfig(db, old)
	if setup == nil || dbID == "" {
		return NewError("Device " + old + " not found")
	}
	setup.Mac = new
	return db.UpdateRecord(pconst.DbConfig, pconst.TbBlinds, dbID, setup)
}

//RemoveBlindConfig remove blind config in database
func RemoveBlindConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbConfig, pconst.TbBlinds, criteria)
}

//RemoveBlindStatus remove led status in database
func RemoveBlindStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbStatus, pconst.TbBlinds, criteria)
}

//GetBlindSwitchStatus get cluster Config list
func GetBlindSwitchStatus(db Database, swMac string) map[string]dblind.Blind {
	res := map[string]dblind.Blind{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(pconst.DbStatus, pconst.TbBlinds, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := dblind.ToBlind(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
}

//GetBlindSwitchSetup get blind Config list
func GetBlindSwitchSetup(db Database, swMac string) map[string]dblind.BlindSetup {
	res := map[string]dblind.BlindSetup{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(pconst.DbConfig, pconst.TbBlinds, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := dblind.ToBlindSetup(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
}

//GetBlindConfig return the sensor configuration
func GetBlindConfig(db Database, mac string) (*dblind.BlindSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbBlinds, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := dblind.ToBlindSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//GetBlindLabelConfig return the sensor configuration
func GetBlindLabelConfig(db Database, label string) (*dblind.BlindSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbBlinds, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := dblind.ToBlindSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//GetBlindsConfig return the blind config list
func GetBlindsConfig(db Database) map[string]dblind.BlindSetup {
	drivers := map[string]dblind.BlindSetup{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbBlinds)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dblind.ToBlindSetup(s)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Mac] = *driver
	}
	return drivers
}

//GetBlindsConfigByLabel return the blind config list
func GetBlindsConfigByLabel(db Database) map[string]dblind.BlindSetup {
	drivers := map[string]dblind.BlindSetup{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbBlinds)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dblind.ToBlindSetup(s)
		if err != nil || driver == nil || driver.Label == nil {
			continue
		}
		drivers[*driver.Label] = *driver
	}
	return drivers
}

//SaveBlindStatus dump blind status in database
func SaveBlindStatus(db Database, status dblind.Blind) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	return SaveOnUpdateObject(db, status, pconst.DbStatus, pconst.TbBlinds, criteria)
}

//GetBlindsStatus return the blind status list
func GetBlindsStatus(db Database) map[string]dblind.Blind {
	drivers := map[string]dblind.Blind{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbBlinds)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dblind.ToBlind(s)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Mac] = *driver
	}
	return drivers
}

//GetBlindsStatusByLabel return the blind status list
func GetBlindsStatusByLabel(db Database) map[string]dblind.Blind {
	drivers := map[string]dblind.Blind{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbBlinds)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dblind.ToBlind(s)
		if err != nil || driver == nil || driver.Label == nil {
			continue
		}
		drivers[*driver.Label] = *driver
	}
	return drivers
}

//GetBlindStatus return the blind status
func GetBlindStatus(db Database, mac string) *dblind.Blind {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(pconst.DbStatus, pconst.TbBlinds, criteria)
	if err != nil || stored == nil {
		return nil
	}
	driver, err := dblind.ToBlind(stored)
	if err != nil {
		return nil
	}
	return driver
}
