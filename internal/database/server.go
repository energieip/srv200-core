package database

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

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
	for _, serviceCfg := range config.Services {
		err := SaveServiceConfig(db, serviceCfg)
		if err != nil {
			issue = err
		}
	}
	for _, model := range config.Models {
		err := SaveModel(db, model)
		if err != nil {
			issue = err
		}
	}
	for _, project := range config.Projects {
		err := SaveProject(db, project)
		if err != nil {
			issue = err
		}
	}
	return issue
}
