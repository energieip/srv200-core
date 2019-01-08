package database

import (
	gm "github.com/energieip/common-group-go/pkg/groupmodel"
)

//SaveGroupConfig dump group config in database
func SaveGroupConfig(db Database, status gm.GroupConfig) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Group"] = status.Group
	grStored, err := db.GetRecord(ConfigDB, GroupsTable, criteria)
	if err == nil && grStored != nil {
		m := grStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, GroupsTable, status)
	} else {
		err = db.UpdateRecord(ConfigDB, GroupsTable, dbID, status)
	}
	return err
}

//RemoveGroupConfig remove group config in database
func RemoveGroupConfig(db Database, grID int) error {
	criteria := make(map[string]interface{})
	criteria["Group"] = grID
	return db.DeleteRecord(ConfigDB, GroupsTable, criteria)
}

//GetGroupConfig return the group configuration
func GetGroupConfig(db Database, grID int) *gm.GroupConfig {
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
	return gr
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
		led := GetLedConfig(db, ledMac)
		if led == nil {
			continue
		}
		switchs[led.SwitchMac] = true
	}
	return switchs
}

//UpdateGroupConfig update group config in database
func UpdateGroupConfig(db Database, config gm.GroupConfig) error {
	criteria := make(map[string]interface{})
	criteria["Group"] = config.Group
	stored, err := db.GetRecord(ConfigDB, GroupsTable, criteria)
	if err != nil || stored == nil {
		return NewError("Group " + string(config.Group) + "not found")
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if !ok {
		id, ok = m["ID"]
	}
	if !ok {
		return NewError("Group " + string(config.Group) + "not found")
	}
	dbID := id.(string)

	setup, err := gm.ToGroupConfig(stored)
	if err != nil || stored == nil {
		return NewError("Group " + string(config.Group) + "not found")
	}

	if config.Leds != nil {
		setup.Leds = config.Leds
	}

	if config.Sensors != nil {
		setup.Sensors = config.Sensors
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

	if config.SlopeStart != nil {
		setup.SlopeStart = config.SlopeStart
	}

	if config.SlopeStop != nil {
		setup.SlopeStop = config.SlopeStop
	}

	if config.SensorRule != nil {
		setup.SensorRule = config.SensorRule
	}

	if config.Auto != nil {
		setup.Auto = config.Auto
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
		if addGroup {
			groups[gr.Group] = *gr
		}
	}
	return groups
}

//SaveGroupStatus dump group status in database
func SaveGroupStatus(db Database, status gm.GroupStatus) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Group"] = status.Group
	grStored, err := db.GetRecord(StatusDB, GroupsTable, criteria)
	if err == nil && grStored != nil {
		m := grStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(StatusDB, GroupsTable, status)
	} else {
		err = db.UpdateRecord(StatusDB, GroupsTable, dbID, status)
	}
	return err
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
