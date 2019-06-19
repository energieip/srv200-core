package database

import (
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/common-components-go/pkg/pconst"
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
	return SaveOnUpdateObject(db, swStatus, pconst.DbStatus, pconst.TbSwitchs, criteria)
}

//UpdateSwitchConfig update server config to database
func UpdateSwitchConfig(db Database, config core.SwitchConfig) error {
	if config.Label == nil {
		return NewError("Switch Mac not found")
	}
	criteria := make(map[string]interface{})
	criteria["Mac"] = config.Mac
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbSwitchs, criteria)
	if err != nil || stored == nil {
		return NewError("Switch " + *config.Mac + " not found")
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if !ok {
		return NewError("Switch " + *config.Mac + " not found")
	}
	dbID := id.(string)

	setup, err := core.ToSwitchConfig(stored)
	if err != nil || stored == nil {
		return NewError("Switch " + *config.Mac + " not found")
	}

	if config.FriendlyName == nil {
		setup.FriendlyName = config.FriendlyName
	}

	if config.DumpFrequency != nil {
		setup.DumpFrequency = config.DumpFrequency
	}

	if config.IP != nil {
		setup.IP = config.IP
	}
	if config.Cluster != nil {
		setup.Cluster = config.Cluster
	}
	if config.Label != nil {
		setup.Label = config.Label
	}

	return db.UpdateRecord(pconst.DbConfig, pconst.TbSwitchs, dbID, setup)
}

//UpdateSwitchLabelConfig update server config to database
func UpdateSwitchLabelConfig(db Database, config core.SwitchConfig) error {
	if config.Label == nil {
		return NewError("Switch Mac not found")
	}
	criteria := make(map[string]interface{})
	criteria["Label"] = config.Label
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbSwitchs, criteria)
	if err != nil || stored == nil {
		return NewError("Switch " + *config.Mac + " not found")
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if !ok {
		return NewError("Switch " + *config.Mac + " not found")
	}
	dbID := id.(string)

	setup, err := core.ToSwitchConfig(stored)
	if err != nil || stored == nil {
		return NewError("Switch " + *config.Mac + " not found")
	}

	if config.FriendlyName == nil {
		setup.FriendlyName = config.FriendlyName
	}

	if config.DumpFrequency != nil {
		setup.DumpFrequency = config.DumpFrequency
	}

	if config.IP != nil {
		setup.IP = config.IP
	}
	if config.Cluster != nil {
		setup.Cluster = config.Cluster
	}
	if config.Label != nil {
		setup.Label = config.Label
	}

	return db.UpdateRecord(pconst.DbConfig, pconst.TbSwitchs, dbID, setup)
}

//SaveSwitchLabelConfig update server config to database
func SaveSwitchLabelConfig(db Database, config core.SwitchConfig) error {
	if config.Label == nil {
		return NewError("Switch Label not found")
	}
	criteria := make(map[string]interface{})
	criteria["Label"] = config.Label
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbSwitchs, criteria)
	if err != nil || stored == nil {
		return NewError("Switch " + *config.Label + " not found")
	}
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if !ok {
		return NewError("Switch " + *config.Label + " not found")
	}
	dbID := id.(string)

	setup, err := core.ToSwitchConfig(stored)
	if err != nil || stored == nil {
		return NewError("Switch " + *config.Label + " not found")
	}

	if config.FriendlyName == nil {
		setup.FriendlyName = config.FriendlyName
	}

	if config.Mac == nil {
		setup.Mac = config.Mac
	}

	if config.FullMac == nil {
		setup.FullMac = config.FullMac
	}

	if config.DumpFrequency != nil {
		setup.DumpFrequency = config.DumpFrequency
	}
	if config.IP != nil {
		setup.IP = config.IP
	}
	if config.Cluster != nil {
		setup.Cluster = config.Cluster
	}
	setup.Label = config.Label

	return db.UpdateRecord(pconst.DbConfig, pconst.TbSwitchs, dbID, setup)
}

//RemoveSwitchConfig remove led config in database
func RemoveSwitchConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbConfig, pconst.TbSwitchs, criteria)
}

//SaveSwitchConfig register switch config in database
func SaveSwitchConfig(db Database, sw core.SwitchConfig) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = sw.Mac
	return SaveOnUpdateObject(db, sw, pconst.DbConfig, pconst.TbSwitchs, criteria)
}

//GetSwitchConfig get switch Config
func GetSwitchConfig(db Database, mac string) (*core.SwitchConfig, string) {
	dbID := ""
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	swStored, err := db.GetRecord(pconst.DbConfig, pconst.TbSwitchs, criteria)
	if err != nil || swStored == nil {
		return nil, dbID
	}
	m := swStored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	sw, err := core.ToSwitchConfig(swStored)
	if err != nil || sw == nil {
		return nil, ""
	}
	return sw, dbID
}

//GetSwitchLabelConfig return the switch configuration
func GetSwitchLabelConfig(db Database, label string) *core.SwitchConfig {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	swStored, err := db.GetRecord(pconst.DbConfig, pconst.TbSwitchs, criteria)
	if err != nil || swStored == nil {
		return nil
	}
	sw, err := core.ToSwitchConfig(swStored)
	if err != nil || sw == nil {
		return nil
	}
	return sw
}

//ReplaceSwitchConfig update sensor config in database
func ReplaceSwitchConfig(db Database, old, oldFull, new, newFull string) error {
	setup, dbID := GetSwitchConfig(db, old)
	if setup == nil || dbID == "" {
		return NewError("Device " + old + "not found")
	}
	setup.FullMac = &newFull
	setup.Mac = &new
	return db.UpdateRecord(pconst.DbConfig, pconst.TbSwitchs, dbID, setup)
}

//GetSwitchsConfig return the switch config list
func GetSwitchsConfig(db Database) map[string]core.SwitchConfig {
	switchs := map[string]core.SwitchConfig{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbSwitchs)
	if err != nil || stored == nil {
		return switchs
	}
	for _, l := range stored {
		sw, err := core.ToSwitchConfig(l)
		if err != nil || sw == nil {
			continue
		}
		if sw.Mac == nil {
			continue
		}
		switchs[*sw.Mac] = *sw
	}
	return switchs
}

//GetSwitchsDump return the switch status list
func GetSwitchsDump(db Database) map[string]core.SwitchDump {
	switchs := map[string]core.SwitchDump{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbSwitchs)
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
func GetCluster(db Database, cluster int) map[string]core.SwitchConfig {
	res := make(map[string]core.SwitchConfig)
	criteria := make(map[string]interface{})
	criteria["Cluster"] = cluster
	swStored, err := db.GetRecords(pconst.DbConfig, pconst.TbSwitchs, criteria)
	if err != nil || swStored == nil {
		return res
	}
	for _, elt := range swStored {
		sw, err := core.ToSwitchConfig(elt)
		if err != nil || sw == nil {
			continue
		}
		if sw.IP == nil {
			continue
		}
		res[*sw.IP] = *sw
	}
	return res
}
