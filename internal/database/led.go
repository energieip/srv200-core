package database

import (
	dl "github.com/energieip/common-led-go/pkg/driverled"
)

//SaveLedConfig dump led config in database
func SaveLedConfig(db Database, ledStatus dl.LedSetup) error {
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
func GetLedConfig(db Database, mac string) *dl.LedSetup {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	ledStored, err := db.GetRecord(ConfigDB, LedsTable, criteria)
	if err != nil || ledStored == nil {
		return nil
	}
	light, err := dl.ToLedSetup(ledStored)
	if err != nil {
		return nil
	}
	return light
}

//UpdateLedConfig update led config in database
func UpdateLedConfig(db Database, config dl.LedConf) error {
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

	setup, err := dl.ToLedSetup(stored)
	if err != nil || stored == nil {
		return NewError("Device " + config.Mac + "not found")
	}

	if config.ThresholdHigh != nil {
		setup.ThresholdHigh = config.ThresholdHigh
	}

	if config.ThresholdLow != nil {
		setup.ThresholdLow = config.ThresholdLow
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

	if config.DumpFrequency != nil {
		setup.DumpFrequency = *config.DumpFrequency
	}

	if config.Watchdog != nil {
		setup.Watchdog = config.Watchdog
	}

	return db.UpdateRecord(ConfigDB, LedsTable, dbID, setup)
}

//GetLedsConfig return the led config list
func GetLedsConfig(db Database) map[string]dl.LedSetup {
	leds := map[string]dl.LedSetup{}
	stored, err := db.FetchAllRecords(ConfigDB, LedsTable)
	if err != nil || stored == nil {
		return leds
	}
	for _, l := range stored {
		light, err := dl.ToLedSetup(l)
		if err != nil || light == nil {
			continue
		}
		leds[light.Mac] = *light
	}
	return leds
}

//SaveLedStatus dump led status in database
func SaveLedStatus(db Database, ledStatus dl.Led) error {
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
func GetLedsStatus(db Database) map[string]dl.Led {
	leds := map[string]dl.Led{}
	stored, err := db.FetchAllRecords(StatusDB, LedsTable)
	if err != nil || stored == nil {
		return leds
	}
	for _, l := range stored {
		light, err := dl.ToLed(l)
		if err != nil || light == nil {
			continue
		}
		leds[light.Mac] = *light
	}
	return leds
}

//GetLedStatus return the led status
func GetLedStatus(db Database, mac string) *dl.Led {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	ledStored, err := db.GetRecord(StatusDB, LedsTable, criteria)
	if err != nil || ledStored == nil {
		return nil
	}
	light, err := dl.ToLed(ledStored)
	if err != nil {
		return nil
	}
	return light
}
