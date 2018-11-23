package service

import (
	"os"

	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
	"github.com/energieip/common-switch-go/pkg/deviceswitch"
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
	server network.ServerNetwork //Remote server
	db     database.Database
	mac    string //Switch mac address
	events chan string
}

//Initialize service
func (s *CoreService) Initialize(confFile string) error {
	clientID := "Server"
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

func (s *CoreService) sendSwitchSetup(switchStatus deviceswitch.SwitchStatus) {
	switchSetup := deviceswitch.SwitchConfig{}
	var ledsSetup map[string]driverled.LedSetup
	var sensorsSetup map[string]driversensor.SensorSetup

	for mac, led := range switchStatus.Leds {
		if !led.IsConfigured {
			lsetup := database.GetLedConfig(s.db, mac)
			if lsetup == nil {
				continue
			}
			ledsSetup[mac] = *lsetup
		}
	}
	switchSetup.LedsSetup = ledsSetup

	for mac, sensor := range switchStatus.Sensors {
		if !sensor.IsConfigured {
			ssetup := database.GetSensorConfig(s.db, mac)
			if ssetup == nil {
				continue
			}
			sensorsSetup[mac] = *ssetup
		}
	}
	switchSetup.SensorsSetup = sensorsSetup

	url := "/write/" + switchStatus.Topic + "/setup/config"
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
				}
			}
		}
	}
}
