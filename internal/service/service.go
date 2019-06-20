package service

import (
	"os"

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

	UrlStatus = "status/dump"
	UrlHello  = "setup/hello"
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
	internalApi *api.InternalAPI
	bufAPI               cmap.ConcurrentMap
	bufConsumption       cmap.ConcurrentMap
	eventsConsumptionAPI chan core.EventConsumption
}

//Initialize service
func (s *CoreService) Initialize(confFile string) error {
	clientID := "GtbServer"
	s.mac, s.ip = tools.GetNetworkInfo()
	s.events = make(chan string)
	s.eventsAPI = make(chan map[string]interface{})
	s.bufAPI = cmap.New()
	s.bufConsumption = cmap.New()
	s.eventsConsumptionAPI = make(chan core.EventConsumption)

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

	err = s.server.LocalConnection(*conf, clientID)
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

	err = s.authServer.LocalConnection(*conf, clientID)
	if err != nil {
		rlog.Error("Cannot connect to drivers broker " + conf.AuthBroker.IP + " error: " + err.Error())
		return err
	}

	internal := api.InitInternalAPI(s.db, *conf)
	s.internalApi = internal

	web := api.InitAPI(s.db, s.historyDb, s.eventsAPI, s.eventsConsumptionAPI, *conf)
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
					s.updateLedCfg(event)
				case "ledSetup":
					s.updateLedSetup(event)
				case "blind":
					s.updateBlindCfg(event)
				case "blindSetup":
					s.updateBlindSetup(event)
				case "hvac":
					s.updateHvacCfg(event)
				case "hvacSetup":
					s.updateHvacSetup(event)
				case "sensor":
					s.updateSensorCfg(event)
				case "sensorSetup":
					s.updateSensorSetup(event)
				case "group":
					s.updateGroupCfg(event)
				case "switch":
					s.updateSwitchCfg(event)
				case "groupCmd":
					s.sendGroupCmd(event)
				case "ledCmd":
					s.sendLedCmd(event)
				case "blindCmd":
					s.sendBlindCmd(event)
				case "hvacCmd":
					s.sendHvacCmd(event)
				case "replaceDriver":
					s.replaceDriver(event)
				case "installDriver":
					s.installDriver(event)
				case "map":
					s.updateMapInfo(event)
				}
			}
			apiEvents = nil

		case internalAPIEvents := <-s.internalApi.EventsToBackend:
			for eventType, event := range internalAPIEvents {
				rlog.Info("get internal API event", eventType, event)
				switch eventType {
				case "groupCmd":
					s.sendGroupCmd(event)
				case "ledCmd":
					s.sendLedCmd(event)
				case "blindCmd":
					s.sendBlindCmd(event)
				case "hvacCmd":
					s.sendHvacCmd(event)
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

//Run service mainloop
func (s *CoreService) Run() error {
	go s.pushAPIEvent()
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
