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

//RemoveLedConfig remove led config in database
func RemoveLedConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(ConfigDB, LedsTable, criteria)
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

//UpdateLedConfig update led config in database
func UpdateLedConfig(db Database, config led.LedConf) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = config.Mac
	stored, err := db.GetRecord(ConfigDB, LedsTable, criteria)
	if err != nil || stored == nil {
		return NewError("Device " + config.Mac + "not found")
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if !ok {
		id, ok = m["ID"]
	}
	if !ok {
		return NewError("Device " + config.Mac + "not found")
	}
	dbID := id.(string)

	setup, err := led.ToLedSetup(stored)
	if err != nil || stored == nil {
		return NewError("Device " + config.Mac + "not found")
	}

	if config.ThresoldHigh != nil {
		setup.ThresoldHigh = config.ThresoldHigh
	}

	if config.FriendlyName != nil {
		setup.FriendlyName = config.FriendlyName
	}

	if config.Group != nil {
		setup.Group = config.Group
	}

	if config.IsBleEnabled != nil {
		setup.IsBleEnabled = config.IsBleEnabled
	}

	return db.UpdateRecord(ConfigDB, LedsTable, dbID, setup)
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

//GetLedsStatus return the led status list
func GetLedsStatus(db Database) map[string]led.Led {
	leds := map[string]led.Led{}
	stored, err := db.FetchAllRecords(StatusDB, LedsTable)
	if err != nil || stored == nil {
		return leds
	}
	for _, l := range stored {
		light, err := led.ToLed(l)
		if err != nil || light == nil {
			continue
		}
		leds[light.Mac] = *light
	}
	return leds
}

//GetLedStatus return the led status
func GetLedStatus(db Database, mac string) *led.Led {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	ledStored, err := db.GetRecord(StatusDB, LedsTable, criteria)
	if err != nil || ledStored == nil {
		return nil
	}
	light, err := led.ToLed(ledStored)
	if err != nil {
		return nil
	}
	return light
}
