package database

import (
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveServiceConfig dump sensor config in database
func SaveServiceConfig(db Database, service pkg.Service) error {
	serv := core.Service{
		Name:        service.Name,
		Systemd:     service.Systemd,
		Version:     service.Version,
		PackageName: service.PackageName,
		ConfigPath:  service.ConfigPath,
	}
	var err error
	criteria := make(map[string]interface{})
	criteria["Name"] = service.Name
	dbID := GetObjectID(db, ConfigDB, ServicesTable, criteria)
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, ServicesTable, serv)
	} else {
		err = db.UpdateRecord(ConfigDB, ServicesTable, dbID, serv)
	}
	return err
}

//GetServiceConfig get service Config
func GetServiceConfig(db Database, name string) *core.Service {
	criteria := make(map[string]interface{})
	criteria["Name"] = name
	stored, err := db.GetRecord(ConfigDB, ServicesTable, criteria)
	if err != nil || stored == nil {
		return nil
	}
	service, err := core.ToService(stored)
	if err != nil || service == nil {
		return nil
	}
	return service
}

//GetServiceConfigs return the sensor configuration
func GetServiceConfigs(db Database, switchIP, serverIP string, cluster int) map[string]pkg.Service {
	services := map[string]pkg.Service{}
	stored, err := db.FetchAllRecords(ConfigDB, ServicesTable)
	if err != nil || stored == nil {
		return services
	}
	for _, s := range stored {
		serv, err := core.ToService(s)
		if err != nil || serv == nil {
			continue
		}
		localBroker := pkg.Broker{
			IP:   "127.0.0.1",
			Port: "1883",
		}
		netBroker := pkg.Broker{
			IP:   serverIP,
			Port: "1883",
		}

		var clusterConnector []pkg.Connector
		clusters := GetCluster(db, cluster)
		for _, c := range clusters {
			if c.IP == switchIP {
				continue
			}
			connector := pkg.Connector{
				IP:   c.IP,
				Port: "29015",
			}
			clusterConnector = append(clusterConnector, connector)
		}
		rack := pkg.Cluster{
			Connectors: clusterConnector,
		}
		dbConnector := pkg.DBConnector{
			ClientIP:   "127.0.0.1",
			ClientPort: "28015",
			DBCluster:  rack,
		}
		cfg := pkg.ServiceConfig{
			LocalBroker:   localBroker,
			NetworkBroker: netBroker,
			DB:            dbConnector,
			LogLevel:      "INFO",
		}
		service := pkg.Service{
			Name:        serv.Name,
			Systemd:     serv.Systemd,
			Version:     serv.Version,
			PackageName: serv.PackageName,
			ConfigPath:  serv.ConfigPath,
			Config:      cfg,
		}
		services[serv.Name] = service
	}
	return services
}

//SaveServiceStatus dump service status in database
func SaveServiceStatus(db Database, status core.ServiceDump) error {
	var err error
	criteria := make(map[string]interface{})
	criteria["Name"] = status.Name
	criteria["SwitchMac"] = status.SwitchMac
	dbID := GetObjectID(db, StatusDB, ServicesTable, criteria)
	if dbID == "" {
		_, err = db.InsertRecord(StatusDB, ServicesTable, status)
	} else {
		err = db.UpdateRecord(StatusDB, ServicesTable, dbID, status)
	}
	return err
}

//RemoveServiceConfig remove sensor config in database
func RemoveServiceConfig(db Database, mac string) error {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	return db.DeleteRecord(ConfigDB, ServicesTable, criteria)
}
