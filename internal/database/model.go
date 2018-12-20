package database

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveModel dump model in database
func SaveModel(db Database, m core.Model) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Name"] = m.Name
	stored, err := db.GetRecord(ConfigDB, ModelsTable, criteria)
	if err == nil && stored != nil {
		m := stored.(map[string]interface{})
		id, ok := m["id"]
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, ModelsTable, m)
	} else {
		err = db.UpdateRecord(ConfigDB, ModelsTable, dbID, m)
	}
	return err
}

//RemoveModel remove ifc config in database
func RemoveModel(db Database, name string) error {
	criteria := make(map[string]interface{})
	criteria["Name"] = name
	return db.DeleteRecord(ConfigDB, ModelsTable, criteria)
}

//GetModel return the model configuration
func GetModel(db Database, name string) *core.Model {
	criteria := make(map[string]interface{})
	criteria["Name"] = name
	stored, err := db.GetRecord(ConfigDB, ModelsTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	model, err := core.ToModel(stored)
	if err != nil {
		return nil
	}
	return model
}

//GetModels return the models configuration
func GetModels(db Database) map[string]core.Model {
	models := map[string]core.Model{}
	stored, err := db.FetchAllRecords(ConfigDB, ModelsTable)
	if err != nil || stored == nil {
		return models
	}
	for _, m := range stored {
		model, err := core.ToModel(m)
		if err != nil || model == nil {
			continue
		}
		models[model.Name] = *model
	}
	return models
}
