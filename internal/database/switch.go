package database

import (
	sdevice "github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveSwitchStatus dump switch status in database
func SaveSwitchStatus(db Database, status sdevice.SwitchStatus) error {
	swStatus := switchDump{}
	swStatus.Mac = status.Mac
	swStatus.IP = status.IP
	swStatus.ErrorCode = status.ErrorCode
	swStatus.IsConfigured = status.IsConfigured
	swStatus.Topic = status.Topic
	swStatus.Protocol = status.Protocol
	swStatus.FriendlyName = status.FriendlyName
	swStatus.LastSystemUpgradeDate = status.LastSystemUpgradeDate

	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = status.Mac
	swStored, err := db.GetRecord(StatusDB, SwitchsTable, criteria)
	if err == nil && swStored != nil {
		m := swStored.(map[string]interface{})
		id, ok := m["id"]
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(StatusDB, SwitchsTable, swStatus)
	} else {
		err = db.UpdateRecord(StatusDB, SwitchsTable, dbID, swStatus)
	}
	return err
}

//SaveSwitchConfig register switch config in database
func SaveSwitchConfig(db Database, sw core.SwitchSetup) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = sw.Mac
	switchStored, err := db.GetRecord(ConfigDB, SwitchsTable, criteria)
	if err == nil && switchStored != nil {
		m := switchStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, SwitchsTable, sw)
	} else {
		err = db.UpdateRecord(ConfigDB, SwitchsTable, dbID, sw)
	}
	return err
}

//GetSwitchConfig get switch Config
func GetSwitchConfig(db Database, mac string) *core.SwitchSetup {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	swStored, err := db.GetRecord(ConfigDB, SwitchsTable, criteria)
	if err != nil || swStored == nil {
		return nil
	}
	sw, err := core.ToSwitchSetup(swStored)
	if err != nil || sw == nil {
		return nil
	}
	return sw
}
