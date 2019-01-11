package database

import (
	ds "github.com/energieip/common-components-go/pkg/dsensor"
)

//SaveSensorConfig dump sensor config in database
func SaveSensorConfig(db Database, sensorStatus ds.SensorSetup) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = sensorStatus.Mac
	sensorStored, err := db.GetRecord(ConfigDB, SensorsTable, criteria)
	if err == nil && sensorStored != nil {
		m := sensorStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, SensorsTable, sensorStatus)
	} else {
		err = db.UpdateRecord(ConfigDB, SensorsTable, dbID, sensorStatus)
	}
	return err
}

//UpdateSensorConfig update sensor config in database
func UpdateSensorConfig(db Database, sensorConfig ds.SensorConf) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = sensorConfig.Mac
	sensorStored, err := db.GetRecord(ConfigDB, SensorsTable, criteria)
	if err != nil || sensorStored == nil {
		return NewError("Device " + sensorConfig.Mac + "not found")
	}
	m := sensorStored.(map[string]interface{})
	id, ok := m["id"]
	if !ok {
		id, ok = m["ID"]
	}
	if !ok {
		return NewError("Device " + sensorConfig.Mac + "not found")
	}
	dbID := id.(string)

	sensorSetup, err := ds.ToSensorSetup(sensorStored)
	if err != nil || sensorSetup == nil {
		return NewError("Device " + sensorConfig.Mac + "not found")
	}

	if sensorConfig.BrightnessCorrectionFactor != nil {
		sensorSetup.BrightnessCorrectionFactor = sensorConfig.BrightnessCorrectionFactor
	}

	if sensorConfig.FriendlyName != nil {
		sensorSetup.FriendlyName = sensorConfig.FriendlyName
	}

	if sensorConfig.Group != nil {
		sensorSetup.Group = sensorConfig.Group
	}

	if sensorConfig.IsBleEnabled != nil {
		sensorSetup.IsBleEnabled = sensorConfig.IsBleEnabled
	}

	if sensorConfig.TemperatureOffset != nil {
		sensorSetup.TemperatureOffset = sensorConfig.TemperatureOffset
	}

	if sensorConfig.ThresholdPresence != nil {
		sensorSetup.ThresholdPresence = sensorConfig.ThresholdPresence
	}

	if sensorConfig.DumpFrequency != nil {
		sensorSetup.DumpFrequency = *sensorConfig.DumpFrequency
	}

	return db.UpdateRecord(ConfigDB, SensorsTable, dbID, sensorSetup)
}

//RemoveSensorConfig remove sensor config in database
func RemoveSensorConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(ConfigDB, SensorsTable, criteria)
}

//GetSensorConfig return the sensor configuration
func GetSensorConfig(db Database, mac string) *ds.SensorSetup {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	sensorStored, err := db.GetRecord(ConfigDB, SensorsTable, criteria)
	if err != nil || sensorStored == nil {
		return nil
	}
	cell, err := ds.ToSensorSetup(sensorStored)
	if err != nil {
		return nil
	}
	return cell
}

//GetSensorsConfig return the sensor config list
func GetSensorsConfig(db Database) map[string]ds.SensorSetup {
	sensors := map[string]ds.SensorSetup{}
	stored, err := db.FetchAllRecords(ConfigDB, SensorsTable)
	if err != nil || stored == nil {
		return sensors
	}
	for _, l := range stored {
		cell, err := ds.ToSensorSetup(l)
		if err != nil || cell == nil {
			continue
		}
		sensors[cell.Mac] = *cell
	}
	return sensors
}

//SaveSensorStatus dump sensor status in database
func SaveSensorStatus(db Database, sensorStatus ds.Sensor) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = sensorStatus.Mac
	sensorStored, err := db.GetRecord(StatusDB, SensorsTable, criteria)
	if err == nil && sensorStored != nil {
		m := sensorStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(StatusDB, SensorsTable, sensorStatus)
	} else {
		err = db.UpdateRecord(StatusDB, SensorsTable, dbID, sensorStatus)
	}
	return err
}

//GetSensorsStatus return the led status list
func GetSensorsStatus(db Database) map[string]ds.Sensor {
	sensors := map[string]ds.Sensor{}
	stored, err := db.FetchAllRecords(StatusDB, SensorsTable)
	if err != nil || stored == nil {
		return sensors
	}
	for _, l := range stored {
		cell, err := ds.ToSensor(l)
		if err != nil || cell == nil {
			continue
		}
		sensors[cell.Mac] = *cell
	}
	return sensors
}

//GetSensorStatus return the led status
func GetSensorStatus(db Database, mac string) *ds.Sensor {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(StatusDB, SensorsTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	cell, err := ds.ToSensor(stored)
	if err != nil {
		return nil
	}
	return cell
}
