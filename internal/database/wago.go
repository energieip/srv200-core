package database

import (
	"github.com/energieip/common-components-go/pkg/dwago"
	"github.com/energieip/common-components-go/pkg/pconst"
)

//SaveWagoConfig dump wago config in database
func SaveWagoConfig(db Database, cfg dwago.WagoSetup) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = cfg.Mac
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbWagos, criteria)
}

//SaveWagoLabelConfig dump wago config in database
func SaveWagoLabelConfig(db Database, cfg dwago.WagoSetup) error {
	criteria := make(map[string]interface{})
	if cfg.Label == nil {
		return NewError("Device " + cfg.Mac + "not found")
	}
	criteria["Label"] = *cfg.Label
	return SaveOnUpdateObject(db, cfg, pconst.DbConfig, pconst.TbWagos, criteria)
}

//RemoveWagoConfig remove wago config in database
func RemoveWagoConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbConfig, pconst.TbWagos, criteria)
}

//RemoveWagoStatus remove wago status in database
func RemoveWagoStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbStatus, pconst.TbWagos, criteria)
}

//GetWagoConfig return the wago configuration
func GetWagoConfig(db Database, mac string) (*dwago.WagoSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbWagos, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := dwago.ToWagoSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//GetWagoLabelConfig return the wago configuration
func GetWagoLabelConfig(db Database, label string) (*dwago.WagoSetup, string) {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbWagos, criteria)
	if err != nil || stored == nil {
		return nil, dbID
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	driver, err := dwago.ToWagoSetup(stored)
	if err != nil {
		return nil, dbID
	}
	return driver, dbID
}

//UpdateWagoConfig update wago config in database
func UpdateWagoConfig(db Database, config dwago.WagoConf) error {
	setup, dbID := GetWagoConfig(db, config.Mac)
	if setup == nil || dbID == "" {
		return NewError("Device " + config.Mac + "not found")
	}

	new := dwago.UpdateConfig(config, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbWagos, dbID, &new)
}

//UpdateWagoLabelConfig update wago config in database
func UpdateWagoLabelConfig(db Database, config dwago.WagoConf) error {
	if config.Label == nil {
		return NewError("Unknow Label")
	}
	setup, dbID := GetWagoLabelConfig(db, *config.Label)
	if setup == nil || dbID == "" {
		return NewError("Device " + *config.Label + "not found")
	}

	new := dwago.UpdateConfig(config, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbWagos, dbID, &new)
}

//UpdateWagoSetup update wago setup in database
func UpdateWagoSetup(db Database, config dwago.WagoSetup) error {
	setup, dbID := GetWagoConfig(db, config.Mac)
	if setup == nil || dbID == "" {
		config := dwago.FillDefaultValue(config)
		return SaveWagoLabelConfig(db, config)
	}

	new := dwago.UpdateSetup(config, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbWagos, dbID, &new)
}

//UpdateWagoLabelSetup update wago setup in database
func UpdateWagoLabelSetup(db Database, config dwago.WagoSetup) error {
	if config.Label == nil {
		return NewError("Device label not found")
	}
	setup, dbID := GetWagoLabelConfig(db, *config.Label)
	if setup == nil || dbID == "" {
		config := dwago.FillDefaultValue(config)
		return SaveWagoLabelConfig(db, config)
	}

	new := dwago.UpdateSetup(config, *setup)
	return db.UpdateRecord(pconst.DbConfig, pconst.TbWagos, dbID, &new)
}

//SwitchWagoConfig update wago config in database
func SwitchWagoConfig(db Database, old, new string) error {
	setup, dbID := GetWagoConfig(db, old)
	if setup == nil || dbID == "" {
		return NewError("Device " + old + "not found")
	}
	setup.Mac = new
	return db.UpdateRecord(pconst.DbConfig, pconst.TbWagos, dbID, setup)
}

//GetWagosConfig return the wago config list
func GetWagosConfig(db Database) map[string]dwago.WagoSetup {
	wagos := map[string]dwago.WagoSetup{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbWagos)
	if err != nil || stored == nil {
		return wagos
	}
	for _, l := range stored {
		w, err := dwago.ToWagoSetup(l)
		if err != nil || w == nil {
			continue
		}
		wagos[w.Mac] = *w
	}
	return wagos
}

//GetWagosConfigByLabel return the wago config list
func GetWagosConfigByLabel(db Database) map[string]dwago.WagoSetup {
	wagos := map[string]dwago.WagoSetup{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbWagos)
	if err != nil || stored == nil {
		return wagos
	}
	for _, l := range stored {
		w, err := dwago.ToWagoSetup(l)
		if err != nil || w == nil || w.Label == nil {
			continue
		}
		wagos[*w.Label] = *w
	}
	return wagos
}

//SaveWagoStatus dump wago status in database
func SaveWagoStatus(db Database, status dwago.Wago) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	return SaveOnUpdateObject(db, status, pconst.DbStatus, pconst.TbWagos, criteria)
}

//GetWagosStatus return the wago status list
func GetWagosStatus(db Database) map[string]dwago.Wago {
	wagos := map[string]dwago.Wago{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbWagos)
	if err != nil || stored == nil {
		return wagos
	}
	for _, l := range stored {
		w, err := dwago.ToWago(l)
		if err != nil || w == nil {
			continue
		}
		wagos[w.Mac] = *w
	}
	return wagos
}

//GetWagosStatusByLabel return the wago status list
func GetWagosStatusByLabel(db Database) map[string]dwago.Wago {
	wagos := map[string]dwago.Wago{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbWagos)
	if err != nil || stored == nil {
		return wagos
	}
	for _, l := range stored {
		w, err := dwago.ToWago(l)
		if err != nil || w == nil || w.Label == nil {
			continue
		}
		wagos[*w.Label] = *w
	}
	return wagos
}

//GetWagoClusterStatus get cluster Config list
func GetWagoClusterStatus(db Database, cluster int) map[string]dwago.Wago {
	res := map[string]dwago.Wago{}
	criteria := make(map[string]interface{})
	criteria["Cluster"] = cluster
	stored, err := db.GetRecords(pconst.DbStatus, pconst.TbWagos, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := dwago.ToWago(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
}

//GetWagoClusterSetup get wago Config list
func GetWagoClusterSetup(db Database, cluster int) map[string]dwago.WagoSetup {
	res := map[string]dwago.WagoSetup{}
	criteria := make(map[string]interface{})
	criteria["Cluster"] = cluster
	stored, err := db.GetRecords(pconst.DbConfig, pconst.TbWagos, criteria)
	if err != nil || stored == nil {
		return res
	}
	for _, elt := range stored {
		driver, err := dwago.ToWagoSetup(elt)
		if err != nil || driver == nil {
			continue
		}
		res[driver.Mac] = *driver
	}
	return res
}

//GetWagoStatus return the wago status
func GetWagoStatus(db Database, mac string) *dwago.Wago {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	stored, err := db.GetRecord(pconst.DbStatus, pconst.TbWagos, criteria)
	if err != nil || stored == nil {
		return nil
	}
	w, err := dwago.ToWago(stored)
	if err != nil {
		return nil
	}
	return w
}
