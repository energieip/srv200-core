package database

import (
	"github.com/energieip/common-components-go/pkg/pconst"
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveFrame dump project in database
func SaveFrame(db Database, cfg core.Frame) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = cfg.Label

	fr, dbID := GetFrame(db, cfg.Label)
	if fr == nil || dbID == "" {
		_, err := db.InsertRecord(pconst.DbConfig, pconst.TbFrames, cfg)
		return err
	}
	return db.UpdateRecord(pconst.DbConfig, pconst.TbFrames, dbID, cfg)
}

//RemoveFrame remove project entry in database
func RemoveFrame(db Database, label string) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	return db.DeleteRecord(pconst.DbConfig, pconst.TbFrames, criteria)
}

//GetFrame return the project configuration
func GetFrame(db Database, label string) (*core.Frame, string) {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbFrames, criteria)
	if err != nil || stored == nil {
		return nil, ""
	}
	var dbID string
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	project, err := core.ToFrame(stored)
	if err != nil {
		return nil, dbID
	}
	return project, dbID
}

//GetFrames return the frame configuration
func GetFrames(db Database) []core.Frame {
	var projects []core.Frame
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbFrames)
	if err != nil || stored == nil {
		return nil
	}
	for _, st := range stored {
		project, err := core.ToFrame(st)
		if err != nil || project == nil {
			continue
		}
		projects = append(projects, *project)
	}
	return projects
}
