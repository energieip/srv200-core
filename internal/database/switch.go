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
	swStatus.Topic = status.Topic
	swStatus.Services = status.Services

	var leds []string
	for mac := range status.Leds {
		leds = append(leds, mac)
	}
	swStatus.Leds = leds

	var sensors []string
	for mac := range status.Sensors {
		sensors = append(sensors, mac)
	}
	swStatus.Sensors = sensors

	var groups []int
	for grID := range status.Groups {
		groups = append(groups, grID)
	}
	swStatus.Groups = groups

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
func GetSwitchConfig(db Database, mac string) *sdevice.SwitchConfig {

	return nil
}
