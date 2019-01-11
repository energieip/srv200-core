package service

import (
	"os"

	gm "github.com/energieip/common-components-go/pkg/dgroup"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/common-components-go/pkg/tools"
	"github.com/energieip/srv200-coreservice-go/internal/api"
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
	server          network.ServerNetwork //Remote server
	db              database.Database
	mac             string
	ip              string
	events          chan string
	installMode     bool
	eventsAPI       chan map[string]core.EventStatus
	eventsToBackend chan map[string]interface{}
	api             *api.API
	bufAPI          map[string]core.EventStatus
}

//Initialize service
func (s *CoreService) Initialize(confFile string) error {
	clientID := "Server"
	s.installMode = false
	s.mac, s.ip = tools.GetNetworkInfo()
	s.events = make(chan string)
	s.eventsAPI = make(chan map[string]core.EventStatus)
	s.bufAPI = make(map[string]core.EventStatus)

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
	web := api.InitAPI(s.db, s.eventsAPI, &s.installMode)
	s.api = web

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

func (s *CoreService) prepareSetupSwitchConfig(switchStatus sd.SwitchStatus) *sd.SwitchConfig {
	config := database.GetSwitchConfig(s.db, switchStatus.Mac)
	if config == nil && !s.installMode {
		return nil
	}

	isConfigured := true
	setup := sd.SwitchConfig{}
	setup.Mac = switchStatus.Mac
	setup.FriendlyName = config.FriendlyName
	setup.IsConfigured = &isConfigured
	setup.Services = database.GetServiceConfigs(s.db, switchStatus.IP, s.ip, config.Cluster)
	if s.installMode {
		switchSetup := core.SwitchConfig{}
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

func (s *CoreService) prepareSwitchConfig(switchStatus sd.SwitchStatus) *sd.SwitchConfig {
	config := database.GetSwitchConfig(s.db, switchStatus.Mac)
	if config == nil && !s.installMode {
		rlog.Warn("Cannot find configuration for switch", switchStatus.Mac)
		return nil
	}

	isConfigured := true
	setup := sd.SwitchConfig{}
	setup.Mac = switchStatus.Mac
	setup.FriendlyName = config.FriendlyName
	setup.IsConfigured = &isConfigured

	defaultGroup := 0
	dumpFreq := 1
	defaultWatchdog := 600

	setup.LedsSetup = make(map[string]dl.LedSetup)
	setup.SensorsSetup = make(map[string]ds.SensorSetup)

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
				dled := dl.LedSetup{
					Mac:           led.Mac,
					IMax:          100,
					Group:         &defaultGroup,
					Watchdog:      &defaultWatchdog,
					IsBleEnabled:  &enableBle,
					ThresholdHigh: &high,
					ThresholdLow:  &low,
					SwitchMac:     switchStatus.Mac,
					DumpFrequency: dumpFreq,
				}
				lsetup = &dled
				// saved default config
				database.SaveLedConfig(s.db, dled)
			}
			if lsetup != nil {
				setup.LedsSetup[mac] = *lsetup
			}
			s.prepareAPIEvent(EventAdd, LedElt, led)
		} else {
			s.prepareAPIEvent(EventUpdate, LedElt, led)
		}
	}

	for mac, sensor := range switchStatus.Sensors {
		if !sensor.IsConfigured {
			ssetup := database.GetSensorConfig(s.db, mac)
			if ssetup == nil && s.installMode {
				enableBle := true
				brightnessCorrection := 1
				thresholdPresence := 10
				temperatureOffset := 0
				dsensor := ds.SensorSetup{
					Mac:                        sensor.Mac,
					Group:                      &defaultGroup,
					IsBleEnabled:               &enableBle,
					BrightnessCorrectionFactor: &brightnessCorrection,
					ThresholdPresence:          &thresholdPresence,
					TemperatureOffset:          &temperatureOffset,
					SwitchMac:                  switchStatus.Mac,
					DumpFrequency:              dumpFreq,
				}
				ssetup = &dsensor
				// saved default config
				database.SaveSensorConfig(s.db, dsensor)
			}
			if ssetup != nil {
				setup.SensorsSetup[mac] = *ssetup
			}
			s.prepareAPIEvent(EventAdd, SensorElt, sensor)
		} else {
			s.prepareAPIEvent(EventUpdate, SensorElt, sensor)
		}
	}
	return &setup
}

func (s *CoreService) sendSwitchSetup(sw sd.SwitchStatus) {
	conf := s.prepareSetupSwitchConfig(sw)
	if conf == nil {
		rlog.Warn("This device " + sw.Mac + " is not authorized")
		return
	}
	switchSetup := *conf

	url := "/write/switch/" + sw.Mac + "/setup/config"
	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send setup config to " + sw.Mac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + sw.Mac + " on topic: " + url + " dump:" + dump)
	}
}

func (s *CoreService) sendSwitchUpdateConfig(sw sd.SwitchStatus) {
	conf := s.prepareSwitchConfig(sw)
	if conf == nil {
		rlog.Warn("This device " + sw.Mac + " is not authorized")
		return
	}
	switchSetup := *conf

	url := "/write/switch/" + sw.Mac + "/update/settings"
	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send update config to " + sw.Mac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + sw.Mac + " on topic: " + url + " dump:" + dump)
	}
}

func (s *CoreService) registerSwitchStatus(switchStatus sd.SwitchStatus) {
	for _, led := range switchStatus.Leds {
		database.SaveLedStatus(s.db, led)
	}
	for _, sensor := range switchStatus.Sensors {
		database.SaveSensorStatus(s.db, sensor)
	}
	for _, group := range switchStatus.Groups {
		database.SaveGroupStatus(s.db, group)
		s.prepareAPIEvent(EventUpdate, GroupElt, group)
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

func (s *CoreService) updateLedCfg(config interface{}) {
	cfg, _ := dl.ToLedConf(config)
	if cfg == nil {
		rlog.Error("Cannot parse ")
		return
	}

	oldLed := database.GetLedConfig(s.db, cfg.Mac)
	if oldLed == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	database.UpdateLedConfig(s.db, *cfg)
	//Get corresponding switchMac
	led := database.GetLedConfig(s.db, cfg.Mac)
	if led == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	if led.Group != nil {
		if oldLed.Group != led.Group {
			if oldLed.Group != nil {
				rlog.Info("Update old group", *oldLed.Group)
				gr := database.GetGroupConfig(s.db, *oldLed.Group)
				if gr != nil {
					for i, v := range gr.Leds {
						if v == led.Mac {
							gr.Leds = append(gr.Leds[:i], gr.Leds[i+1:]...)
							break
						}
					}
					rlog.Info("Old group will be ", gr.Leds)
					s.updateGroupCfg(gr)
				}
			}
			rlog.Info("Update new group", *led.Group)
			grNew := database.GetGroupConfig(s.db, *led.Group)
			if grNew != nil {
				grNew.Leds = append(grNew.Leds, cfg.Mac)
				rlog.Info("new group will be", grNew.Leds)
				s.updateGroupCfg(grNew)
			}
		}
	}
	url := "/write/switch/" + led.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = led.SwitchMac
	switchSetup.LedsConfig = make(map[string]dl.LedConf)

	switchSetup.LedsConfig[cfg.Mac] = *cfg

	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send update config to " + led.SwitchMac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + led.SwitchMac + " on topic: " + url + " dump:" + dump)
	}
}

func (s *CoreService) updateGroupCfg(config interface{}) {
	cfg, _ := gm.ToGroupConfig(config)

	gr := database.GetGroupConfig(s.db, cfg.Group)
	if gr != nil {
		database.UpdateGroupConfig(s.db, *cfg)
	} else {
		database.SaveGroupConfig(s.db, *cfg)
		for _, led := range cfg.Leds {
			light := dl.LedConf{
				Mac:   led,
				Group: &cfg.Group,
			}
			s.updateLedCfg(light)
		}
		for _, sensor := range cfg.Sensors {
			cell := ds.SensorConf{
				Mac:   sensor,
				Group: &cfg.Group,
			}
			s.updateSensorCfg(cell)
		}
	}

	for sw := range database.GetGroupSwitchs(s.db, cfg.Group) {
		url := "/write/switch/" + sw + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sw
		switchSetup.Groups = make(map[int]gm.GroupConfig)
		switchSetup.Groups[cfg.Group] = *cfg
		dump, _ := switchSetup.ToJSON()
		err := s.server.SendCommand(url, dump)
		if err != nil {
			rlog.Error("Cannot send update group config to " + sw + " on topic: " + url + " err:" + err.Error())
		} else {
			rlog.Info("Send update group config to " + sw + " on topic: " + url + " dump:" + dump)
		}
	}
}

func (s *CoreService) updateSwitchCfg(config interface{}) {
	cfg, _ := core.ToSwitchConfig(config)
	sw := database.GetSwitchConfig(s.db, cfg.Mac)
	if sw != nil {
		database.UpdateSwitchConfig(s.db, *cfg)
	} else {
		database.SaveSwitchConfig(s.db, *cfg)
	}

	url := "/write/switch/" + cfg.Mac + "/update/settings"
	switchCfg := sd.SwitchConfig{}
	switchCfg.Mac = cfg.Mac
	if cfg.DumpFrequency != nil {
		switchCfg.DumpFrequency = *cfg.DumpFrequency
	}
	switchCfg.FriendlyName = cfg.FriendlyName
	// TODO Resend service configuration if the cluster change
	if cfg.IsConfigured != nil {
		switchCfg.IsConfigured = cfg.IsConfigured
	}

	dump, _ := switchCfg.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send update config to " + cfg.Mac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + cfg.Mac + " on topic: " + url + " dump:" + dump)
	}
}

func (s *CoreService) updateSensorCfg(config interface{}) {
	cfg, _ := ds.ToSensorConf(config)

	oldSensor := database.GetSensorConfig(s.db, cfg.Mac)
	if oldSensor == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	database.UpdateSensorConfig(s.db, *cfg)
	//Get correspnding switchMac
	sensor := database.GetSensorConfig(s.db, cfg.Mac)
	if sensor == nil {
		rlog.Error("Cannot find config for " + cfg.Mac)
		return
	}

	if sensor.Group != nil {
		if oldSensor.Group != sensor.Group {
			if oldSensor.Group != nil {
				rlog.Info("Update old group", *oldSensor.Group)
				gr := database.GetGroupConfig(s.db, *oldSensor.Group)
				if gr != nil {
					for i, v := range gr.Sensors {
						if v == sensor.Mac {
							gr.Sensors = append(gr.Sensors[:i], gr.Sensors[i+1:]...)
							break
						}
					}
					rlog.Info("Old group will be ", gr.Sensors)
					s.updateGroupCfg(gr)
				}
			}
			rlog.Info("Update new group", *sensor.Group)
			grNew := database.GetGroupConfig(s.db, *sensor.Group)
			if grNew != nil {
				grNew.Sensors = append(grNew.Sensors, cfg.Mac)
				rlog.Info("new group will be", grNew.Sensors)
				s.updateGroupCfg(grNew)
			}
		}
	}

	url := "/write/switch/" + sensor.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = sensor.SwitchMac
	switchSetup.SensorsConfig = make(map[string]ds.SensorConf)

	switchSetup.SensorsConfig[cfg.Mac] = *cfg

	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send update config to " + sensor.SwitchMac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + sensor.SwitchMac + " on topic: " + url + " dump:" + dump)
	}
}

func (s *CoreService) sendGroupCmd(cmd interface{}) {
	cmdGr, _ := core.ToGroupCmd(cmd)
	if cmdGr == nil {
		rlog.Error("Cannot parse cmd")
		return
	}
	for sw := range database.GetGroupSwitchs(s.db, cmdGr.Group) {
		url := "/write/switch/" + sw + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sw
		switchSetup.Groups = make(map[int]gm.GroupConfig)
		cfg := gm.GroupConfig{}
		cfg.Group = cmdGr.Group
		cfg.Auto = &cmdGr.Auto
		cfg.SetpointLeds = &cmdGr.SetpointLeds
		switchSetup.Groups[cmdGr.Group] = cfg
		dump, _ := switchSetup.ToJSON()
		err := s.server.SendCommand(url, dump)
		if err != nil {
			rlog.Error("Cannot group command to " + sw + " on topic: " + url + " err:" + err.Error())
		} else {
			rlog.Info("Send group command to " + sw + " on topic: " + url + " dump:" + dump)
		}
	}

}

func (s *CoreService) sendLedCmd(cmd interface{}) {
	cmdLed, _ := core.ToLedCmd(cmd)
	if cmdLed == nil {
		rlog.Error("Cannot parse cmd")
		return
	}
	//Get correspnding switchMac
	led := database.GetLedConfig(s.db, cmdLed.Mac)
	if led == nil {
		rlog.Error("Cannot find config for " + cmdLed.Mac)
		return
	}
	url := "/write/switch/" + led.SwitchMac + "/update/settings"
	switchSetup := sd.SwitchConfig{}
	switchSetup.Mac = led.SwitchMac
	switchSetup.LedsConfig = make(map[string]dl.LedConf)

	auto := cmdLed.Auto
	setpoint := cmdLed.Setpoint

	ledCfg := dl.LedConf{
		Mac:            led.Mac,
		Auto:           &auto,
		SetpointManual: &setpoint,
	}
	rlog.Info("Ready to send ", ledCfg)
	rlog.Info("To switch", led.SwitchMac)
	switchSetup.LedsConfig[led.Mac] = ledCfg

	dump, _ := switchSetup.ToJSON()
	err := s.server.SendCommand(url, dump)
	if err != nil {
		rlog.Error("Cannot send update config to " + led.SwitchMac + " on topic: " + url + " err:" + err.Error())
	} else {
		rlog.Info("Send update config to " + led.SwitchMac + " on topic: " + url + " dump:" + dump)
	}

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
				}
			}
		}
	}
}

//Run service mainloop
func (s *CoreService) Run() error {
	go s.pushAPIEvent()
	go s.readAPIEvents()
	for {
		select {
		case serverEvents := <-s.server.Events:
			for eventType, event := range serverEvents {
				switch eventType {
				case network.EventHello:
					s.sendSwitchSetup(event)
					s.registerSwitchStatus(event)
				case network.EventDump:
					s.sendSwitchUpdateConfig(event)
					s.registerSwitchStatus(event)
				}
			}
		}
	}
}
