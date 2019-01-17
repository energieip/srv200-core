package service

import (
	"github.com/energieip/common-components-go/pkg/dblind"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/energieip/srv200-coreservice-go/internal/history"
	"github.com/romana/rlog"
)

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

func (s *CoreService) registerSwitchStatus(switchStatus sd.SwitchStatus) {
	for _, led := range switchStatus.Leds {
		database.SaveLedStatus(s.db, led)
	}
	for _, sensor := range switchStatus.Sensors {
		database.SaveSensorStatus(s.db, sensor)
	}
	for _, blind := range switchStatus.Blinds {
		database.SaveBlindStatus(s.db, blind)
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
	setup.BlindsSetup = make(map[string]dblind.BlindSetup)

	driversMac := make(map[string]bool)
	for _, led := range switchStatus.Leds {
		driversMac[led.Mac] = true
	}
	for _, blind := range switchStatus.Blinds {
		driversMac[blind.Mac] = true
	}
	setup.Groups = database.GetGroupConfigs(s.db, driversMac)

	for mac, led := range switchStatus.Leds {
		if !led.IsConfigured {
			lsetup, _ := database.GetLedConfig(s.db, mac)
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
					FriendlyName:  &led.Mac,
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
			history.SaveLedHistory(s.historyDb, led)
			s.prepareAPIConsumption(LedElt, led.LinePower)
		}
	}

	for mac, blind := range switchStatus.Blinds {
		if !blind.IsConfigured {
			bsetup, _ := database.GetBlindConfig(s.db, mac)
			if bsetup == nil && s.installMode {
				enableBle := false
				confBlind := dblind.BlindSetup{
					Mac:           blind.Mac,
					Group:         &defaultGroup,
					IsBleEnabled:  &enableBle,
					SwitchMac:     switchStatus.Mac,
					DumpFrequency: dumpFreq,
					FriendlyName:  &blind.Mac,
				}
				bsetup = &confBlind
				// saved default config
				database.SaveBlindConfig(s.db, confBlind)
			}
			if bsetup != nil {
				setup.BlindsSetup[mac] = *bsetup
			}
			s.prepareAPIEvent(EventAdd, BlindElt, blind)
		} else {
			s.prepareAPIEvent(EventUpdate, BlindElt, blind)
		}
	}

	for mac, sensor := range switchStatus.Sensors {
		if !sensor.IsConfigured {
			ssetup, _ := database.GetSensorConfig(s.db, mac)
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
					FriendlyName:               &sensor.Mac,
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
