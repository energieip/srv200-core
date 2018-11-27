package service

import (
	"os"

	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
	"github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/energieip/srv200-coreservice-go/internal/network"
	"github.com/energieip/srv200-coreservice-go/pkg/config"
	"github.com/romana/rlog"
)

const (
	ActionReload = "ReloadConfig"
	ActionSetup  = "Setup"
	ActionDump   = "DumpStatus"
	ActionRemove = "remove"

	UrlStatus = "status/dump"
	UrlHello  = "setup/hello"
)

//CoreService content
type CoreService struct {
	server      network.ServerNetwork //Remote server
	db          database.Database
	mac         string //Switch mac address
	events      chan string
	installMode bool
}

//Initialize service
func (s *CoreService) Initialize(confFile string) error {
	clientID := "Server"
	s.installMode = false
	s.events = make(chan string)

	conf, err := config.ReadConfig(confFile)
	if err != nil {
		rlog.Error("Cannot parse configuration file " + err.Error())
		return err
	}
	os.Setenv("RLOG_LOG_LEVEL", *conf.LogLevel)
	os.Setenv("RLOG_LOG_NOTIME", "yes")
	rlog.UpdateEnv()
	rlog.Info("Starting ServerCore service")

	db, err := database.ConnectDatabase(conf.DatabaseIP, conf.DatabasePort)
	if err != nil {
		rlog.Error("Cannot connect to database " + err.Error())
		return err
	}
	s.db = *db

	serverNet, err := network.CreateServerNetwork()
	if err != nil {
		rlog.Error("Cannot connect to broker " + conf.ServerIP + " error: " + err.Error())
		return err
	}
	s.server = *serverNet

	err = s.server.LocalConnection(*conf, clientID)
	if err != nil {
		rlog.Error("Cannot connect to drivers broker " + conf.ServerIP + " error: " + err.Error())
		return err
	}

	rlog.Info("ServerCore service started")
	return nil
}

//Stop service
func (s *CoreService) Stop() {
	rlog.Info("Stopping ServerCore service")
	s.server.Disconnect()
	s.db.Close()
	rlog.Info("ServerCore service stopped")
}

func (s *CoreService) prepareSwitchConfig(switchStatus deviceswitch.SwitchStatus) *deviceswitch.SwitchConfig {
	var conf *deviceswitch.SwitchConfig
	if s.installMode {
		defaultGroup := 0
		defaultWatchdog := 600

		var setup deviceswitch.SwitchConfig
		config := database.GetSwitchConfig(s.db, switchStatus.Mac)
		if config != nil {
			setup = *config
		} else {
			isConfigured := true
			setup = deviceswitch.SwitchConfig{}
			setup.Mac = switchStatus.Mac
			setup.IsConfigured = &isConfigured

			switchSetup := core.SwitchSetup{}
			switchSetup.Mac = setup.Mac
			switchSetup.IP = switchStatus.IP
			switchSetup.Protocol = switchStatus.Protocol
			switchSetup.Topic = switchStatus.Topic
			database.SaveSwitchConfig(s.db, switchSetup)
		}

		var ledsSetup map[string]driverled.LedSetup
		var sensorsSetup map[string]driversensor.SensorSetup

		for mac, led := range switchStatus.Leds {
			if !led.IsConfigured {
				lsetup := database.GetLedConfig(s.db, mac)
				if lsetup == nil {
					enableBle := false
					low := 0
					high := 100
					dled := driverled.LedSetup{
						Mac:          led.Mac,
						IMax:         100,
						Group:        &defaultGroup,
						Watchdog:     &defaultWatchdog,
						IsBleEnabled: &enableBle,
						ThresoldHigh: &high,
						ThresoldLow:  &low,
					}
					lsetup = &dled
					// saved default config
					database.SaveLedConfig(s.db, dled)
				}
				ledsSetup[mac] = *lsetup
			}
		}
		setup.LedsSetup = ledsSetup

		for mac, sensor := range switchStatus.Sensors {
			if !sensor.IsConfigured {
				ssetup := database.GetSensorConfig(s.db, mac)
				if ssetup == nil {
					enableBle := true
					brightnessCorrection := 1
					thresoldPresence := 10
					temperatureOffset := 0
					dsensor := driversensor.SensorSetup{
						Mac:                        sensor.Mac,
						Group:                      &defaultGroup,
						IsBleEnabled:               &enableBle,
						BrigthnessCorrectionFactor: &brightnessCorrection,
						ThresoldPresence:           &thresoldPresence,
						TemperatureOffset:          &temperatureOffset,
					}
					ssetup = &dsensor
					// saved default config
					database.SaveSensorConfig(s.db, dsensor)
				}
				sensorsSetup[mac] = *ssetup
			}
		}
		setup.SensorsSetup = sensorsSetup
		conf = &setup

	} else {
		// standard mode
		conf = database.GetSwitchConfig(s.db, switchStatus.Mac)
	}
	return conf
}

func (s *CoreService) sendSwitchSetup(switchStatus deviceswitch.SwitchStatus) {
	conf := s.prepareSwitchConfig(switchStatus)
	if conf == nil {
		rlog.Warn("This device " + switchStatus.Mac + " is not authorized")
		return
	}
	switchSetup := *conf

	url := "/write/" + switchStatus.Topic + "/setup/config"
	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) sendSwitchUpdateConfig(sw deviceswitch.SwitchStatus) {
	conf := s.prepareSwitchConfig(sw)
	if conf == nil {
		rlog.Warn("This device " + sw.Mac + " is not authorized")
		return
	}
	switchSetup := *conf

	url := "/write/" + sw.Topic + "/update/settings"
	dump, _ := switchSetup.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) sendSwitchCommand() {

}

func (s *CoreService) registerSwitchStatus(switchStatus deviceswitch.SwitchStatus) {
	for _, led := range switchStatus.Leds {
		database.SaveLedStatus(s.db, led)
	}
	for _, sensor := range switchStatus.Sensors {
		database.SaveSensorStatus(s.db, sensor)
	}
	for _, group := range switchStatus.Groups {
		database.SaveGroupStatus(s.db, group)
	}
	database.SaveSwitchStatus(s.db, switchStatus)
}

func (s *CoreService) registerConfig(config core.ServerConfig) {
	database.SaveServerConfig(s.db, config)
}

//Run service mainloop
func (s *CoreService) Run() error {
	for {
		select {
		case serverEvents := <-s.server.Events:
			for eventType, event := range serverEvents {
				switch eventType {
				case network.EventHello:
					s.sendSwitchSetup(event)
					s.registerSwitchStatus(event)

				case network.EventDump:
					s.registerSwitchStatus(event)
					s.sendSwitchUpdateConfig(event)
				}
			}
		case configEvents := <-s.server.EventsCfg:
			for eventType, event := range configEvents {
				switch eventType {
				case network.EventWriteCfg:
					s.registerConfig(event)
				}
			}
		}
	}
}
