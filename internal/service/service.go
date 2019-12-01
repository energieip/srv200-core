package service

import (
	"os"
	"time"

	"github.com/energieip/common-components-go/pkg/dserver"
	"github.com/energieip/common-components-go/pkg/duser"

	sd "github.com/energieip/common-components-go/pkg/dswitch"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/common-components-go/pkg/tools"
	"github.com/energieip/srv200-coreservice-go/internal/api"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/energieip/srv200-coreservice-go/internal/history"
	"github.com/energieip/srv200-coreservice-go/internal/network"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/romana/rlog"
)

const (
	ActionReload = "reload"
	ActionSetup  = "setup"
	ActionDump   = "dump"
	ActionRemove = "remove"
)

//CoreService content
type CoreService struct {
	server               network.ServerNetwork //Local server
	authServer           network.AuthNetwork   //Authentication server
	db                   database.Database
	historyDb            history.HistoryDb
	dataPath             string
	mac                  string
	ip                   string
	events               chan string
	eventsAPI            chan map[string]interface{}
	api                  *api.API
	internalApi          *api.InternalAPI
	bufConsumption       cmap.ConcurrentMap
	eventsConsumptionAPI chan core.EventConsumption
	uploadValue          string
	timerDump            time.Duration
	switchsSeen          cmap.ConcurrentMap
}

//Initialize service
func (s *CoreService) Initialize(confFile string) error {
	s.mac, s.ip = tools.GetNetworkInfo()
	s.timerDump = 1000
	s.events = make(chan string)
	s.eventsAPI = make(chan map[string]interface{})
	s.bufConsumption = cmap.New()
	s.switchsSeen = cmap.New()
	s.eventsConsumptionAPI = make(chan core.EventConsumption)
	s.uploadValue = "none"

	conf, err := pkg.ReadServiceConfig(confFile)
	if err != nil {
		rlog.Error("Cannot parse configuration file " + err.Error())
		return err
	}
	s.dataPath = conf.DataPath
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

	historydb, err := history.ConnectDatabase(conf.HistoryDB.ClientIP, conf.HistoryDB.ClientPort)
	if err != nil {
		rlog.Error("Cannot connect to database " + err.Error())
		return err
	}
	s.historyDb = *historydb

	serverNet, err := network.CreateServerNetwork()
	if err != nil {
		rlog.Error("Cannot connect to broker " + conf.NetworkBroker.IP + " error: " + err.Error())
		return err
	}
	s.server = *serverNet

	err = s.server.LocalConnection(*conf)
	if err != nil {
		rlog.Error("Cannot connect to drivers broker " + conf.NetworkBroker.IP + " error: " + err.Error())
		return err
	}

	authNet, err := network.CreateAuthNetwork()
	if err != nil {
		rlog.Error("Cannot connect to broker " + conf.AuthBroker.IP + " error: " + err.Error())
		return err
	}
	s.authServer = *authNet

	err = s.authServer.LocalConnection(*conf)
	if err != nil {
		rlog.Error("Cannot connect to drivers broker " + conf.AuthBroker.IP + " error: " + err.Error())
		return err
	}

	internal := api.InitInternalAPI(s.db, *conf)
	s.internalApi = internal

	web := api.InitAPI(s.db, s.historyDb, s.eventsAPI, s.eventsConsumptionAPI, &s.uploadValue, *conf)
	s.api = web

	serv := dserver.ServerConfig{}
	serv.Mac = s.mac
	serv.IP = s.ip
	serv.Protocol = "MQTTS"
	dump, _ := serv.ToJSON()
	topic := "/read/server/" + s.mac + "/setup/hello"
	s.authServer.SendCommand(topic, dump)

	rlog.Info("ServerCore service started")
	return nil
}

//Stop service
func (s *CoreService) Stop() {
	rlog.Info("Stopping ServerCore service")
	s.server.Disconnect()
	s.authServer.Disconnect()
	s.db.Close()
	s.historyDb.Close()
	rlog.Info("ServerCore service stopped")
}

func (s *CoreService) readAPIEvents() {
	for {
		select {
		case apiEvents := <-s.api.EventsToBackend:
			for eventType, event := range apiEvents {
				rlog.Info("get API event", eventType, event)
				switch eventType {
				case "led":
					go s.updateLedCfg(event)
				case "ledSetup":
					go s.updateLedSetup(event)
				case "blind":
					go s.updateBlindCfg(event)
				case "blindSetup":
					go s.updateBlindSetup(event)
				case "hvac":
					go s.updateHvacCfg(event)
				case "hvacSetup":
					go s.updateHvacSetup(event)
				case "sensor":
					go s.updateSensorCfg(event)
				case "sensorSetup":
					go s.updateSensorSetup(event)
				case "wago":
					go s.updateWagoCfg(event)
				case "nano":
					go s.updateNanoCfg(event)
				case "wagoSetup":
					go s.updateWagoSetup(event)
				case "group":
					go s.updateGroupCfg(event)
				case "switch":
					go s.updateSwitchCfg(event)
				case "groupCmd":
					go s.sendGroupCmd(event)
				case "ledCmd":
					go s.sendLedCmd(event)
				case "blindCmd":
					go s.sendBlindCmd(event)
				case "hvacCmd":
					go s.sendHvacCmd(event)
				case "replaceDriver":
					go s.replaceDriver(event)
				case "installDriver":
					go s.installDriver(event)
				case "map":
					go s.updateMapInfo(event)
				}
			}
			apiEvents = nil

		case internalAPIEvents := <-s.internalApi.EventsToBackend:
			for eventType, event := range internalAPIEvents {
				rlog.Info("get internal API event", eventType, event)
				switch eventType {
				case "led":
					go s.updateLedCfg(event)
				case "ledSetup":
					go s.updateLedSetup(event)
				case "blind":
					go s.updateBlindCfg(event)
				case "blindSetup":
					go s.updateBlindSetup(event)
				case "hvac":
					go s.updateHvacCfg(event)
				case "hvacSetup":
					go s.updateHvacSetup(event)
				case "sensor":
					go s.updateSensorCfg(event)
				case "sensorSetup":
					go s.updateSensorSetup(event)
				case "wago":
					go s.updateWagoCfg(event)
				case "nano":
					go s.updateNanoCfg(event)
				case "wagoSetup":
					go s.updateWagoSetup(event)
				case "group":
					go s.updateGroupCfg(event)
				case "switch":
					go s.updateSwitchCfg(event)
				case "groupCmd":
					go s.sendGroupCmd(event)
				case "ledCmd":
					go s.sendLedCmd(event)
				case "blindCmd":
					go s.sendBlindCmd(event)
				case "hvacCmd":
					go s.sendHvacCmd(event)
				case "replaceDriver":
					go s.replaceDriver(event)
				case "installDriver":
					go s.installDriver(event)
				case "map":
					go s.updateMapInfo(event)
				}
			}
		}
	}
}

func (s *CoreService) manageMQTTEvent(eventType string, event sd.SwitchStatus) {
	switch eventType {
	case network.EventHello:
		s.sendSwitchSetup(event)
		s.registerSwitchStatus(event)
	case network.EventDump:
		s.sendSwitchUpdateConfig(event)
		s.registerSwitchStatus(event)
	}
}

func (s *CoreService) manageAuthMQTTEvent(eventType string, event duser.UserAccess) {
	switch eventType {
	case network.EventNewUser:
		s.addNewUser(event)
	case network.EventRemoveUser:
		s.removeUser(event)
	}
}

func (s *CoreService) manageAuthMQTTDumpEvent(users map[string]duser.UserAccess) {
	database.SetUsersDump(s.db, users)
}

func (s *CoreService) cleanupOldStatus() {
	timeNow := time.Now().UTC()
	toRemove := make(map[string]bool)
	switchs := database.GetSwitchsDump(s.db)
	for _, driver := range switchs {
		sw, _ := sd.ToSwitch(driver)
		val, ok := s.switchsSeen.Get(sw.Mac)
		if ok && val != nil {
			maxDuration := time.Duration(5*sw.DumpFrequency) * time.Millisecond
			if timeNow.Sub(val.(time.Time)) > maxDuration {
				rlog.Info("Switch " + sw.Mac + " timeout")
				s.switchsSeen.Remove(sw.Mac)
				toRemove[sw.Mac] = true
			}
		} else {
			toRemove[sw.Mac] = true
		}
	}
	for mac := range toRemove {
		database.RemoveSwitchStatus(s.db, mac)
		database.RemoveSwitchLedStatus(s.db, mac)
		database.RemoveSwitchBlindStatus(s.db, mac)
		database.RemoveSwitchSensorStatus(s.db, mac)
		database.RemoveSwitchHvacStatus(s.db, mac)
	}
}

func (s *CoreService) cronCleanup() {
	timerDump := time.NewTicker(s.timerDump * time.Millisecond)
	for {
		select {
		case <-timerDump.C:
			s.cleanupOldStatus()
			timerDump.Stop()
			timerDump = time.NewTicker(s.timerDump * time.Millisecond)
		}
	}
}

//Run service mainloop
func (s *CoreService) Run() error {
	go s.cronCleanup()
	go s.pushConsumptionEvent()
	go s.readAPIEvents()
	for {
		select {
		case serverEvents := <-s.server.Events:
			for eventType, event := range serverEvents {
				go s.manageMQTTEvent(eventType, event)
			}
		case authEvents := <-s.authServer.Events:
			for eventType, event := range authEvents {
				go s.manageAuthMQTTEvent(eventType, event)
			}
		case authEvents := <-s.authServer.EventDump:
			go s.manageAuthMQTTDumpEvent(authEvents)
		}
	}
}
