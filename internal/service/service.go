package service

import (
	"os"

	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
	pkg "github.com/energieip/common-service-go/pkg/service"
	"github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/energieip/srv200-coreservice-go/internal/network"
	"github.com/romana/rlog"
)

const (
	ActionReload = "reload"
	ActionSetup  = "setup"
	ActionDump   = "dump"
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

	conf, err := pkg.ReadServiceConfig(confFile)
	if err != nil {
		rlog.Error("Cannot parse configuration file " + err.Error())
		return err
	}
	os.Setenv("RLOG_LOG_LEVEL", conf.LogLevel)
	os.Setenv("RLOG_LOG_NOTIME", "yes")
	rlog.UpdateEnv()
	rlog.Info("Starting ServerCore service")

	db, err := database.ConnectDatabase(conf.DB.ClientIP, conf.DB.ClientPort)
	if err != nil {
		rlog.Error("Cannot connect to database " + err.Error())
		return err
	}
	s.db = *db

	serverNet, err := network.CreateServerNetwork()
	if err != nil {
		rlog.Error("Cannot connect to broker " + conf.NetworkBroker.IP + " error: " + err.Error())
		return err
	}
	s.server = *serverNet

	err = s.server.LocalConnection(*conf, clientID)
	if err != nil {
		rlog.Error("Cannot connect to drivers broker " + conf.NetworkBroker.IP + " error: " + err.Error())
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
	config := database.GetSwitchConfig(s.db, switchStatus.Mac)
	if config == nil && !s.installMode {
		return nil
	}
	defaultGroup := 0
	defaultWatchdog := 600

	isConfigured := true
	setup := deviceswitch.SwitchConfig{}
	setup.Mac = switchStatus.Mac
	setup.FriendlyName = config.FriendlyName
	setup.IsConfigured = &isConfigured
	setup.LedsSetup = make(map[string]driverled.LedSetup)
	setup.SensorsSetup = make(map[string]driversensor.SensorSetup)
	setup.Services = database.GetServiceConfigs(s.db)

	driversMac := make(map[string]bool)
	for _, led := range switchStatus.Leds {
		driversMac[led.Mac] = true
	}
	setup.Groups = database.GetGroupConfigs(s.db, driversMac)

	for mac, led := range switchStatus.Leds {
		if !led.IsConfigured {
			lsetup := database.GetLedConfig(s.db, mac)
			if lsetup == nil && s.installMode {
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
			setup.LedsSetup[mac] = *lsetup
		}
	}

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
			setup.SensorsSetup[mac] = *ssetup
		}
	}

	if s.installMode {
		switchSetup := core.SwitchSetup{}
		switchSetup.Mac = setup.Mac
		switchSetup.IP = switchStatus.IP
		switchSetup.Cluster = 0
		switchSetup.FriendlyName = switchStatus.FriendlyName
		database.SaveSwitchConfig(s.db, switchSetup)
	}
	if config.IP == "" {
		config.IP = switchStatus.IP
		database.SaveSwitchConfig(s.db, *config)
	}
	return &setup
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

	for _, service := range switchStatus.Services {
		serv := core.ServiceDump{}
		serv.Name = service.Name
		serv.PackageName = service.PackageName
		serv.Version = service.Version
		serv.Status = service.Status
		serv.SwitchMac = switchStatus.Mac
		database.SaveServiceStatus(s.db, serv)
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
		case installModeEvent := <-s.server.EventsInstallMode:
			s.installMode = installModeEvent
		}
	}
}
