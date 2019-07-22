package database

import (
	"github.com/energieip/common-components-go/pkg/pconst"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveServiceConfig dump sensor config in database
func SaveServiceConfig(db Database, service pkg.Service) error {
	criteria := make(map[string]interface{})
	criteria["Name"] = service.Name
	return SaveOnUpdateObject(db, service, pconst.DbConfig, pconst.TbServices, criteria)
}

//GetServiceConfig get service Config
func GetServiceConfig(db Database, name string) *pkg.Service {
	criteria := make(map[string]interface{})
	criteria["Name"] = name
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbServices, criteria)
	if err != nil || stored == nil {
		return nil
	}
	service, err := pkg.ToService(stored)
	if err != nil || service == nil {
		return nil
	}
	return service
}

//GetServiceConfigs return the sensor configuration
func GetServiceConfigs(db Database) map[string]pkg.Service {
	services := map[string]pkg.Service{}
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbServices)
	if err != nil || stored == nil {
		return services
	}
	for _, s := range stored {
		serv, _ := pkg.ToService(s)
		services[serv.Name] = *serv
	}
	return services
}

//SaveServiceStatus dump service status in database
func SaveServiceStatus(db Database, status core.ServiceDump) error {
	criteria := make(map[string]interface{})
	criteria["Name"] = status.Name
	criteria["SwitchMac"] = status.SwitchMac
	return SaveOnUpdateObject(db, status, pconst.DbStatus, pconst.TbServices, criteria)
}

//RemoveServiceConfig remove sensor config in database
func RemoveServiceConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(pconst.DbConfig, pconst.TbServices, criteria)
}
