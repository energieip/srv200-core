package database

import sensor "github.com/energieip/common-sensor-go/pkg/driversensor"

//SaveSensorConfig dump sensor config in database
func SaveSensorConfig(db Database, sensorStatus sensor.SensorSetup) error {
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

//GetSensorConfig return the sensor configuration
func GetSensorConfig(db Database, mac string) *sensor.SensorSetup {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	sensorStored, err := db.GetRecord(ConfigDB, SensorsTable, criteria)
	if err != nil || sensorStored == nil {
		return nil
	}
	cell, err := sensor.ToSensorSetup(sensorStored)
	if err != nil {
		return nil
	}
	return cell
}

//SaveSensorStatus dump sensor status in database
func SaveSensorStatus(db Database, sensorStatus sensor.Sensor) error {
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
