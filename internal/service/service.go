package service

import (
	"os"

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
	server               network.ServerNetwork //Remote server
	db                   database.Database
	historyDb            history.HistoryDb
	mac                  string
	ip                   string
	events               chan string
	installMode          bool
	eventsAPI            chan map[string]interface{}
	eventsToBackend      chan map[string]interface{}
	api                  *api.API
	bufAPI               cmap.ConcurrentMap
	bufConsumption       cmap.ConcurrentMap
	eventsConsumptionAPI chan core.EventConsumption
}

//Initialize service
func (s *CoreService) Initialize(confFile string) error {
	clientID := "Server"
	s.installMode = false
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
	web := api.InitAPI(s.db, s.historyDb, s.eventsAPI, s.eventsConsumptionAPI, &s.installMode, *conf)
	s.api = web

	rlog.Info("ServerCore service started")
	return nil
}

//Stop service
func (s *CoreService) Stop() {
	rlog.Info("Stopping ServerCore service")
	s.server.Disconnect()
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
				case "blind":
					s.updateBlindCfg(event)
				case "hvac":
					s.updateHvacCfg(event)
				case "sensor":
					s.updateSensorCfg(event)
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
				}
			}
			apiEvents = nil
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
		}
	}
}
