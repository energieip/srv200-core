package database

import group "github.com/energieip/common-group-go/pkg/groupmodel"

//SaveGroupConfig dump group config in database
func SaveGroupConfig(db Database, status group.GroupConfig) error {
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

//GetGroupConfigs get group Config
func GetGroupConfigs(db Database, driversMac map[string]bool) map[int]group.GroupConfig {
	groups := make(map[int]group.GroupConfig)
	stored, err := db.FetchAllRecords(ConfigDB, GroupsTable)
	if err != nil || stored == nil {
		return groups
	}
	for _, val := range stored {
		gr, err := group.ToGroupConfig(val)
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
func SaveGroupStatus(db Database, status group.GroupStatus) error {
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
