package database

import (
	gm "github.com/energieip/common-components-go/pkg/dgroup"
)

//SaveGroupConfig dump group config in database
func SaveGroupConfig(db Database, cfg gm.GroupConfig) error {
	criteria := make(map[string]interface{})
	criteria["Group"] = cfg.Group
	return SaveOnUpdateObject(db, cfg, ConfigDB, GroupsTable, criteria)
}

//RemoveGroupConfig remove group config in database
func RemoveGroupConfig(db Database, grID int) error {
	criteria := make(map[string]interface{})
	criteria["Group"] = grID
	return db.DeleteRecord(ConfigDB, GroupsTable, criteria)
}

//GetGroupConfig return the group configuration
func GetGroupConfig(db Database, grID int) (*gm.GroupConfig, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Group"] = grID
	stored, err := db.GetRecord(ConfigDB, GroupsTable, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	gr, err := gm.ToGroupConfig(stored)
	if err != nil {
		return nil, dbID
	}
	return gr, dbID
}

//GetGroupSwitchs return the corresponding running switch list
func GetGroupSwitchs(db Database, grID int) map[string]bool {
	switchs := make(map[string]bool)
	criteria := make(map[string]interface{})
	criteria["Group"] = grID
	stored, err := db.GetRecord(ConfigDB, GroupsTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	gr, err := gm.ToGroupConfig(stored)
	if err != nil {
		return nil
	}
	for _, ledMac := range gr.Leds {
		led, _ := GetLedConfig(db, ledMac)
		if led == nil {
			continue
		}
		switchs[led.SwitchMac] = true
	}
	for _, blindMac := range gr.Blinds {
		blind, _ := GetBlindConfig(db, blindMac)
		if blind == nil {
			continue
		}
		switchs[blind.SwitchMac] = true
	}
	return switchs
}

//UpdateGroupConfig update group config in database
func UpdateGroupConfig(db Database, config gm.GroupConfig) error {
	setup, dbID := GetGroupConfig(db, config.Group)
	if setup == nil || dbID == "" {
		return NewError("Group " + string(config.Group) + "not found")
	}

	if config.Leds != nil {
		setup.Leds = config.Leds
	}

	if config.Sensors != nil {
		setup.Sensors = config.Sensors
	}

	if config.Blinds != nil {
		setup.Blinds = config.Blinds
	}

	if config.FriendlyName != nil {
		setup.FriendlyName = config.FriendlyName
	}

	if config.CorrectionInterval != nil {
		setup.CorrectionInterval = config.CorrectionInterval
	}

	if config.Watchdog != nil {
		setup.Watchdog = config.Watchdog
	}

	if config.SlopeStartManual != nil {
		setup.SlopeStartManual = config.SlopeStartManual
	}

	if config.SlopeStopManual != nil {
		setup.SlopeStopManual = config.SlopeStopManual
	}

	if config.SlopeStartAuto != nil {
		setup.SlopeStartAuto = config.SlopeStartAuto
	}

	if config.SlopeStopAuto != nil {
		setup.SlopeStopAuto = config.SlopeStopAuto
	}

	if config.SensorRule != nil {
		setup.SensorRule = config.SensorRule
	}

	if config.Auto != nil {
		setup.Auto = config.Auto
	}

	if config.RuleBrightness != nil {
		setup.RuleBrightness = config.RuleBrightness
	}

	if config.RulePresence != nil {
		setup.RulePresence = config.RulePresence
	}

	if config.FirstDay != nil {
		setup.FirstDay = config.FirstDay
	}

	if config.FirstDayOffset != nil {
		setup.FirstDayOffset = config.FirstDayOffset
	}
	return db.UpdateRecord(ConfigDB, GroupsTable, dbID, setup)
}

//GetGroupConfigs get group Config
func GetGroupConfigs(db Database, driversMac map[string]bool) map[int]gm.GroupConfig {
	groups := make(map[int]gm.GroupConfig)
	stored, err := db.FetchAllRecords(ConfigDB, GroupsTable)
	if err != nil || stored == nil {
		return groups
	}
	for _, val := range stored {
		gr, err := gm.ToGroupConfig(val)
		if err != nil || gr == nil {
			continue
		}
		addGroup := false
		for _, mac := range gr.Leds {
			if _, ok := driversMac[mac]; ok {
				addGroup = true
				break
			}
		}
		if addGroup != true {
			for _, mac := range gr.Blinds {
				if _, ok := driversMac[mac]; ok {
					addGroup = true
					break
				}
			}
		}
		if addGroup {
			groups[gr.Group] = *gr
		}
	}
	return groups
}

//SaveGroupStatus dump group status in database
func SaveGroupStatus(db Database, status gm.GroupStatus) error {
	criteria := make(map[string]interface{})
	criteria["Group"] = status.Group
	return SaveOnUpdateObject(db, status, StatusDB, GroupsTable, criteria)
}

//GetGroupStatus return the group status
func GetGroupStatus(db Database, grID int) *gm.GroupStatus {
	criteria := make(map[string]interface{})
	criteria["Group"] = grID
	stored, err := db.GetRecord(StatusDB, GroupsTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	gr, err := gm.ToGroupStatus(stored)
	if err != nil {
		return nil
	}
	return gr
}
