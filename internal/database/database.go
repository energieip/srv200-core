package database

import (
	"github.com/energieip/common-components-go/pkg/database"
	"github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/dhvac"
	dl "github.com/energieip/common-components-go/pkg/dled"
	"github.com/energieip/common-components-go/pkg/dnanosense"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/common-components-go/pkg/dwago"
	"github.com/energieip/common-components-go/pkg/pconst"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/romana/rlog"
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

	for _, dbName := range []string{pconst.DbConfig, pconst.DbStatus} {
		err = db.CreateDB(dbName)
		if err != nil {
			rlog.Warn("Create DB ", err.Error())
		}

		tableCfg := make(map[string]interface{})
		if dbName == pconst.DbConfig {
			tableCfg[pconst.TbLeds] = dl.LedSetup{}
			tableCfg[pconst.TbSensors] = ds.SensorSetup{}
			tableCfg[pconst.TbHvacs] = dhvac.HvacSetup{}
			tableCfg[pconst.TbGroups] = gm.GroupConfig{}
			tableCfg[pconst.TbSwitchs] = core.SwitchConfig{}
			tableCfg[pconst.TbServices] = pkg.Service{}
			tableCfg[pconst.TbModels] = core.Model{}
			tableCfg[pconst.TbProjects] = core.Project{}
			tableCfg[pconst.TbBlinds] = dblind.BlindSetup{}
			tableCfg[pconst.TbAccess] = duser.UserAccess{}
			tableCfg[pconst.TbFrames] = core.Frame{}
			tableCfg[pconst.TbWagos] = dwago.WagoSetup{}
			tableCfg[pconst.TbNanosenses] = dnanosense.NanosenseSetup{}
		} else {
			tableCfg[pconst.TbLeds] = dl.Led{}
			tableCfg[pconst.TbSensors] = ds.Sensor{}
			tableCfg[pconst.TbHvacs] = dhvac.Hvac{}
			tableCfg[pconst.TbGroups] = gm.GroupStatus{}
			tableCfg[pconst.TbSwitchs] = core.SwitchDump{}
			tableCfg[pconst.TbServices] = pkg.ServiceStatus{}
			tableCfg[pconst.TbBlinds] = dblind.Blind{}
			tableCfg[pconst.TbWagos] = dwago.Wago{}
			tableCfg[pconst.TbNanosenses] = dnanosense.Nanosense{}
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
