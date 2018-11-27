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

//GetGroupConfig get group Config
func GetGroupConfig(db Database, grID int) *group.GroupConfig {
	criteria := make(map[string]interface{})
	criteria["Group"] = grID
	stored, err := db.GetRecord(ConfigDB, GroupsTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	gr, err := group.ToGroupConfig(stored)
	if err != nil {
		return nil
	}
	return gr
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
