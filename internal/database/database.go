package database

import (
	"github.com/energieip/common-database-go/pkg/database"
	group "github.com/energieip/common-group-go/pkg/groupmodel"
	led "github.com/energieip/common-led-go/pkg/driverled"
	sensor "github.com/energieip/common-sensor-go/pkg/driversensor"
	pkg "github.com/energieip/common-service-go/pkg/service"
	sdevice "github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/romana/rlog"
)

const (
	ConfigDB = "configs"
	StatusDB = "status"

	LedsTable     = "leds"
	SensorsTable  = "sensors"
	GroupsTable   = "groups"
	SwitchsTable  = "switchs"
	ServicesTable = "services"
)

type Database = database.DatabaseInterface

type switchDump struct {
	sdevice.Switch
	ErrorCode *int `json:"errorCode"`
}

//ConnectDatabase plug datbase
func ConnectDatabase(ip, port string) (*Database, error) {
	db, err := database.NewDatabase(database.RETHINKDB)
	if err != nil {
		rlog.Error("database err " + err.Error())
		return nil, err
	}

	confDb := database.DatabaseConfig{
		IP:   ip,
		Port: port,
	}
	err = db.Initialize(confDb)
	if err != nil {
		rlog.Error("Cannot connect to database " + err.Error())
		return nil, err
	}

	for _, dbName := range []string{ConfigDB, StatusDB} {
		err = db.CreateDB(dbName)
		if err != nil {
			rlog.Warn("Create DB ", err.Error())
		}

		tableCfg := make(map[string]interface{})
		if dbName == ConfigDB {
			tableCfg[LedsTable] = led.LedSetup{}
			tableCfg[SensorsTable] = sensor.SensorSetup{}
			tableCfg[GroupsTable] = group.GroupConfig{}
			tableCfg[SwitchsTable] = core.SwitchSetup{}
			tableCfg[ServicesTable] = pkg.Service{}
		} else {
			tableCfg[LedsTable] = led.Led{}
			tableCfg[SensorsTable] = sensor.Sensor{}
			tableCfg[GroupsTable] = group.GroupStatus{}
			tableCfg[SwitchsTable] = switchDump{}
			tableCfg[ServicesTable] = pkg.ServiceStatus{}
		}
		for tableName, objs := range tableCfg {
			err = db.CreateTable(dbName, tableName, &objs)
			if err != nil {
				rlog.Warn("Create table ", err.Error())
			}
		}
	}

	return &db, nil
}
