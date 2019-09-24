package database

import (
	"github.com/energieip/common-components-go/pkg/dserver"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/common-components-go/pkg/pconst"
)

//SaveSwitchStatus dump switch status in database
func SaveSwitchStatus(db Database, status sd.SwitchStatus) error {
	swStatus := dserver.SwitchDump{}
	swStatus.Mac = status.Mac
	swStatus.DumpFrequency = status.DumpFrequency
	swStatus.IP = status.IP
	swStatus.Cluster = status.Cluster
	swStatus.Label = status.Label
	swStatus.Profil = status.Profil
	swStatus.StatePuls1 = status.StatePuls1
	swStatus.StatePuls2 = status.StatePuls2
	swStatus.StatePuls3 = status.StatePuls3
	swStatus.StatePuls4 = status.StatePuls4
	swStatus.StatePuls5 = status.StatePuls5
	swStatus.StateBaes = status.StateBaes
	swStatus.LedsPower = status.LedsPower
	swStatus.BlindsPower = status.BlindsPower
	swStatus.HvacsPower = status.HvacsPower
	swStatus.TotalPower = status.TotalPower
	swStatus.HvacsEnergy = status.HvacsEnergy
	swStatus.LedsEnergy = status.LedsEnergy
	swStatus.BlindsEnergy = status.BlindsEnergy
	swStatus.TotalEnergy = status.TotalEnergy
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
func UpdateSwitchConfig(db Database, config dserver.SwitchConfig) error {
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

	setup, err := dserver.ToSwitchConfig(stored)
	if err != nil || stored == nil {
		return NewError("Switch " + *config.Mac + " not found")
	}

	if config.FriendlyName != nil {
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
	if config.Profil != "" {
		setup.Profil = config.Profil
	}
	if config.Mac != nil {
		setup.Mac = config.Mac
	}

	return db.UpdateRecord(pconst.DbConfig, pconst.TbSwitchs, dbID, setup)
}

//UpdateSwitchLabelConfig update server config to database
func UpdateSwitchLabelConfig(db Database, config dserver.SwitchConfig) error {
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

	setup, err := dserver.ToSwitchConfig(stored)
	if err != nil || stored == nil {
		return NewError("Switch " + *config.Label + " not found")
	}

	if config.FriendlyName != nil {
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
	if config.Profil != "" {
		setup.Profil = config.Profil
	}
	if config.Mac != nil {
		setup.Mac = config.Mac
	}

	return db.UpdateRecord(pconst.DbConfig, pconst.TbSwitchs, dbID, setup)
}

//SaveSwitchLabelConfig update server config to database
func SaveSwitchLabelConfig(db Database, config dserver.SwitchConfig) error {
	if config.Label == nil {
		return NewError("Switch Label not found")
	}
	criteria := make(map[string]interface{})
	criteria["Label"] = config.Label
	return SaveOnUpdateObject(db, config, pconst.DbConfig, pconst.TbSwitchs, criteria)
}

//RemoveSwitchConfig remove led config in database
func RemoveSwitchConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbConfig, pconst.TbSwitchs, criteria)
}

//RemoveSwitchStatus remove switch status in database
func RemoveSwitchStatus(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbStatus, pconst.TbSwitchs, criteria)
}

//SaveSwitchConfig register switch config in database
func SaveSwitchConfig(db Database, sw dserver.SwitchConfig) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = sw.Mac
	return SaveOnUpdateObject(db, sw, pconst.DbConfig, pconst.TbSwitchs, criteria)
}

//GetSwitchConfig get switch Config
func GetSwitchConfig(db Database, mac string) (*dserver.SwitchConfig, string) {
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
	sw, err := dserver.ToSwitchConfig(swStored)
	if err != nil || sw == nil {
		return nil, ""
	}
	return sw, dbID
}

//GetSwitchLabelConfig return the switch configuration
func GetSwitchLabelConfig(db Database, label string) *dserver.SwitchConfig {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	swStored, err := db.GetRecord(pconst.DbConfig, pconst.TbSwitchs, criteria)
	if err != nil || swStored == nil {
		return nil
	}
	sw, err := dserver.ToSwitchConfig(swStored)
	if err != nil || sw == nil {
		return nil
	}
	return sw
}

//ReplaceSwitchConfig update sensor config in database
func ReplaceSwitchConfig(db Database, old, new string) error {
	setup, dbID := GetSwitchConfig(db, old)
	if setup == nil || dbID == "" {
		return NewError("Device " + old + "not found")
	}
	setup.Mac = &new
	return db.UpdateRecord(pconst.DbConfig, pconst.TbSwitchs, dbID, setup)
}

//GetSwitchsConfig return the switch config list
func GetSwitchsConfig(db Database) map[string]dserver.SwitchConfig {
	switchs := map[string]dserver.SwitchConfig{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbSwitchs)
	if err != nil || stored == nil {
		return switchs
	}
	for _, l := range stored {
		sw, err := dserver.ToSwitchConfig(l)
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

//GetSwitchsConfigByLabel return the switch config list
func GetSwitchsConfigByLabel(db Database) map[string]dserver.SwitchConfig {
	switchs := map[string]dserver.SwitchConfig{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbSwitchs)
	if err != nil || stored == nil {
		return switchs
	}
	for _, l := range stored {
		sw, err := dserver.ToSwitchConfig(l)
		if err != nil || sw == nil {
			continue
		}
		if sw.Label == nil {
			continue
		}
		switchs[*sw.Label] = *sw
	}
	return switchs
}

//GetSwitchsDump return the switch status list
func GetSwitchsDump(db Database) map[string]dserver.SwitchDump {
	switchs := map[string]dserver.SwitchDump{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbSwitchs)
	if err != nil || stored == nil {
		return switchs
	}
	for _, l := range stored {
		sw, err := dserver.ToSwitchDump(l)
		if err != nil || sw == nil {
			continue
		}
		switchs[sw.Mac] = *sw
	}
	return switchs
}

//GetSwitchsDumpByLabel return the switch status list
func GetSwitchsDumpByLabel(db Database) map[string]dserver.SwitchDump {
	switchs := map[string]dserver.SwitchDump{}
	stored, err := db.FetchAllRecords(pconst.DbStatus, pconst.TbSwitchs)
	if err != nil || stored == nil {
		return switchs
	}
	for _, l := range stored {
		sw, err := dserver.ToSwitchDump(l)
		if err != nil || sw == nil || sw.Label == nil {
			continue
		}
		switchs[*sw.Label] = *sw
	}
	return switchs
}

//GetCluster get cluster Config list
func GetCluster(db Database, cluster int) map[string]dserver.SwitchConfig {
	res := make(map[string]dserver.SwitchConfig)
	criteria := make(map[string]interface{})
	criteria["Cluster"] = cluster
	swStored, err := db.GetRecords(pconst.DbConfig, pconst.TbSwitchs, criteria)
	if err != nil || swStored == nil {
		return res
	}
	for _, elt := range swStored {
		sw, err := dserver.ToSwitchConfig(elt)
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

//GetSwitchStatusCluster get cluster Config list
func GetSwitchStatusCluster(db Database, cluster int) map[string]dserver.SwitchDump {
	res := make(map[string]dserver.SwitchDump)
	criteria := make(map[string]interface{})
	criteria["Cluster"] = cluster
	swStored, err := db.GetRecords(pconst.DbStatus, pconst.TbSwitchs, criteria)
	if err != nil || swStored == nil {
		return res
	}
	for _, elt := range swStored {
		sw, err := dserver.ToSwitchDump(elt)
		if err != nil || sw == nil {
			continue
		}
		res[sw.Mac] = *sw
	}
	return res
}
