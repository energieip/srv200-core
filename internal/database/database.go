package database

import (
	"github.com/energieip/common-database-go/pkg/database"
	group "github.com/energieip/common-group-go/pkg/groupmodel"
	led "github.com/energieip/common-led-go/pkg/driverled"
	sensor "github.com/energieip/common-sensor-go/pkg/driversensor"
	sdevice "github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/romana/rlog"
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

	err = db.CreateDB("status")
	if err != nil {
		rlog.Warn("Create DB ", err.Error())
	}

	err = db.CreateDB("config")
	if err != nil {
		rlog.Warn("Create DB ", err.Error())
	}

	err = db.CreateTable("config", "leds", &led.LedSetup{})
	if err != nil {
		rlog.Warn("Create table ", err.Error())
	}

	err = db.CreateTable("config", "sensors", &sensor.SensorSetup{})
	if err != nil {
		rlog.Warn("Create table ", err.Error())
	}

	err = db.CreateTable("config", "groups", &group.GroupConfig{})
	if err != nil {
		rlog.Warn("Create table ", err.Error())
	}

	err = db.CreateTable("status", "leds", &led.Led{})
	if err != nil {
		rlog.Warn("Create table ", err.Error())
	}

	err = db.CreateTable("status", "sensors", &sensor.Sensor{})
	if err != nil {
		rlog.Warn("Create table ", err.Error())
	}

	err = db.CreateTable("status", "groups", &group.GroupStatus{})
	if err != nil {
		rlog.Warn("Create table ", err.Error())
	}

	err = db.CreateTable("status", "switchs", &switchDump{})
	if err != nil {
		rlog.Warn("Create table ", err.Error())
	}

	return &db, nil
}

//GetLedConfig return the led configuration
func GetLedConfig(db Database, mac string) *led.LedSetup {
	criteria := make(map[string]interface{})
	criteria["Mac"] = mac
	ledStored, err := db.GetRecord("config", led.TableName, criteria)
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
	sensorStored, err := db.GetRecord("config", sensor.TableName, criteria)
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
	ledStored, err := db.GetRecord(led.DbName, led.TableName, criteria)
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
		_, err = db.InsertRecord(led.DbName, led.TableName, ledStatus)
	} else {
		err = db.UpdateRecord(led.DbName, led.TableName, dbID, ledStatus)
	}
	return err
}

//SaveSensorStatus dump sensor status in database
func SaveSensorStatus(db Database, sensorStatus sensor.Sensor) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Mac"] = sensorStatus.Mac
	sensorStored, err := db.GetRecord(sensor.DbName, sensor.TableName, criteria)
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
		_, err = db.InsertRecord(sensor.DbName, sensor.TableName, sensorStatus)
	} else {
		err = db.UpdateRecord(sensor.DbName, sensor.TableName, dbID, sensorStatus)
	}
	return err
}

//SaveGroupStatus dump group status in database
func SaveGroupStatus(db Database, status group.GroupStatus) error {
	var dbID string
	criteria := make(map[string]interface{})
	criteria["Group"] = status.Group
	grStored, err := db.GetRecord(group.DbStatusName, group.TableStatusName, criteria)
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
		_, err = db.InsertRecord(group.DbStatusName, group.TableStatusName, status)
	} else {
		err = db.UpdateRecord(group.DbStatusName, group.TableStatusName, dbID, status)
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
	swStored, err := db.GetRecord("status", "switchs", criteria)
	if err == nil && swStored != nil {
		m := swStored.(map[string]interface{})
		id, ok := m["id"]
		if ok {
			dbID = id.(string)
		}
	}
	if dbID == "" {
		_, err = db.InsertRecord("status", "switchs", swStatus)
	} else {
		err = db.UpdateRecord("status", "switchs", dbID, swStatus)
	}
	return err
}
