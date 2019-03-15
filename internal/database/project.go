package database

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveProject dump project in database
func SaveProject(db Database, cfg core.Project) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = cfg.Label

	proj, dbID := GetProject(db, cfg.Label)
	if proj == nil || dbID == "" {
		_, err := db.InsertRecord(ConfigDB, ProjectsTable, cfg)
		return err
	}

	if cfg.Mac != nil {
		proj.Mac = cfg.Mac
	}
	if cfg.FullMac != nil {
		proj.FullMac = cfg.FullMac
	}
	if cfg.ModelName != nil {
		proj.ModelName = cfg.ModelName
	}
	err := db.UpdateRecord(ConfigDB, ProjectsTable, dbID, proj)
	return err

}

//RemoveProject remove project entry in database
func RemoveProject(db Database, label string) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	return db.DeleteRecord(ConfigDB, ProjectsTable, criteria)
}

//GetProject return the project configuration
func GetProject(db Database, label string) (*core.Project, string) {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(ConfigDB, ProjectsTable, criteria)
	if err != nil || stored == nil {
		return nil, ""
	}
	var dbID string
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	project, err := core.ToProject(stored)
	if err != nil {
		return nil, dbID
	}
	return project, dbID
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

//GetProjectByFullMac return the project configuration
func GetProjectByFullMac(db Database, mac string) *core.Project {
	criteria := make(map[string]interface{})
	criteria["FullMac"] = mac
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
