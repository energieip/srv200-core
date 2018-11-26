package database

import (
	"github.com/energieip/common-database-go/pkg/database"
	group "github.com/energieip/common-group-go/pkg/groupmodel"
	led "github.com/energieip/common-led-go/pkg/driverled"
	sensor "github.com/energieip/common-sensor-go/pkg/driversensor"
	sdevice "github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/romana/rlog"
)

const (
	ConfigDB = "configs"
	StatusDB = "status"

	LedsTable    = "leds"
	SensorsTable = "sensors"
	GroupsTable  = "groups"
	SwitchsTable = "switchs"
)

type Database = database.DatabaseInterface

type switchDump struct {
	sdevice.Switch
	ErrorCode *int                             `json:"errorCode"`
	Services  map[string]sdevice.ServiceStatus `json:"services"`
	Leds      []string                         `json:"leds"`
	Sensors   []string                         `json:"sensors"`
	Groups    []int                            `json:"groups"`
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
		} else {
			tableCfg[LedsTable] = led.Led{}
			tableCfg[SensorsTable] = sensor.Sensor{}
			tableCfg[GroupsTable] = group.GroupStatus{}
			tableCfg[SwitchsTable] = switchDump{}
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

//GetLedConfig return the led configuration
func GetLedConfig(db Database, mac string) *led.LedSetup {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	ledStored, err := db.GetRecord(ConfigDB, LedsTable, criteria)
	if err != nil || ledStored == nil {
		return nil
	}
	light, err := led.ToLedSetup(ledStored)
	if err != nil {
		return nil
	}
	return light
}

//GetSensorConfig return the sensor configuration
func GetSensorConfig(db Database, mac string) *sensor.SensorSetup {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	sensorStored, err := db.GetRecord(ConfigDB, SensorsTable, criteria)
	if err != nil || sensorStored == nil {
		return nil
	}
	cell, err := sensor.ToSensorSetup(sensorStored)
	if err != nil {
		return nil
	}
	return cell
}

//SaveLedStatus dump led status in database
func SaveLedStatus(db Database, ledStatus led.Led) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = ledStatus.Mac
	ledStored, err := db.GetRecord(StatusDB, LedsTable, criteria)
	if err == nil && ledStored != nil {
		m := ledStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(StatusDB, LedsTable, ledStatus)
	} else {
		err = db.UpdateRecord(StatusDB, LedsTable, dbID, ledStatus)
	}
	return err
}

//SaveSensorStatus dump sensor status in database
func SaveSensorStatus(db Database, sensorStatus sensor.Sensor) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = sensorStatus.Mac
	sensorStored, err := db.GetRecord(StatusDB, SensorsTable, criteria)
	if err == nil && sensorStored != nil {
		m := sensorStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(StatusDB, SensorsTable, sensorStatus)
	} else {
		err = db.UpdateRecord(StatusDB, SensorsTable, dbID, sensorStatus)
	}
	return err
}

//SaveGroupStatus dump group status in database
func SaveGroupStatus(db Database, status group.GroupStatus) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Group"] = status.Group
	grStored, err := db.GetRecord(StatusDB, GroupsTable, criteria)
	if err == nil && grStored != nil {
		m := grStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(StatusDB, GroupsTable, status)
	} else {
		err = db.UpdateRecord(StatusDB, GroupsTable, dbID, status)
	}
	return err
}

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

//SaveLedConfig dump led config in database
func SaveLedConfig(db Database, ledStatus led.LedSetup) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = ledStatus.Mac
	ledStored, err := db.GetRecord(StatusDB, LedsTable, criteria)
	if err == nil && ledStored != nil {
		m := ledStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, LedsTable, ledStatus)
	} else {
		err = db.UpdateRecord(ConfigDB, LedsTable, dbID, ledStatus)
	}
	return err
}

//SaveSensorConfig dump sensor config in database
func SaveSensorConfig(db Database, sensorStatus sensor.SensorSetup) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = sensorStatus.Mac
	sensorStored, err := db.GetRecord(ConfigDB, SensorsTable, criteria)
	if err == nil && sensorStored != nil {
		m := sensorStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, SensorsTable, sensorStatus)
	} else {
		err = db.UpdateRecord(ConfigDB, SensorsTable, dbID, sensorStatus)
	}
	return err
}

//SaveGroupConfig dump group config in database
func SaveGroupConfig(db Database, status group.GroupConfig) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Group"] = status.Group
	grStored, err := db.GetRecord(ConfigDB, GroupsTable, criteria)
	if err == nil && grStored != nil {
		m := grStored.(map[string]interface{})
		id, ok := m["id"]
		if !ok {
			id, ok = m["ID"]
		}
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord(ConfigDB, GroupsTable, status)
	} else {
		err = db.UpdateRecord(ConfigDB, GroupsTable, dbID, status)
	}
	return err
}

//SaveServerConfig dump group status in database
func SaveServerConfig(db Database, config core.ServerConfig) error {
	var issue error
	for _, grCfg := range config.Groups {
		err := SaveGroupConfig(db, grCfg)
		if err != nil {
			issue = err
		}
	}
	for _, ledCfg := range config.Leds {
		err := SaveLedConfig(db, ledCfg)
		if err != nil {
			issue = err
		}
	}
	for _, sensorCfg := range config.Sensors {
		err := SaveSensorConfig(db, sensorCfg)
		if err != nil {
			issue = err
		}
	}
	for _, switchCfg := range config.Switchs {
		err := SaveSwitchConfig(db, switchCfg)
		if err != nil {
			issue = err
		}
	}
	return issue
}
