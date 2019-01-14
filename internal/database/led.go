package database

import (
	dl "github.com/energieip/common-components-go/pkg/dled"
)

//SaveLedConfig dump led config in database
func SaveLedConfig(db Database, cfg dl.LedSetup) error {
	var err error
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Mac
	dbID := GetObjectID(db, ConfigDB, LedsTable, criteria)
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, LedsTable, cfg)
	} else {
		err = db.UpdateRecord(ConfigDB, LedsTable, dbID, cfg)
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
func GetLedConfig(db Database, mac string) (*dl.LedSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(ConfigDB, LedsTable, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := dl.ToLedSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//UpdateLedConfig update led config in database
func UpdateLedConfig(db Database, config dl.LedConf) error {
	setup, dbID := GetLedConfig(db, config.Mac)
	if setup == nil || dbID == "" {
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
func SaveLedStatus(db Database, status dl.Led) error {
	var err error
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	dbID := GetObjectID(db, StatusDB, LedsTable, criteria)
	if dbID == "" {
		_, err = db.InsertRecord(StatusDB, LedsTable, status)
	} else {
		err = db.UpdateRecord(StatusDB, LedsTable, dbID, status)
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
