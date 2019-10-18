package database

import (
	dl "github.com/energieip/common-components-go/pkg/dled"
	"github.com/energieip/common-components-go/pkg/pconst"
)

//SaveLedConfig dump led config in database
func SaveLedConfig(db Database, cfg dl.LedSetup) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Mac
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbLeds, criteria)
}

//SaveLedLabelConfig dump led config in database
func SaveLedLabelConfig(db Database, cfg dl.LedSetup) error {
	criteria := make(map[string]interface{})
	if cfg.Label == nil {
		return NewError("Device " + cfg.Mac + "not found")
	}
	criteria["Label"] = *cfg.Label
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbLeds, criteria)
}

//RemoveLedConfig remove led config in database
func RemoveLedConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbConfig, pconst.TbLeds, criteria)
}

//RemoveLedStatus remove led status in database
func RemoveLedStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbStatus, pconst.TbLeds, criteria)
}

//RemoveSwitchLedStatus remove led status in database
func RemoveSwitchLedStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = mac
	return db.DeleteRecord(pconst.DbStatus, pconst.TbLeds, criteria)
}

//GetLedConfig return the led configuration
func GetLedConfig(db Database, mac string) (*dl.LedSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbLeds, criteria)
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
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbLeds, criteria)
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

	new := dl.UpdateConfig(config, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbLeds, dbID, &new)
}

//UpdateLedLabelConfig update led config in database
func UpdateLedLabelConfig(db Database, config dl.LedConf) error {
	if config.Label == nil {
		return NewError("Unknow Label")
	}
	setup, dbID := GetLedLabelConfig(db, *config.Label)
	if setup == nil || dbID == "" {
		return NewError("Device " + *config.Label + "not found")
	}

	new := dl.UpdateConfig(config, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbLeds, dbID, &new)
}

//UpdateLedSetup update led setup in database
func UpdateLedSetup(db Database, config dl.LedSetup) error {
	setup, dbID := GetLedConfig(db, config.Mac)
	if setup == nil || dbID == "" {
		config := dl.FillDefaultValue(config)
		return SaveLedLabelConfig(db, config)
	}

	new := dl.UpdateSetup(config, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbLeds, dbID, &new)
}

//CreateLedLabelSetup update led setup in database
func CreateLedLabelSetup(db Database, config dl.LedSetup) error {
	if config.Label == nil {
		return NewError("Device label not found")
	}
	setup, dbID := GetLedLabelConfig(db, *config.Label)
	if setup == nil || dbID == "" {
		config := dl.FillDefaultValue(config)
		return SaveLedLabelConfig(db, config)
	}

	return nil
}

//UpdateLedLabelSetup update led setup in database
func UpdateLedLabelSetup(db Database, config dl.LedSetup) error {
	if config.Label == nil {
		return NewError("Device label not found")
	}
	setup, dbID := GetLedLabelConfig(db, *config.Label)
	if setup == nil || dbID == "" {
		config := dl.FillDefaultValue(config)
		return SaveLedLabelConfig(db, config)
	}

	new := dl.UpdateSetup(config, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbLeds, dbID, &new)
}

//SwitchLedConfig update led config in database
func SwitchLedConfig(db Database, old, new string) error {
	setup, dbID := GetLedConfig(db, old)
	if setup == nil || dbID == "" {
		return NewError("Device " + old + "not found")
	}
	setup.Mac = new
	return db.UpdateRecord(pconst.DbConfig, pconst.TbLeds, dbID, setup)
}

//GetLedsConfig return the led config list
func GetLedsConfig(db Database) map[string]dl.LedSetup {
	leds := map[string]dl.LedSetup{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbLeds)
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

//GetLedsConfigByLabel return the led config list
func GetLedsConfigByLabel(db Database) map[string]dl.LedSetup {
	leds := map[string]dl.LedSetup{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbLeds)
	if err != nil || stored == nil {
		return leds
	}
	for _, l := range stored {
		light, err := dl.ToLedSetup(l)
		if err != nil || light == nil || light.Label == nil {
			continue
		}
		leds[*light.Label] = *light
	}
	return leds
}

//SaveLedStatus dump led status in database
func SaveLedStatus(db Database, status dl.Led) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	return SaveOnUpdateObject(db, status, pconst.DbStatus, pconst.TbLeds, criteria)
}

//GetLedsStatus return the led status list
func GetLedsStatus(db Database) map[string]dl.Led {
	leds := map[string]dl.Led{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbLeds)
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

//GetLedsStatusByLabel return the led status list
func GetLedsStatusByLabel(db Database) map[string]dl.Led {
	leds := map[string]dl.Led{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbLeds)
	if err != nil || stored == nil {
		return leds
	}
	for _, l := range stored {
		light, err := dl.ToLed(l)
		if err != nil || light == nil || light.Label == nil {
			continue
		}
		leds[*light.Label] = *light
	}
	return leds
}

//GetLedSwitchStatus get cluster Config list
func GetLedSwitchStatus(db Database, swMac string) map[string]dl.Led {
	res := map[string]dl.Led{}
	criteria := make(map[string]interface{})
	criteria["SwitchMac"] = swMac
	stored, err := db.GetRecords(pconst.DbStatus, pconst.TbLeds, criteria)
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
	stored, err := db.GetRecords(pconst.DbConfig, pconst.TbLeds, criteria)
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
	ledStored, err := db.GetRecord(pconst.DbStatus, pconst.TbLeds, criteria)
	if err != nil || ledStored == nil {
		return nil
	}
	light, err := dl.ToLed(ledStored)
	if err != nil {
		return nil
	}
	return light
}
