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
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Name"] = serv.Name
	stored, err := db.GetRecord(ConfigDB, ServicesTable, criteria)
	if err == nil && stored != nil {
		m := stored.(map[string]interface{})
		id, ok := m["id"]
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, ServicesTable, serv)
	} else {
		err = db.UpdateRecord(ConfigDB, ServicesTable, dbID, serv)
	}
	return err
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
