package database

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveModel dump model in database
func SaveModel(db Database, cfg core.Model) error {
	var err error
	criteria := make(map[string]interface{})
	criteria["Name"] = cfg.Name
	dbID := GetObjectID(db, ConfigDB, ModelsTable, criteria)
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, ModelsTable, cfg)
	} else {
		err = db.UpdateRecord(ConfigDB, ModelsTable, dbID, cfg)
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
