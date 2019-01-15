package database

import (
	"github.com/energieip/common-components-go/pkg/database"
	"github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/romana/rlog"
)

const (
	ConfigDB = "config"
	StatusDB = "status"

	LedsTable     = "leds"
	BlindsTable   = "blinds"
	SensorsTable  = "sensors"
	GroupsTable   = "groups"
	SwitchsTable  = "switchs"
	ServicesTable = "services"
	ModelsTable   = "models"
	ProjectsTable = "projects"
)

type databaseError struct {
	s string
}

func (e *databaseError) Error() string {
	return e.s
}

// NewError raise an error
func NewError(text string) error {
	return &databaseError{text}
}

type Database = database.DatabaseInterface

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
			tableCfg[LedsTable] = dl.LedSetup{}
			tableCfg[SensorsTable] = ds.SensorSetup{}
			tableCfg[GroupsTable] = gm.GroupConfig{}
			tableCfg[SwitchsTable] = core.SwitchConfig{}
			tableCfg[ServicesTable] = pkg.Service{}
			tableCfg[ModelsTable] = core.Model{}
			tableCfg[ProjectsTable] = core.Project{}
			tableCfg[BlindsTable] = dblind.BlindSetup{}
		} else {
			tableCfg[LedsTable] = dl.Led{}
			tableCfg[SensorsTable] = ds.Sensor{}
			tableCfg[GroupsTable] = gm.GroupStatus{}
			tableCfg[SwitchsTable] = core.SwitchDump{}
			tableCfg[ServicesTable] = pkg.ServiceStatus{}
			tableCfg[BlindsTable] = dblind.Blind{}
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

//GetObjectID return id
func GetObjectID(db Database, dbName, tbName string, criteria map[string]interface{}) string {
	stored, err := db.GetRecord(dbName, tbName, criteria)
	if err == nil && stored != nil {
		m := stored.(map[string]interface{})
		id, ok := m["id"]
		if ok {
			return id.(string)
		}
	}
	return ""
}

//SaveOnUpdateObject in database
func SaveOnUpdateObject(db Database, obj interface{}, dbName, tbName string, criteria map[string]interface{}) error {
	var err error
	dbID := GetObjectID(db, dbName, tbName, criteria)
	if dbID == "" {
		_, err = db.InsertRecord(dbName, tbName, obj)
	} else {
		err = db.UpdateRecord(dbName, tbName, dbID, obj)
	}
	return err
}
