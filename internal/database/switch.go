package database

import (
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveSwitchStatus dump switch status in database
func SaveSwitchStatus(db Database, status sd.SwitchStatus) error {
	swStatus := core.SwitchDump{}
	swStatus.Mac = status.Mac
	swStatus.IP = status.IP
	swStatus.ErrorCode = status.ErrorCode
	swStatus.IsConfigured = status.IsConfigured
	swStatus.Protocol = status.Protocol
	swStatus.FriendlyName = status.FriendlyName
	swStatus.LastSystemUpgradeDate = status.LastSystemUpgradeDate

	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	return SaveOnUpdateObject(db, swStatus, StatusDB, SwitchsTable, criteria)
}

//UpdateSwitchConfig update server config to database
func UpdateSwitchConfig(db Database, config core.SwitchConfig) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = config.Mac
	stored, err := db.GetRecord(ConfigDB, SwitchsTable, criteria)
	if err != nil || stored == nil {
		return NewError("Switch " + config.Mac + "not found")
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if !ok {
		return NewError("Switch " + config.Mac + "not found")
	}
	dbID := id.(string)

	setup, err := core.ToSwitchConfig(stored)
	if err != nil || stored == nil {
		return NewError("Switch " + config.Mac + "not found")
	}

	setup.FriendlyName = config.FriendlyName
	if config.DumpFrequency != nil {
		setup.DumpFrequency = config.DumpFrequency
	}

	setup.IP = config.IP
	setup.Cluster = config.Cluster

	return db.UpdateRecord(ConfigDB, SwitchsTable, dbID, setup)
}

//RemoveSwitchConfig remove led config in database
func RemoveSwitchConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(ConfigDB, SwitchsTable, criteria)
}

//SaveSwitchConfig register switch config in database
func SaveSwitchConfig(db Database, sw core.SwitchConfig) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = sw.Mac
	return SaveOnUpdateObject(db, sw, ConfigDB, SwitchsTable, criteria)
}

//GetSwitchConfig get switch Config
func GetSwitchConfig(db Database, mac string) *core.SwitchConfig {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	swStored, err := db.GetRecord(ConfigDB, SwitchsTable, criteria)
	if err != nil || swStored == nil {
		return nil
	}
	sw, err := core.ToSwitchConfig(swStored)
	if err != nil || sw == nil {
		return nil
	}
	return sw
}

//GetSwitchsConfig return the switch config list
func GetSwitchsConfig(db Database) map[string]core.SwitchConfig {
	switchs := map[string]core.SwitchConfig{}
	stored, err := db.FetchAllRecords(ConfigDB, SwitchsTable)
	if err != nil || stored == nil {
		return switchs
	}
	for _, l := range stored {
		sw, err := core.ToSwitchConfig(l)
		if err != nil || sw == nil {
			continue
		}
		switchs[sw.Mac] = *sw
	}
	return switchs
}

//GetSwitchsDump return the switch status list
func GetSwitchsDump(db Database) map[string]core.SwitchDump {
	switchs := map[string]core.SwitchDump{}
	stored, err := db.FetchAllRecords(StatusDB, SwitchsTable)
	if err != nil || stored == nil {
		return switchs
	}
	for _, l := range stored {
		sw, err := core.ToSwitchDump(l)
		if err != nil || sw == nil {
			continue
		}
		switchs[sw.Mac] = *sw
	}
	return switchs
}

//GetCluster get cluster Config list
func GetCluster(db Database, cluster int) []core.SwitchConfig {
	var res []core.SwitchConfig
	criteria := make(map[string]interface{})
	criteria["Cluster"] = cluster
	swStored, err := db.GetRecords(ConfigDB, SwitchsTable, criteria)
	if err != nil || swStored == nil {
		return res
	}
	for _, elt := range swStored {
		sw, err := core.ToSwitchConfig(elt)
		if err != nil || sw == nil {
			continue
		}
		res = append(res, *sw)
	}
	return res
}
