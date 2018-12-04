package database

import (
	led "github.com/energieip/common-led-go/pkg/driverled"
)

//SaveLedConfig dump led config in database
func SaveLedConfig(db Database, ledStatus led.LedSetup) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = ledStatus.Mac
	ledStored, err := db.GetRecord(ConfigDB, LedsTable, criteria)
	if err == nil && ledStored != nil {
		m := ledStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, LedsTable, ledStatus)
	} else {
		err = db.UpdateRecord(ConfigDB, LedsTable, dbID, ledStatus)
	}
	return err
}

//GetLedConfig return the led configuration
func GetLedConfig(db Database, mac string) *led.LedSetup {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	ledStored, err := db.GetRecord(ConfigDB, LedsTable, criteria)
	if err != nil || ledStored == nil {
		return nil
	}
	light, err := led.ToLedSetup(ledStored)
	if err != nil {
		return nil
	}
	return light
}

//SaveLedStatus dump led status in database
func SaveLedStatus(db Database, ledStatus led.Led) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = ledStatus.Mac
	ledStored, err := db.GetRecord(StatusDB, LedsTable, criteria)
	if err == nil && ledStored != nil {
		m := ledStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(StatusDB, LedsTable, ledStatus)
	} else {
		err = db.UpdateRecord(StatusDB, LedsTable, dbID, ledStatus)
	}
	return err
}
