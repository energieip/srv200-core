package database

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveProject dump project in database
func SaveProject(db Database, m core.Project) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Label"] = m.Label
	stored, err := db.GetRecord(ConfigDB, ProjectsTable, criteria)
	if err == nil && stored != nil {
		m := stored.(map[string]interface{})
		id, ok := m["id"]
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, ProjectsTable, m)
	} else {
		err = db.UpdateRecord(ConfigDB, ProjectsTable, dbID, m)
	}
	return err
}

//GetProject return the project configuration
func GetProject(db Database, label string) *core.Project {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(ConfigDB, ProjectsTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	project, err := core.ToProject(stored)
	if err != nil {
		return nil
	}
	return project
}
