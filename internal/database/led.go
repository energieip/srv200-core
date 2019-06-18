package database

import (
	dl "github.com/energieip/common-components-go/pkg/dled"
)

//SaveLedConfig dump led config in database
func SaveLedConfig(db Database, cfg dl.LedSetup) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Mac
	return SaveOnUpdateObject(db, cfg, ConfigDB, LedsTable, criteria)
}

//SaveLedLabelConfig dump led config in database
func SaveLedLabelConfig(db Database, cfg dl.LedSetup) error {
	criteria := make(map[string]interface{})
	if cfg.Label == nil {
		return NewError("Device " + cfg.Mac + "not found")
	}
	criteria["Label"] = *cfg.Label
	return SaveOnUpdateObject(db, cfg, ConfigDB, LedsTable, criteria)
}

//RemoveLedConfig remove led config in database
func RemoveLedConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(ConfigDB, LedsTable, criteria)
}

//RemoveLedStatus remove led status in database
func RemoveLedStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(StatusDB, LedsTable, criteria)
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

//GetLedLabelConfig return the led configuration
func GetLedLabelConfig(db Database, label string) (*dl.LedSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Label"] = label
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

	if config.SlopeStartAuto != nil {
		setup.SlopeStartAuto = config.SlopeStartAuto
	}

	if config.SlopeStartManual != nil {
		setup.SlopeStartManual = config.SlopeStartManual
	}

	if config.SlopeStopAuto != nil {
		setup.SlopeStopAuto = config.SlopeStopAuto
	}

	if config.SlopeStopManual != nil {
		setup.SlopeStopManual = config.SlopeStopManual
	}

	if config.BleMode != nil {
		setup.BleMode = config.BleMode
	}

	if config.IBeaconMajor != nil {
		setup.IBeaconMajor = config.IBeaconMajor
	}

	if config.IBeaconMinor != nil {
		setup.IBeaconMinor = config.IBeaconMinor
	}

	if config.IBeaconTxPower != nil {
		setup.IBeaconTxPower = config.IBeaconTxPower
	}

	if config.IBeaconUUID != nil {
		setup.IBeaconUUID = config.IBeaconUUID
	}

	if config.DumpFrequency != nil {
		setup.DumpFrequency = *config.DumpFrequency
	}

	if config.Watchdog != nil {
		setup.Watchdog = config.Watchdog
	}

	if config.Label != nil {
		setup.Label = config.Label
	}

	return db.UpdateRecord(ConfigDB, LedsTable, dbID, setup)
}

//UpdateLedSetup update led setup in database
func UpdateLedSetup(db Database, config dl.LedSetup) error {
	setup, dbID := GetLedConfig(db, config.Mac)
	if setup == nil || dbID == "" {
		if config.IsBleEnabled == nil {
			enabled := false
			config.IsBleEnabled = &enabled
		}
		if config.DumpFrequency == 0 {
			config.DumpFrequency = 1000
		}
		if config.Auto == nil {
			auto := true
			config.Auto = &auto
		}
		if config.ThresholdHigh == nil {
			high := 100
			config.ThresholdHigh = &high
		}
		if config.ThresholdLow == nil {
			low := 0
			config.ThresholdLow = &low
		}
		if config.Group == nil {
			group := 0
			config.Group = &group
		}
		if config.Watchdog == nil {
			watchdog := 600
			config.Watchdog = &watchdog
		}
		slope := 10000
		if config.SlopeStartManual == nil {
			config.SlopeStartManual = &slope
		}
		if config.SlopeStopManual == nil {
			config.SlopeStopManual = &slope
		}
		if config.DefaultSetpoint == nil {
			defaultValue := 0
			config.DefaultSetpoint = &defaultValue
		}
		if config.BleMode == nil {
			defaultMode := "service"
			config.BleMode = &defaultMode
		}
		if config.PMax == 0 {
			config.PMax = 5
		}
		if config.FriendlyName == nil {
			name := *config.Label
			config.FriendlyName = &name
		}
		return SaveLedLabelConfig(db, config)
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

	if config.SlopeStartAuto != nil {
		setup.SlopeStartAuto = config.SlopeStartAuto
	}

	if config.SlopeStartManual != nil {
		setup.SlopeStartManual = config.SlopeStartManual
	}

	if config.SlopeStopAuto != nil {
		setup.SlopeStopAuto = config.SlopeStopAuto
	}

	if config.SlopeStopManual != nil {
		setup.SlopeStopManual = config.SlopeStopManual
	}

	if config.BleMode != nil {
		setup.BleMode = config.BleMode
	}

	if config.IBeaconMajor != nil {
		setup.IBeaconMajor = config.IBeaconMajor
	}

	if config.IBeaconMinor != nil {
		setup.IBeaconMinor = config.IBeaconMinor
	}

	if config.IBeaconTxPower != nil {
		setup.IBeaconTxPower = config.IBeaconTxPower
	}

	if config.IBeaconUUID != nil {
		setup.IBeaconUUID = config.IBeaconUUID
	}

	if config.DumpFrequency != 0 {
		setup.DumpFrequency = config.DumpFrequency
	}

	if config.Watchdog != nil {
		setup.Watchdog = config.Watchdog
	}

	if config.PMax != 0 {
		setup.PMax = config.PMax
	}

	if config.Label != nil {
		setup.Label = config.Label
	}

	return db.UpdateRecord(ConfigDB, LedsTable, dbID, setup)
}

//UpdateLedLabelSetup update led setup in database
func UpdateLedLabelSetup(db Database, config dl.LedSetup) error {
	if config.Label == nil {
		return NewError("Device label not found")
	}
	setup, dbID := GetLedLabelConfig(db, *config.Label)
	if setup == nil || dbID == "" {
		if config.IsBleEnabled == nil {
			enabled := false
			config.IsBleEnabled = &enabled
		}
		if config.DumpFrequency == 0 {
			config.DumpFrequency = 1000
		}
		if config.Auto == nil {
			auto := true
			config.Auto = &auto
		}
		if config.ThresholdHigh == nil {
			high := 100
			config.ThresholdHigh = &high
		}
		if config.ThresholdLow == nil {
			low := 0
			config.ThresholdLow = &low
		}
		if config.Group == nil {
			group := 0
			config.Group = &group
		}
		if config.Watchdog == nil {
			watchdog := 600
			config.Watchdog = &watchdog
		}
		slope := 10000
		if config.SlopeStartManual == nil {
			config.SlopeStartManual = &slope
		}
		if config.SlopeStopManual == nil {
			config.SlopeStopManual = &slope
		}
		if config.DefaultSetpoint == nil {
			defaultValue := 0
			config.DefaultSetpoint = &defaultValue
		}
		if config.BleMode == nil {
			defaultMode := "service"
			config.BleMode = &defaultMode
		}
		if config.PMax == 0 {
			config.PMax = 5
		}
		if config.FriendlyName == nil {
			name := *config.Label
			config.FriendlyName = &name
		}
		return SaveLedLabelConfig(db, config)
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

	if config.SlopeStartAuto != nil {
		setup.SlopeStartAuto = config.SlopeStartAuto
	}

	if config.SlopeStartManual != nil {
		setup.SlopeStartManual = config.SlopeStartManual
	}

	if config.SlopeStopAuto != nil {
		setup.SlopeStopAuto = config.SlopeStopAuto
	}

	if config.SlopeStopManual != nil {
		setup.SlopeStopManual = config.SlopeStopManual
	}

	if config.BleMode != nil {
		setup.BleMode = config.BleMode
	}

	if config.IBeaconMajor != nil {
		setup.IBeaconMajor = config.IBeaconMajor
	}

	if config.IBeaconMinor != nil {
		setup.IBeaconMinor = config.IBeaconMinor
	}

	if config.IBeaconTxPower != nil {
		setup.IBeaconTxPower = config.IBeaconTxPower
	}

	if config.IBeaconUUID != nil {
		setup.IBeaconUUID = config.IBeaconUUID
	}

	if config.DumpFrequency != 0 {
		setup.DumpFrequency = config.DumpFrequency
	}

	if config.Watchdog != nil {
		setup.Watchdog = config.Watchdog
	}

	if config.PMax != 0 {
		setup.PMax = config.PMax
	}

	if config.Label != nil {
		setup.Label = config.Label
	}

	return db.UpdateRecord(ConfigDB, LedsTable, dbID, setup)
}

//SwitchLedConfig update led config in database
func SwitchLedConfig(db Database, old, oldFull, new, newFull string) error {
	setup, dbID := GetLedConfig(db, old)
	if setup == nil || dbID == "" {
		return NewError("Device " + old + "not found")
	}
	setup.FullMac = newFull
	setup.Mac = new
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
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	return SaveOnUpdateObject(db, status, StatusDB, LedsTable, criteria)
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

//GetLedSwitchStatus get cluster Config list
func GetLedSwitchStatus(db Database, swMac string) map[string]dl.Led {
	res := map[string]dl.Led{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(StatusDB, LedsTable, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := dl.ToLed(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
}

//GetLedSwitchConfig get led Config list
func GetLedSwitchSetup(db Database, swMac string) map[string]dl.LedSetup {
	res := map[string]dl.LedSetup{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(StatusDB, LedsTable, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := dl.ToLedSetup(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
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
