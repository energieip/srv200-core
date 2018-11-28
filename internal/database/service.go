package database

import (
	sdevice "github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveServiceConfig dump sensor config in database
func SaveServiceConfig(db Database, service sdevice.Service) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Name"] = service.Name
	stored, err := db.GetRecord(ConfigDB, ServicesTable, criteria)
	if err == nil && stored != nil {
		m := stored.(map[string]interface{})
		id, ok := m["id"]
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, ServicesTable, service)
	} else {
		err = db.UpdateRecord(ConfigDB, ServicesTable, dbID, service)
	}
	return err
}

//GetServiceConfigs return the sensor configuration
func GetServiceConfigs(db Database) map[string]sdevice.Service {
	services := map[string]sdevice.Service{}
	stored, err := db.FetchAllRecords(ConfigDB, ServicesTable)
	if err != nil || stored == nil {
		return services
	}
	for _, s := range stored {
		serv, err := sdevice.ToService(s)
		if err != nil || serv == nil {
			continue
		}
		services[serv.Name] = *serv
	}
	return services
}

//SaveServiceStatus dump service status in database
func SaveServiceStatus(db Database, status core.ServiceDump) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Name"] = status.Name
	criteria["SwitchMac"] = status.SwitchMac
	stored, err := db.GetRecord(StatusDB, ServicesTable, criteria)
	if err == nil && stored != nil {
		m := stored.(map[string]interface{})
		id, ok := m["id"]
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(StatusDB, ServicesTable, status)
	} else {
		err = db.UpdateRecord(StatusDB, ServicesTable, dbID, status)
	}
	return err
}
