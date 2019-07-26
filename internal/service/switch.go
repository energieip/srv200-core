package service

import (
	"strings"

	"github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/dhvac"
	dl "github.com/energieip/common-components-go/pkg/dled"
	"github.com/energieip/common-components-go/pkg/dnanosense"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/common-components-go/pkg/dwago"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/energieip/srv200-coreservice-go/internal/history"
	"github.com/romana/rlog"
)

func (s *CoreService) updateSwitchCfg(config interface{}) {
	cfg, _ := core.ToSwitchConfig(config)
	if cfg == nil || cfg.Mac == nil {
		return
	}
	sw, _ := database.GetSwitchConfig(s.db, *cfg.Mac)
	if sw != nil {
		database.UpdateSwitchConfig(s.db, *cfg)
	} else {
		if cfg.Label != nil {
			sw = database.GetSwitchLabelConfig(s.db, *cfg.Label)
			if sw != nil {
				database.UpdateSwitchLabelConfig(s.db, *cfg)
			}
		}
	}
	if sw == nil {
		database.SaveSwitchConfig(s.db, *cfg)
	}

	url := "/write/switch/" + *cfg.Mac + "/update/settings"
	switchCfg := sd.SwitchConfig{}
	switchCfg.Mac = *cfg.Mac
	if cfg.DumpFrequency != nil {
		switchCfg.DumpFrequency = *cfg.DumpFrequency
	}
	if cfg.FriendlyName != nil {
		switchCfg.FriendlyName = *cfg.FriendlyName
	}
	if cfg.IsConfigured != nil {
		switchCfg.IsConfigured = cfg.IsConfigured
	}

	dump, _ := switchCfg.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) updateSwitchLabelCfg(config interface{}) {
	cfg, _ := core.ToSwitchConfig(config)
	if cfg == nil || cfg.Label == nil {
		return
	}
	sw := database.GetSwitchLabelConfig(s.db, *cfg.Label)
	if sw != nil {
		database.UpdateSwitchLabelConfig(s.db, *cfg)
	} else {
		database.SaveSwitchLabelConfig(s.db, *cfg)
	}

	mac := ""
	if cfg.Mac != nil {
		mac = *cfg.Mac
	} else {
		if sw != nil && sw.Mac != nil {
			mac = *sw.Mac
		}
	}
	mac = strings.ToUpper(mac)
	if mac == "" {
		return
	}
	url := "/write/switch/" + mac + "/update/settings"
	switchCfg := sd.SwitchConfig{}
	switchCfg.Mac = mac
	if cfg.DumpFrequency != nil {
		switchCfg.DumpFrequency = *cfg.DumpFrequency
	}
	if cfg.FriendlyName != nil {
		switchCfg.FriendlyName = *cfg.FriendlyName
	}
	if cfg.IsConfigured != nil {
		switchCfg.IsConfigured = cfg.IsConfigured
	}

	dump, _ := switchCfg.ToJSON()
	s.server.SendCommand(url, dump)
}

func (s *CoreService) registerSwitchStatus(switchStatus sd.SwitchStatus) {
	oldLeds := database.GetLedSwitchStatus(s.db, switchStatus.Mac)
	for mac, led := range switchStatus.Leds {
		database.SaveLedStatus(s.db, led)
		oldCfg, _ := database.GetLedConfig(s.db, led.Mac)
		if oldCfg != nil {
			oldCfg.SwitchMac = led.SwitchMac
			database.SaveLedConfig(s.db, *oldCfg)
			if oldCfg.Group != nil {
				gr, _ := database.GetGroupConfig(s.db, *oldCfg.Group)
				if gr != nil {
					leds := []string{}
					found := false
					for _, v := range gr.Leds {
						if v == led.Mac {
							found = true
							break
						} else {
							leds = append(leds, v)
						}
					}
					if !found {
						leds = append(leds, led.Mac)
						gr.Leds = leds
						database.SaveGroupConfig(s.db, *gr)
						s.sendGroupConfigUpdate(*gr)
					}
				}
			}
		}
		_, ok := oldLeds[mac]
		if ok {
			delete(oldLeds, mac)
		}
	}
	for _, led := range oldLeds {
		database.RemoveLedStatus(s.db, led.Mac)
		s.prepareAPIEvent(EventRemove, LedElt, led)
	}

	oldSensors := database.GetSensorSwitchStatus(s.db, switchStatus.Mac)
	for mac, sensor := range switchStatus.Sensors {
		database.SaveSensorStatus(s.db, sensor)
		oldCfg, _ := database.GetSensorConfig(s.db, sensor.Mac)
		if oldCfg != nil {
			oldCfg.SwitchMac = sensor.SwitchMac
			database.SaveSensorConfig(s.db, *oldCfg)
			if oldCfg.Group != nil {
				gr, _ := database.GetGroupConfig(s.db, *oldCfg.Group)
				if gr != nil {
					sensors := []string{}
					found := false
					for _, v := range gr.Sensors {
						if v == sensor.Mac {
							found = true
							break
						} else {
							sensors = append(sensors, v)
						}
					}
					if !found {
						sensors = append(sensors, sensor.Mac)
						gr.Sensors = sensors
						database.SaveGroupConfig(s.db, *gr)
						s.sendGroupConfigUpdate(*gr)
					}
				}
			}
		}
		_, ok := oldSensors[mac]
		if ok {
			delete(oldSensors, mac)
		}
	}
	for _, sensor := range oldSensors {
		database.RemoveSensorStatus(s.db, sensor.Mac)
		s.prepareAPIEvent(EventRemove, SensorElt, sensor)
	}

	oldBlinds := database.GetBlindSwitchStatus(s.db, switchStatus.Mac)
	for mac, blind := range switchStatus.Blinds {
		database.SaveBlindStatus(s.db, blind)
		oldCfg, _ := database.GetBlindConfig(s.db, blind.Mac)
		if oldCfg != nil {
			oldCfg.SwitchMac = blind.SwitchMac
			database.SaveBlindConfig(s.db, *oldCfg)
			if oldCfg.Group != nil {
				gr, _ := database.GetGroupConfig(s.db, *oldCfg.Group)
				if gr != nil {
					blinds := []string{}
					found := false
					for _, v := range gr.Blinds {
						if v == blind.Mac {
							found = true
							break
						} else {
							blinds = append(blinds, v)
						}
					}
					if !found {
						blinds = append(blinds, blind.Mac)
						gr.Blinds = blinds
						database.SaveGroupConfig(s.db, *gr)
						s.sendGroupConfigUpdate(*gr)
					}
				}
			}
		}
		_, ok := oldBlinds[mac]
		if ok {
			delete(oldBlinds, mac)
		}
	}
	for _, blind := range oldBlinds {
		database.RemoveBlindStatus(s.db, blind.Mac)
		s.prepareAPIEvent(EventRemove, BlindElt, blind)
	}

	oldNanos := database.GetNanoSwitchStatus(s.db, switchStatus.Cluster)
	for label, driver := range switchStatus.Nanos {
		database.SaveNanoStatus(s.db, driver)
		oldCfg, _ := database.GetNanoConfig(s.db, driver.Mac)
		if oldCfg != nil {
			gr, _ := database.GetGroupConfig(s.db, oldCfg.Group)
			if gr != nil {
				nanos := []string{}
				found := false
				for _, v := range gr.Nanosenses {
					if v == driver.Mac {
						found = true
						break
					} else {
						nanos = append(nanos, v)
					}
				}
				if !found {
					nanos = append(nanos, driver.Mac)
					gr.Nanosenses = nanos
					database.SaveGroupConfig(s.db, *gr)
					s.sendGroupConfigUpdate(*gr)
				}
			}
		}
		_, ok := oldNanos[label]
		if ok {
			delete(oldNanos, label)
		}
	}
	for _, driver := range oldNanos {
		database.RemoveNanoStatus(s.db, driver.Label)
		s.prepareAPIEvent(EventRemove, NanoElt, driver)
	}

	oldHvacs := database.GetHvacSwitchStatus(s.db, switchStatus.Mac)
	for mac, hvac := range switchStatus.Hvacs {
		database.SaveHvacStatus(s.db, hvac)
		oldCfg, _ := database.GetHvacConfig(s.db, hvac.Mac)
		if oldCfg != nil {
			oldCfg.SwitchMac = hvac.SwitchMac
			database.SaveHvacConfig(s.db, *oldCfg)
			if oldCfg.Group != nil {
				gr, _ := database.GetGroupConfig(s.db, *oldCfg.Group)
				if gr != nil {
					hvacs := []string{}
					found := false
					for _, v := range gr.Hvacs {
						if v == hvac.Mac {
							found = true
							break
						} else {
							hvacs = append(hvacs, v)
						}
					}
					if !found {
						hvacs = append(hvacs, hvac.Mac)
						gr.Hvacs = hvacs
						database.SaveGroupConfig(s.db, *gr)
						s.sendGroupConfigUpdate(*gr)
					}
				}
			}
		}
		_, ok := oldHvacs[mac]
		if ok {
			delete(oldHvacs, mac)
		}
	}
	for _, hvac := range oldHvacs {
		database.RemoveHvacStatus(s.db, hvac.Mac)
		s.prepareAPIEvent(EventRemove, HvacElt, hvac)
	}

	oldWagos := database.GetWagoClusterStatus(s.db, switchStatus.Cluster)
	for mac, wago := range switchStatus.Wagos {
		database.SaveWagoStatus(s.db, wago)
		_, ok := oldWagos[mac]
		if ok {
			delete(oldWagos, mac)
		}
	}
	for _, wago := range oldWagos {
		database.RemoveWagoStatus(s.db, wago.Mac)
		s.prepareAPIEvent(EventRemove, WagoElt, wago)
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
	s.server.SendCommand(url, dump)
}

func (s *CoreService) sendSwitchRemoveConfig(sw sd.SwitchConfig) {
	url := "/remove/switch/" + sw.Mac + "/update/settings"
	dump, _ := sw.ToJSON()
	s.server.SendCommand(url, dump)
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
	s.server.SendCommand(url, dump)
}

func (s *CoreService) prepareSetupSwitchConfig(switchStatus sd.SwitchStatus) *sd.SwitchConfig {
	config, _ := database.GetSwitchConfig(s.db, switchStatus.Mac)
	if config == nil {
		return nil
	}

	isConfigured := true
	setup := sd.SwitchConfig{}
	setup.Mac = switchStatus.Mac
	if config.FriendlyName != nil {
		setup.FriendlyName = *config.FriendlyName
	}
	setup.Cluster = config.Cluster
	setup.Label = config.Label
	setup.Profil = config.Profil
	setup.IsConfigured = &isConfigured
	setup.LedsSetup = database.GetLedSwitchSetup(s.db, switchStatus.Mac)
	setup.SensorsSetup = database.GetSensorSwitchSetup(s.db, switchStatus.Mac)
	setup.BlindsSetup = database.GetBlindSwitchSetup(s.db, switchStatus.Mac)
	setup.HvacsSetup = database.GetHvacSwitchSetup(s.db, switchStatus.Mac)
	setup.WagosSetup = database.GetWagoSwitchSetup(s.db, switchStatus.Cluster)
	setup.NanosSetup = database.GetNanoSwitchSetup(s.db, switchStatus.Cluster)
	newGroups := make(map[int]bool)

	driversMac := make(map[string]bool)
	for mac := range setup.LedsSetup {
		driversMac[mac] = true
	}
	for mac := range setup.BlindsSetup {
		driversMac[mac] = true
	}
	for mac := range setup.HvacsSetup {
		driversMac[mac] = true
	}

	setup.Groups = database.GetGroupConfigs(s.db, driversMac)
	for _, gr := range setup.Groups {
		newGroups[gr.Group] = true
	}
	setup.Users = database.GetUserConfigs(s.db, newGroups, true)

	services := make(map[string]pkg.Service)
	srv := database.GetServiceConfigs(s.db)
	for _, service := range srv {
		val, ok := switchStatus.Services[service.Name]
		if !ok || val.Version != service.Version {
			services[service.Name] = service
		}
	}

	setup.Services = services
	if config.IP == nil {
		config.IP = &switchStatus.IP
		database.SaveSwitchConfig(s.db, *config)
	}

	//Prepare Cluster
	var clusters map[string]core.SwitchConfig
	switchCluster := make(map[string]sd.SwitchCluster)
	if config.Cluster != nil {
		clusters = database.GetCluster(s.db, *config.Cluster)
		setup.WagosSetup = database.GetWagoClusterSetup(s.db, *config.Cluster)
	}
	for _, cluster := range clusters {
		if cluster.Mac == nil {
			continue
		}
		if *cluster.Mac != switchStatus.Mac {
			br := sd.SwitchCluster{
				Mac: *cluster.Mac,
			}
			if cluster.IP != nil {
				br.IP = *cluster.IP
			}
			switchCluster[*cluster.Mac] = br
		}
	}
	setup.ClusterBroker = switchCluster
	return &setup
}

func (s *CoreService) prepareSwitchConfig(switchStatus sd.SwitchStatus) *sd.SwitchConfig {
	config, _ := database.GetSwitchConfig(s.db, switchStatus.Mac)
	if config == nil {
		rlog.Warn("Cannot find configuration for switch", switchStatus.Mac)
		return nil
	}
	if config.IP == nil {
		config.IP = &switchStatus.IP
		database.SaveSwitchConfig(s.db, *config)
	}

	isConfigured := true
	setup := sd.SwitchConfig{}
	setup.Mac = switchStatus.Mac
	setup.IP = *config.IP
	if config.FriendlyName != nil {
		setup.FriendlyName = *config.FriendlyName
	}
	if config.Label != nil {
		setup.Label = config.Label
	}

	if config.Profil != "" {
		setup.Profil = config.Profil
	}
	setup.Cluster = config.Cluster
	setup.IsConfigured = &isConfigured

	setup.LedsSetup = make(map[string]dl.LedSetup)
	setup.SensorsSetup = make(map[string]ds.SensorSetup)
	setup.BlindsSetup = make(map[string]dblind.BlindSetup)
	setup.HvacsSetup = make(map[string]dhvac.HvacSetup)
	setup.WagosSetup = make(map[string]dwago.WagoSetup)

	nanos := database.GetNanoSwitchSetup(s.db, switchStatus.Cluster)
	wagos := database.GetWagoSwitchSetup(s.db, switchStatus.Cluster)
	setup.WagosSetup = make(map[string]dwago.WagoSetup)
	setup.NanosSetup = make(map[string]dnanosense.NanosenseSetup)

	grList := make(map[int]bool)

	driversMac := make(map[string]bool)
	for _, led := range switchStatus.Leds {
		driversMac[led.Mac] = true
	}
	for _, blind := range switchStatus.Blinds {
		driversMac[blind.Mac] = true
	}
	for _, hvac := range switchStatus.Hvacs {
		driversMac[hvac.Mac] = true
	}
	newGroups := make(map[int]gm.GroupConfig)
	groups := database.GetGroupConfigs(s.db, driversMac)
	for _, gr := range groups {
		old, ok := switchStatus.Groups[gr.Group]
		if ok && !s.isGroupRequiredUpdate(old, gr) {
			continue
		}
		newGroups[gr.Group] = gr
		grList[gr.Group] = true
	}
	setup.Groups = newGroups
	setup.Users = database.GetUserConfigs(s.db, grList, true)

	for mac, led := range switchStatus.Leds {
		if !led.IsConfigured {
			lsetup, _ := database.GetLedConfig(s.db, mac)
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
			if bsetup != nil {
				setup.BlindsSetup[mac] = *bsetup
			}
			s.prepareAPIEvent(EventAdd, BlindElt, blind)
		} else {
			s.prepareAPIEvent(EventUpdate, BlindElt, blind)
			history.SaveBlindHistory(s.historyDb, blind)
			s.prepareAPIConsumption(BlindElt, blind.LinePower)
		}
	}

	for mac, hvac := range switchStatus.Hvacs {
		if !hvac.IsConfigured {
			bsetup, _ := database.GetHvacConfig(s.db, mac)
			if bsetup != nil {
				setup.HvacsSetup[mac] = *bsetup
			}
			s.prepareAPIEvent(EventAdd, HvacElt, hvac)
		} else {
			s.prepareAPIEvent(EventUpdate, HvacElt, hvac)
			history.SaveHvacHistory(s.historyDb, hvac)
			s.prepareAPIConsumption(HvacElt, hvac.LinePower)
		}
	}

	for mac, sensor := range switchStatus.Sensors {
		if !sensor.IsConfigured {
			ssetup, _ := database.GetSensorConfig(s.db, mac)
			if ssetup != nil {
				setup.SensorsSetup[mac] = *ssetup
			}
			s.prepareAPIEvent(EventAdd, SensorElt, sensor)
		} else {
			s.prepareAPIEvent(EventUpdate, SensorElt, sensor)
		}
	}

	wagoSeen := make(map[string]bool)
	for mac, wago := range switchStatus.Wagos {
		_, ok := wagos[mac]
		if !ok {
			s.prepareAPIEvent(EventAdd, WagoElt, wago)
		} else {
			s.prepareAPIEvent(EventUpdate, WagoElt, wago)
		}
		wagoSeen[mac] = true
	}
	for mac, driver := range wagos {
		_, ok := wagoSeen[mac]
		if ok {
			continue
		}
		setup.WagosSetup[mac] = driver
	}

	nanoSeen := make(map[string]bool)
	for _, driver := range switchStatus.Nanos {
		_, ok := nanos[driver.Label]
		if !ok {
			s.prepareAPIEvent(EventAdd, NanoElt, driver)
		} else {
			s.prepareAPIEvent(EventUpdate, NanoElt, driver)
		}
		nanoSeen[driver.Label] = true
	}
	for mac, driver := range nanos {
		_, ok := nanoSeen[driver.Label]
		if ok {
			continue
		}
		setup.NanosSetup[mac] = driver
	}

	//Prepare Cluster
	var clusters map[string]core.SwitchConfig
	switchCluster := make(map[string]sd.SwitchCluster)
	if config.Cluster != nil {
		clusters = database.GetCluster(s.db, *config.Cluster)
		for _, cluster := range clusters {
			if cluster.Mac == nil {
				continue
			}
			_, ok := switchStatus.ClusterBroker[*cluster.Mac]
			if !ok {
				//add only new cluster member only
				if *cluster.Mac != switchStatus.Mac {
					br := sd.SwitchCluster{
						Mac: *cluster.Mac,
					}
					if cluster.IP != nil {
						br.IP = *cluster.IP
					}
					switchCluster[*cluster.Mac] = br
				}
			}
		}
	}
	setup.ClusterBroker = switchCluster
	return &setup
}
