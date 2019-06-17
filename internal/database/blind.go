package database

import (
	"github.com/energieip/common-components-go/pkg/dblind"
)

//SaveBlindConfig dump blind config in database
func SaveBlindConfig(db Database, cfg dblind.BlindSetup) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Mac
	return SaveOnUpdateObject(db, cfg, ConfigDB, BlindsTable, criteria)
}

//SaveBlindLabelConfig dump blind config in database
func SaveBlindLabelConfig(db Database, cfg dblind.BlindSetup) error {
	criteria := make(map[string]interface{})
	if cfg.Label == nil {
		return NewError("Device " + cfg.Mac + "not found")
	}
	criteria["Label"] = *cfg.Label
	return SaveOnUpdateObject(db, cfg, ConfigDB, BlindsTable, criteria)
}

//UpdateBlindConfig update blind config in database
func UpdateBlindConfig(db Database, cfg dblind.BlindConf) error {
	setup, dbID := GetBlindConfig(db, cfg.Mac)
	if setup == nil || dbID == "" {
		return NewError("Device " + cfg.Mac + " not found")
	}

	if cfg.FriendlyName != nil {
		setup.FriendlyName = cfg.FriendlyName
	}

	if cfg.Group != nil {
		setup.Group = cfg.Group
	}

	if cfg.IsBleEnabled != nil {
		setup.IsBleEnabled = cfg.IsBleEnabled
	}

	if cfg.DumpFrequency != nil {
		setup.DumpFrequency = *cfg.DumpFrequency
	}

	if cfg.IBeaconMajor != nil {
		setup.IBeaconMajor = cfg.IBeaconMajor
	}

	if cfg.IBeaconMinor != nil {
		setup.IBeaconMinor = cfg.IBeaconMinor
	}

	if cfg.IBeaconTxPower != nil {
		setup.IBeaconTxPower = cfg.IBeaconTxPower
	}

	if cfg.IBeaconUUID != nil {
		setup.IBeaconUUID = cfg.IBeaconUUID
	}

	if cfg.BleMode != nil {
		setup.BleMode = cfg.BleMode
	}
	if cfg.Label != nil {
		setup.Label = cfg.Label
	}
	return db.UpdateRecord(ConfigDB, BlindsTable, dbID, setup)
}

//UpdateBlindLabelSetup update blind config in database
func UpdateBlindLabelSetup(db Database, cfg dblind.BlindSetup) error {
	setup, dbID := GetBlindLabelConfig(db, *cfg.Label)
	if setup == nil || dbID == "" {
		if cfg.FriendlyName != nil {
			name := *cfg.Label
			setup.FriendlyName = &name
		}
		if cfg.DumpFrequency == 0 {
			cfg.DumpFrequency = 1000
		}
		if cfg.BleMode == nil {
			ble := "service"
			cfg.BleMode = &ble
		}
		if cfg.IsBleEnabled == nil {
			bleEnable := false
			cfg.IsBleEnabled = &bleEnable
		}
		if cfg.Group == nil {
			group := 0
			cfg.Group = &group
		}
		if cfg.FriendlyName == nil {
			name := *cfg.Label
			cfg.FriendlyName = &name
		}
		return SaveBlindLabelConfig(db, cfg)
	}

	if cfg.FriendlyName != nil {
		setup.FriendlyName = cfg.FriendlyName
	}

	if cfg.Group != nil {
		setup.Group = cfg.Group
	}

	if cfg.IsBleEnabled != nil {
		setup.IsBleEnabled = cfg.IsBleEnabled
	}

	if cfg.DumpFrequency != 0 {
		setup.DumpFrequency = cfg.DumpFrequency
	}

	if cfg.Label != nil {
		setup.Label = cfg.Label
	}

	if cfg.IBeaconMajor != nil {
		setup.IBeaconMajor = cfg.IBeaconMajor
	}

	if cfg.IBeaconMinor != nil {
		setup.IBeaconMinor = cfg.IBeaconMinor
	}

	if cfg.IBeaconTxPower != nil {
		setup.IBeaconTxPower = cfg.IBeaconTxPower
	}

	if cfg.IBeaconUUID != nil {
		setup.IBeaconUUID = cfg.IBeaconUUID
	}

	if cfg.BleMode != nil {
		setup.BleMode = cfg.BleMode
	}

	return db.UpdateRecord(ConfigDB, BlindsTable, dbID, setup)
}

//SwitchBlindConfig update blind config in database
func SwitchBlindConfig(db Database, old, oldFull, new, newFull string) error {
	setup, dbID := GetBlindConfig(db, old)
	if setup == nil || dbID == "" {
		return NewError("Device " + old + " not found")
	}
	setup.FullMac = newFull
	setup.Mac = new
	return db.UpdateRecord(ConfigDB, BlindsTable, dbID, setup)
}

//RemoveBlindConfig remove blind config in database
func RemoveBlindConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(ConfigDB, BlindsTable, criteria)
}

//RemoveBlindStatus remove led status in database
func RemoveBlindStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(StatusDB, BlindsTable, criteria)
}

//GetBlindSwitchStatus get cluster Config list
func GetBlindSwitchStatus(db Database, swMac string) map[string]dblind.Blind {
	res := map[string]dblind.Blind{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(StatusDB, BlindsTable, criteria)
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
	stored, err := db.GetRecords(StatusDB, BlindsTable, criteria)
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
	stored, err := db.GetRecord(ConfigDB, BlindsTable, criteria)
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
	stored, err := db.GetRecord(ConfigDB, BlindsTable, criteria)
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
	stored, err := db.FetchAllRecords(ConfigDB, BlindsTable)
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

//SaveBlindStatus dump blind status in database
func SaveBlindStatus(db Database, status dblind.Blind) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	return SaveOnUpdateObject(db, status, StatusDB, BlindsTable, criteria)
}

//GetBlindsStatus return the blind status list
func GetBlindsStatus(db Database) map[string]dblind.Blind {
	drivers := map[string]dblind.Blind{}
	stored, err := db.FetchAllRecords(StatusDB, BlindsTable)
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

//GetBlindStatus return the blind status
func GetBlindStatus(db Database, mac string) *dblind.Blind {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(StatusDB, BlindsTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	driver, err := dblind.ToBlind(stored)
	if err != nil {
		return nil
	}
	return driver
}
