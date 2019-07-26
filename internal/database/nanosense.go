package database

import (
	"github.com/energieip/common-components-go/pkg/dnanosense"
	"github.com/energieip/common-components-go/pkg/pconst"
)

//SaveNanoConfig dump nano config in database
func SaveNanoConfig(db Database, cfg dnanosense.NanosenseSetup) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Label
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbNanosenses, criteria)
}

//UpdateNanoConfig update nano config in database
func UpdateNanoConfig(db Database, cfg dnanosense.NanosenseConf) error {
	setup, dbID := GetNanoConfig(db, cfg.Label)
	if setup == nil || dbID == "" {
		return NewError("Device " + cfg.Label + " not found")
	}

	new := dnanosense.UpdateConfig(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbNanosenses, dbID, &new)
}

//SaveNanoConfig dump nano config in database
func SaveNanoLabelConfig(db Database, cfg dnanosense.NanosenseSetup) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = cfg.Label
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbNanosenses, criteria)
}

//UpdateNanoLabelConfig update nano config in database
func UpdateNanoLabelConfig(db Database, cfg dnanosense.NanosenseConf) error {
	setup, dbID := GetNanoLabelConfig(db, cfg.Label)
	if setup == nil || dbID == "" {
		return NewError("Device " + cfg.Label + " not found")
	}

	new := dnanosense.UpdateConfig(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbNanosenses, dbID, &new)
}

//UpdateNanoSetup update nano config in database
func UpdateNanoSetup(db Database, cfg dnanosense.NanosenseSetup) error {
	setup, dbID := GetNanoConfig(db, cfg.Label)
	if setup == nil || dbID == "" {
		cfg = dnanosense.FillDefaultValue(cfg)
		return SaveNanoConfig(db, cfg)
	}

	new := dnanosense.UpdateSetup(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbNanosenses, dbID, &new)
}

//RemoveNanoConfig remove nano config in database
func RemoveNanoConfig(db Database, label string) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	return db.DeleteRecord(pconst.DbConfig, pconst.TbNanosenses, criteria)
}

//RemoveNanoStatus remove nano status in database
func RemoveNanoStatus(db Database, label string) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	return db.DeleteRecord(pconst.DbStatus, pconst.TbNanosenses, criteria)
}

//GetNanoSwitchStatus get cluster Config list
func GetNanoSwitchStatus(db Database, cluster int) map[string]dnanosense.Nanosense {
	res := map[string]dnanosense.Nanosense{}
	criteria := make(map[string]interface{})
	criteria["Cluster"] = cluster
	stored, err := db.GetRecords(pconst.DbStatus, pconst.TbNanosenses, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := dnanosense.ToNanosense(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Label] = *driver
	}
	return res
}

//GetNanoSwitchSetup get nano Config list
func GetNanoSwitchSetup(db Database, swCluster int) map[string]dnanosense.NanosenseSetup {
	res := map[string]dnanosense.NanosenseSetup{}
	criteria := make(map[string]interface{})
	criteria["Cluster"] = swCluster
	stored, err := db.GetRecords(pconst.DbConfig, pconst.TbNanosenses, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := dnanosense.ToNanosenseSetup(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Label] = *driver
	}
	return res
}

//GetNanoLabelConfig return the nano configuration
func GetNanoLabelConfig(db Database, label string) (*dnanosense.NanosenseSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbNanosenses, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := dnanosense.ToNanosenseSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//GetNanoConfig return the nano configuration
func GetNanoConfig(db Database, label string) (*dnanosense.NanosenseSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = label
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbNanosenses, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := dnanosense.ToNanosenseSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//GetNanosConfig return the nano config list
func GetNanosConfig(db Database) map[string]dnanosense.NanosenseSetup {
	drivers := map[string]dnanosense.NanosenseSetup{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbNanosenses)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dnanosense.ToNanosenseSetup(s)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Label] = *driver
	}
	return drivers
}

//SaveNanoStatus dump nano status in database
func SaveNanoStatus(db Database, status dnanosense.Nanosense) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = status.Label
	return SaveOnUpdateObject(db, status, pconst.DbStatus, pconst.TbNanosenses, criteria)
}

//GetNanosStatus return the nano status list
func GetNanosStatus(db Database) map[string]dnanosense.Nanosense {
	drivers := map[string]dnanosense.Nanosense{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbNanosenses)
	if err != nil || stored == nil {
		return drivers
	}
	for _, s := range stored {
		driver, err := dnanosense.ToNanosense(s)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Mac] = *driver
	}
	return drivers
}

//GetNanoStatus return the nano status
func GetNanoStatus(db Database, label string) *dnanosense.Nanosense {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(pconst.DbStatus, pconst.TbNanosenses, criteria)
	if err != nil || stored == nil {
		return nil
	}
	driver, err := dnanosense.ToNanosense(stored)
	if err != nil {
		return nil
	}
	return driver
}

//UpdateNanoLabelSetup update led setup in database
func UpdateNanoLabelSetup(db Database, config dnanosense.NanosenseSetup) error {
	if config.Label == "" {
		return NewError("Device label not found")
	}
	setup, dbID := GetNanoLabelConfig(db, config.Label)
	if setup == nil || dbID == "" {
		config := dnanosense.FillDefaultValue(config)
		return SaveNanoLabelConfig(db, config)
	}

	new := dnanosense.UpdateSetup(config, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbNanosenses, dbID, &new)
}

//GetNanosStatusByLabel return the nano status list
func GetNanosStatusByLabel(db Database) map[string]dnanosense.Nanosense {
	drivers := map[string]dnanosense.Nanosense{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbNanosenses)
	if err != nil || stored == nil {
		return drivers
	}
	for _, l := range stored {
		driver, err := dnanosense.ToNanosense(l)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Label] = *driver
	}
	return drivers
}

//GetNanosConfigByLabel return the sensor config list
func GetNanosConfigByLabel(db Database) map[string]dnanosense.NanosenseSetup {
	drivers := map[string]dnanosense.NanosenseSetup{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbNanosenses)
	if err != nil || stored == nil {
		return drivers
	}
	for _, l := range stored {
		driver, err := dnanosense.ToNanosenseSetup(l)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Label] = *driver
	}
	return drivers
}
