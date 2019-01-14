package database

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveProject dump project in database
func SaveProject(db Database, cfg core.Project) error {
	var err error
	criteria := make(map[string]interface{})
	criteria["Label"] = cfg.Label
	dbID := GetObjectID(db, ConfigDB, ProjectsTable, criteria)
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, ProjectsTable, cfg)
	} else {
		err = db.UpdateRecord(ConfigDB, ProjectsTable, dbID, cfg)
	}
	return err
}

//RemoveProject remove project entry in database
func RemoveProject(db Database, label string) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	return db.DeleteRecord(ConfigDB, ProjectsTable, criteria)
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

//GetProjectByMac return the project configuration
func GetProjectByMac(db Database, mac string) *core.Project {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
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

//GetProjects return the project configuration
func GetProjects(db Database) []core.Project {
	var projects []core.Project
	stored, err := db.FetchAllRecords(ConfigDB, ProjectsTable)
	if err != nil || stored == nil {
		return nil
	}
	for _, st := range stored {
		project, err := core.ToProject(st)
		if err != nil || project == nil {
			continue
		}
		projects = append(projects, *project)
	}
	return projects
}
