package database

import (
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	"github.com/energieip/common-components-go/pkg/pconst"
)

//SaveSensorConfig dump sensor config in database
func SaveSensorConfig(db Database, cfg ds.SensorSetup) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Mac
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbSensors, criteria)
}

//SaveSensorLabelConfig dump sensor config in database
func SaveSensorLabelConfig(db Database, cfg ds.SensorSetup) error {
	criteria := make(map[string]interface{})
	if cfg.Label == nil {
		return NewError("Device " + cfg.Mac + "not found")
	}
	criteria["Label"] = *cfg.Label
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbSensors, criteria)
}

//UpdateSensorConfig update sensor config in database
func UpdateSensorConfig(db Database, cfg ds.SensorConf) error {
	setup, dbID := GetSensorConfig(db, cfg.Mac)
	if setup == nil || dbID == "" {
		return NewError("Device " + cfg.Mac + "not found")
	}

	new := ds.UpdateConfig(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbSensors, dbID, &new)
}

//UpdateSensorSetup update sensor config in database
func UpdateSensorSetup(db Database, cfg ds.SensorSetup) error {
	setup, dbID := GetSensorConfig(db, cfg.Mac)
	if setup == nil || dbID == "" {
		cfg = ds.FillDefaultValue(cfg)
		return SaveSensorConfig(db, cfg)
	}
	new := ds.UpdateSetup(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbSensors, dbID, &new)
}

//UpdateSensorLabelSetup update sensor config in database
func UpdateSensorLabelSetup(db Database, cfg ds.SensorSetup) error {
	if cfg.Label == nil {
		return NewError("Device " + cfg.Mac + "not found")
	}
	setup, dbID := GetSensorLabelConfig(db, *cfg.Label)
	if setup == nil || dbID == "" {
		cfg = ds.FillDefaultValue(cfg)
		return SaveSensorLabelConfig(db, cfg)
	}
	new := ds.UpdateSetup(cfg, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbSensors, dbID, &new)
}

//RemoveSensorConfig remove sensor config in database
func RemoveSensorConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbConfig, pconst.TbSensors, criteria)
}

//RemoveSensorStatus remove led status in database
func RemoveSensorStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbStatus, pconst.TbSensors, criteria)
}

//GetSensorSwitchStatus get cluster Config list
func GetSensorSwitchStatus(db Database, swMac string) map[string]ds.Sensor {
	res := map[string]ds.Sensor{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(pconst.DbStatus, pconst.TbSensors, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := ds.ToSensor(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
}

//GetSensorSwitchSetup get sensor Config list
func GetSensorSwitchSetup(db Database, swMac string) map[string]ds.SensorSetup {
	res := map[string]ds.SensorSetup{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(pconst.DbStatus, pconst.TbSensors, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := ds.ToSensorSetup(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
}

//GetSensorConfig return the sensor configuration
func GetSensorConfig(db Database, mac string) (*ds.SensorSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbSensors, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := ds.ToSensorSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//GetSensorLabelConfig return the sensor configuration
func GetSensorLabelConfig(db Database, label string) (*ds.SensorSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbSensors, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := ds.ToSensorSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//SwitchSensorConfig update sensor config in database
func SwitchSensorConfig(db Database, old, oldFull, new, newFull string) error {
	setup, dbID := GetSensorConfig(db, old)
	if setup == nil || dbID == "" {
		return NewError("Device " + old + "not found")
	}
	setup.FullMac = newFull
	setup.Mac = new
	return db.UpdateRecord(pconst.DbConfig, pconst.TbSensors, dbID, setup)
}

//GetSensorsConfig return the sensor config list
func GetSensorsConfig(db Database) map[string]ds.SensorSetup {
	drivers := map[string]ds.SensorSetup{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbSensors)
	if err != nil || stored == nil {
		return drivers
	}
	for _, l := range stored {
		driver, err := ds.ToSensorSetup(l)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Mac] = *driver
	}
	return drivers
}

//SaveSensorStatus dump sensor status in database
func SaveSensorStatus(db Database, status ds.Sensor) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	return SaveOnUpdateObject(db, status, pconst.DbStatus, pconst.TbSensors, criteria)
}

//GetSensorsStatus return the led status list
func GetSensorsStatus(db Database) map[string]ds.Sensor {
	drivers := map[string]ds.Sensor{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbSensors)
	if err != nil || stored == nil {
		return drivers
	}
	for _, l := range stored {
		driver, err := ds.ToSensor(l)
		if err != nil || driver == nil {
			continue
		}
		drivers[driver.Mac] = *driver
	}
	return drivers
}

//GetSensorStatus return the led status
func GetSensorStatus(db Database, mac string) *ds.Sensor {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(pconst.DbStatus, pconst.TbSensors, criteria)
	if err != nil || stored == nil {
		return nil
	}
	driver, err := ds.ToSensor(stored)
	if err != nil {
		return nil
	}
	return driver
}
