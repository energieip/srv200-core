package database

import (
	"strings"

	"github.com/energieip/common-components-go/pkg/pconst"
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveProject dump project in database
func SaveProject(db Database, cfg core.Project) error {
	proj, dbID := GetProject(db, cfg.Label)
	if proj == nil || dbID == "" {
		if cfg.Mac != nil {
			mac := strings.ToUpper(*cfg.Mac)
			cfg.Mac = &mac
		}
		if cfg.ModelName != nil {
			model := strings.ToUpper(*cfg.ModelName)
			cfg.ModelName = &model
		}

		_, err := db.InsertRecord(pconst.DbConfig, pconst.TbProjects, cfg)
		return err
	}

	if cfg.Mac != nil {
		mac := strings.ToUpper(*cfg.Mac)
		proj.Mac = &mac
	}
	if cfg.ModelName != nil {
		model := strings.ToUpper(*cfg.ModelName)
		proj.ModelName = &model
	}
	if cfg.ModbusID != nil {
		proj.ModbusID = cfg.ModbusID
	}
	if cfg.SlaveID != nil {
		proj.SlaveID = cfg.SlaveID
	}
	if cfg.CommissioningDate != nil {
		proj.CommissioningDate = cfg.CommissioningDate
	}
	err := db.UpdateRecord(pconst.DbConfig, pconst.TbProjects, dbID, proj)
	return err
}

//RemoveProject remove project entry in database
func RemoveProject(db Database, label string) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	return db.DeleteRecord(pconst.DbConfig, pconst.TbProjects, criteria)
}

//GetProject return the project configuration
func GetProject(db Database, label string) (*core.Project, string) {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbProjects, criteria)
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
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbProjects, criteria)
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
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbProjects)
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
