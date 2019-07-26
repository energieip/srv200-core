package service

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/dhvac"
	dl "github.com/energieip/common-components-go/pkg/dled"
	"github.com/energieip/common-components-go/pkg/dnanosense"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) isGroupRequiredUpdate(old gm.GroupStatus, new gm.GroupConfig) bool {
	if len(old.Leds) != len(new.Leds) || len(old.Sensors) != len(new.Sensors) ||
		len(old.Blinds) != len(new.Blinds) || len(old.FirstDay) != len(new.FirstDay) {
		return true
	}

	for i, v := range old.Leds {
		if v != new.Leds[i] {
			return true
		}
	}
	if old.RuleBrightness != nil && new.RuleBrightness != nil {
		if *old.RuleBrightness != *new.RuleBrightness {
			return true
		}
	}
	if old.RuleBrightness != nil && new.RuleBrightness == nil {
		return true
	}
	if old.RuleBrightness == nil && new.RuleBrightness != nil {
		return true
	}

	if old.FirstDayOffset != nil && new.FirstDayOffset != nil {
		if *old.FirstDayOffset != *new.FirstDayOffset {
			return true
		}
	}
	if old.FirstDayOffset != nil && new.FirstDayOffset == nil {
		return true
	}
	if old.FirstDayOffset == nil && new.FirstDayOffset != nil {
		return true
	}

	if old.RulePresence != nil && new.RulePresence != nil {
		if *old.RulePresence != *new.RulePresence {
			return true
		}
	}
	if old.RulePresence != nil && new.RulePresence == nil {
		return true
	}
	if old.RulePresence == nil && new.RulePresence != nil {
		return true
	}
	for i, v := range old.Sensors {
		if v != new.Sensors[i] {
			return true
		}
	}
	for i, v := range old.Blinds {
		if v != new.Blinds[i] {
			return true
		}
	}

	for i, v := range old.FirstDay {
		if v != new.FirstDay[i] {
			return true
		}
	}

	if new.FriendlyName != nil {
		if old.FriendlyName != *new.FriendlyName {
			return true
		}
	}
	if new.SensorRule != nil {
		if old.SensorRule != *new.SensorRule {
			return true
		}
	}
	if new.SlopeStartAuto != nil {
		if old.SlopeStartAuto != *new.SlopeStartAuto {
			return true
		}
	}

	if new.SlopeStopAuto != nil {
		if old.SlopeStopAuto != *new.SlopeStopAuto {
			return true
		}
	}
	if new.SlopeStartManual != nil {
		if old.SlopeStartManual != *new.SlopeStartManual {
			return true
		}
	}

	if new.SlopeStopManual != nil {
		if old.SlopeStopManual != *new.SlopeStopManual {
			return true
		}
	}
	if new.CorrectionInterval != nil {
		if old.CorrectionInterval != *new.CorrectionInterval {
			return true
		}
	}
	if new.Watchdog != nil {
		if old.Watchdog != *new.Watchdog {
			return true
		}
	}
	return false
}

func inArray(v interface{}, in interface{}) bool {
	val := reflect.Indirect(reflect.ValueOf(in))
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			if ok := v == val.Index(i).Interface(); ok {
				return true
			}
		}
	}
	return false
}

func (s *CoreService) sendSaveGroupCfg(cfg gm.GroupConfig) {
	oldSwitch := database.GetGroupSwitchs(s.db, cfg.Group)
	database.UpdateGroupConfig(s.db, cfg)
	newSwitch := database.GetGroupSwitchs(s.db, cfg.Group)
	for sw := range oldSwitch {
		if sw == "" {
			continue
		}
		url := "/write/switch/" + sw + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sw
		switchSetup.Groups = make(map[int]gm.GroupConfig)
		switchSetup.Groups[cfg.Group] = cfg
		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
	for sw := range newSwitch {
		if sw == "" {
			continue
		}
		url := "/write/switch/" + sw + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sw
		switchSetup.Groups = make(map[int]gm.GroupConfig)
		switchSetup.Groups[cfg.Group] = cfg
		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) updateLedGroup(mac string, grID int) {
	if mac == "" {
		return
	}
	mac = strings.ToUpper(mac)
	oldLed, _ := database.GetLedConfig(s.db, mac)
	if oldLed == nil {
		return
	}
	cfgLed := dl.LedConf{
		Mac:   mac,
		Group: &grID,
	}

	database.UpdateLedConfig(s.db, cfgLed)

	led, _ := database.GetLedConfig(s.db, mac)
	if led.SwitchMac != "" {
		url := "/write/switch/" + led.SwitchMac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = led.SwitchMac
		switchSetup.LedsConfig = make(map[string]dl.LedConf)

		switchSetup.LedsConfig[led.Mac] = cfgLed

		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
	if grID == 0 {
		//register led in group 0 == remove it from the current group
		newGr, _ := database.GetGroupConfig(s.db, grID)
		if newGr != nil {
			if inArray(mac, newGr.Leds) {
				return
			}
			newGr.Leds = append(newGr.Leds, mac)
			s.sendSaveGroupCfg(*newGr)
		}
		return
	}
	if oldLed.Group == nil || *oldLed.Group == grID {
		return
	}

	oldGr, _ := database.GetGroupConfig(s.db, *oldLed.Group)
	if oldGr != nil {
		leds := []string{}
		for _, mac := range oldGr.Leds {
			if mac != led.Mac {
				leds = append(leds, mac)
			}
		}
		oldGr.Leds = leds
		firstDays := []string{}
		for _, mac := range oldGr.FirstDay {
			if mac != led.Mac {
				firstDays = append(firstDays, mac)
			}
		}
		oldGr.FirstDay = firstDays
		s.sendSaveGroupCfg(*oldGr)
	}
}

func (s *CoreService) updateSensorGroup(mac string, grID int) {
	if mac == "" {
		return
	}
	mac = strings.ToUpper(mac)
	oldSensor, _ := database.GetSensorConfig(s.db, mac)
	if oldSensor == nil {
		return
	}
	cfgSensor := ds.SensorConf{
		Mac:   mac,
		Group: &grID,
	}
	database.UpdateSensorConfig(s.db, cfgSensor)

	sensor, _ := database.GetSensorConfig(s.db, mac)
	if sensor.SwitchMac != "" {
		url := "/write/switch/" + sensor.SwitchMac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sensor.SwitchMac
		switchSetup.SensorsConfig = make(map[string]ds.SensorConf)

		switchSetup.SensorsConfig[sensor.Mac] = cfgSensor

		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
	if grID == 0 {
		//register sensor in group 0
		newGr, _ := database.GetGroupConfig(s.db, grID)
		if newGr != nil {
			if inArray(mac, newGr.Sensors) {
				return
			}
			newGr.Sensors = append(newGr.Sensors, mac)
			s.sendSaveGroupCfg(*newGr)
		}
		return
	}
	if oldSensor.Group == nil || *oldSensor.Group == grID {
		return
	}
	oldGr, _ := database.GetGroupConfig(s.db, *oldSensor.Group)
	if oldGr != nil {
		sensors := []string{}
		for _, mac := range oldGr.Sensors {
			if mac != sensor.Mac {
				sensors = append(sensors, mac)
			}
		}
		oldGr.Sensors = sensors
		s.sendSaveGroupCfg(*oldGr)
	}
}

func (s *CoreService) updateHvacGroup(mac string, grID int) {
	if mac == "" {
		return
	}
	mac = strings.ToUpper(mac)
	old, _ := database.GetHvacConfig(s.db, mac)
	if old == nil {
		return
	}
	cfgHvac := dhvac.HvacConf{
		Mac:   mac,
		Group: &grID,
	}
	database.UpdateHvacConfig(s.db, cfgHvac)

	driver, _ := database.GetHvacConfig(s.db, mac)
	if driver.SwitchMac != "" {
		url := "/write/switch/" + driver.SwitchMac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = driver.SwitchMac
		switchSetup.HvacsConfig = make(map[string]dhvac.HvacConf)
		switchSetup.HvacsConfig[driver.Mac] = cfgHvac

		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
	if grID == 0 {
		//register hvac in group 0
		newGr, _ := database.GetGroupConfig(s.db, grID)
		if newGr != nil {
			if inArray(mac, newGr.Blinds) {
				return
			}
			newGr.Hvacs = append(newGr.Hvacs, mac)
			s.sendSaveGroupCfg(*newGr)
		}
		return
	}
	if old.Group == nil || *old.Group == grID {
		return
	}
	oldGr, _ := database.GetGroupConfig(s.db, *old.Group)
	if oldGr != nil {
		hvacs := []string{}
		for _, mac := range oldGr.Hvacs {
			if mac != driver.Mac {
				hvacs = append(hvacs, mac)
			}
		}
		oldGr.Hvacs = hvacs
		s.sendSaveGroupCfg(*oldGr)
	}
}

func (s *CoreService) updateNanoGroup(mac string, grID int) {
	if mac == "" {
		return
	}
	mac = strings.ToUpper(mac)
	old, _ := database.GetNanoConfig(s.db, mac)
	if old == nil {
		return
	}
	cfgNano := dnanosense.NanosenseConf{
		Mac:   mac,
		Group: &grID,
	}
	database.UpdateNanoConfig(s.db, cfgNano)

	driver, _ := database.GetNanoConfig(s.db, mac)
	for _, sw := range database.GetCluster(s.db, driver.Cluster) {
		if sw.Mac == nil {
			continue
		}
		url := "/write/switch/" + *sw.Mac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = *sw.Mac
		switchSetup.NanosConfig = make(map[string]dnanosense.NanosenseConf)
		switchSetup.NanosConfig[driver.Mac] = cfgNano

		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
	if grID == 0 {
		//register hvac in group 0
		newGr, _ := database.GetGroupConfig(s.db, grID)
		if newGr != nil {
			if inArray(mac, newGr.Nanosenses) {
				return
			}
			newGr.Nanosenses = append(newGr.Nanosenses, mac)
			s.sendSaveGroupCfg(*newGr)
		}
		return
	}
	if old.Group == grID {
		return
	}
	oldGr, _ := database.GetGroupConfig(s.db, old.Group)
	if oldGr != nil {
		nanos := []string{}
		for _, mac := range oldGr.Nanosenses {
			if mac != driver.Mac {
				nanos = append(nanos, mac)
			}
		}
		oldGr.Nanosenses = nanos
		s.sendSaveGroupCfg(*oldGr)
	}
}

func (s *CoreService) updateBlindGroup(mac string, grID int) {
	if mac == "" {
		return
	}
	mac = strings.ToUpper(mac)
	oldBlind, _ := database.GetBlindConfig(s.db, mac)
	if oldBlind == nil {
		return
	}
	cfgBlind := dblind.BlindConf{
		Mac:   mac,
		Group: &grID,
	}
	database.UpdateBlindConfig(s.db, cfgBlind)

	blind, _ := database.GetBlindConfig(s.db, mac)
	if blind.SwitchMac != "" {
		url := "/write/switch/" + blind.SwitchMac + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = blind.SwitchMac
		switchSetup.BlindsConfig = make(map[string]dblind.BlindConf)

		switchSetup.BlindsConfig[blind.Mac] = cfgBlind

		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
	if grID == 0 {
		//register blind in group 0
		newGr, _ := database.GetGroupConfig(s.db, grID)
		if newGr != nil {
			if inArray(mac, newGr.Blinds) {
				return
			}
			newGr.Blinds = append(newGr.Blinds, mac)
			s.sendSaveGroupCfg(*newGr)
		}
		return
	}
	if oldBlind.Group == nil || *oldBlind.Group == grID {
		return
	}
	oldGr, _ := database.GetGroupConfig(s.db, *oldBlind.Group)
	if oldGr != nil {
		blinds := []string{}
		for _, mac := range oldGr.Blinds {
			if mac != blind.Mac {
				blinds = append(blinds, mac)
			}
		}
		oldGr.Blinds = blinds
		s.sendSaveGroupCfg(*oldGr)
	}
}

func (s *CoreService) createGroup(cfg *gm.GroupConfig) {
	//Force default value
	if cfg.CorrectionInterval == nil {
		correction := 10
		cfg.CorrectionInterval = &correction
	}
	if cfg.FriendlyName == nil {
		name := "Group " + strconv.Itoa(cfg.Group)
		cfg.FriendlyName = &name
	}
	if cfg.RuleBrightness == nil {
		brigthness := 400
		cfg.RuleBrightness = &brigthness
	}
	if cfg.RulePresence == nil {
		presence := 1200
		cfg.RulePresence = &presence
	}
	if cfg.Watchdog == nil {
		watchdog := 600
		cfg.Watchdog = &watchdog
	}
	if cfg.SensorRule == nil {
		rule := "average"
		cfg.SensorRule = &rule
	}
	if cfg.FirstDayOffset == nil {
		offset := 20
		cfg.FirstDayOffset = &offset
	}
	slope := 10000
	if cfg.SlopeStartAuto == nil {
		cfg.SlopeStartAuto = &slope
	}
	if cfg.SlopeStopAuto == nil {
		cfg.SlopeStopAuto = &slope
	}
	if cfg.SlopeStartManual == nil {
		cfg.SlopeStartManual = &slope
	}
	if cfg.SlopeStopManual == nil {
		cfg.SlopeStopManual = &slope
	}
	database.SaveGroupConfig(s.db, *cfg)
}

func (s *CoreService) updateGroupCfg(config interface{}) {
	cfg, _ := gm.ToGroupConfig(config)
	if cfg == nil {
		return
	}

	old, _ := database.GetGroupConfig(s.db, cfg.Group)
	if old != nil {
		database.UpdateGroupConfig(s.db, *cfg)
		new, _ := database.GetGroupConfig(s.db, cfg.Group)
		seen := make(map[string]bool)
		for _, mac := range new.Leds {
			if mac != "" && !inArray(mac, old.Leds) {
				s.updateLedGroup(mac, cfg.Group)
			}
			seen[mac] = true
		}
		for _, mac := range old.Leds {
			_, ok := seen[mac]
			if !ok {
				s.updateLedGroup(mac, 0)
			}
		}

		seen = make(map[string]bool)
		for _, sensor := range new.Sensors {
			if sensor != "" && !inArray(sensor, old.Sensors) {
				s.updateSensorGroup(sensor, cfg.Group)
			}
			seen[sensor] = true
		}
		for _, mac := range old.Sensors {
			_, ok := seen[mac]
			if !ok {
				s.updateSensorGroup(mac, 0)
			}
		}

		seen = make(map[string]bool)
		for _, mac := range new.Blinds {
			if mac != "" && !inArray(mac, old.Blinds) {
				s.updateBlindGroup(mac, cfg.Group)
			}
			seen[mac] = true
		}
		for _, mac := range old.Blinds {
			_, ok := seen[mac]
			if !ok {
				s.updateBlindGroup(mac, 0)
			}
		}

		seen = make(map[string]bool)
		for _, mac := range new.Hvacs {
			if mac != "" && !inArray(mac, old.Hvacs) {
				s.updateHvacGroup(mac, cfg.Group)
			}
			seen[mac] = true
		}
		for _, mac := range old.Hvacs {
			_, ok := seen[mac]
			if !ok {
				s.updateHvacGroup(mac, 0)
			}
		}

		seen = make(map[string]bool)
		for _, mac := range new.Nanosenses {
			if mac != "" && !inArray(mac, old.Nanosenses) {
				s.updateNanoGroup(mac, cfg.Group)
			}
			seen[mac] = true
		}
		for _, mac := range old.Nanosenses {
			_, ok := seen[mac]
			if !ok {
				s.updateNanoGroup(mac, 0)
			}
		}

	} else {
		s.createGroup(cfg)
		for _, led := range cfg.Leds {
			s.updateLedGroup(led, cfg.Group)
		}
		for _, sensor := range cfg.Sensors {
			s.updateSensorGroup(sensor, cfg.Group)
		}
		for _, blind := range cfg.Blinds {
			s.updateBlindGroup(blind, cfg.Group)
		}
		for _, hvac := range cfg.Hvacs {
			s.updateHvacGroup(hvac, cfg.Group)
		}
		for _, nano := range cfg.Nanosenses {
			s.updateNanoGroup(nano, cfg.Group)
		}
	}

	s.sendGroupConfigUpdate(*cfg)
}

func (s *CoreService) sendGroupConfigUpdate(cfg gm.GroupConfig) {
	for sw := range database.GetGroupSwitchs(s.db, cfg.Group) {
		if sw == "" {
			continue
		}
		url := "/write/switch/" + sw + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sw
		switchSetup.Groups = make(map[int]gm.GroupConfig)
		switchSetup.Groups[cfg.Group] = cfg
		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}

func (s *CoreService) sendGroupCmd(cmd interface{}) {
	cmdGr, _ := core.ToGroupCmd(cmd)
	if cmdGr == nil {
		rlog.Error("Cannot parse cmd")
		return
	}
	for sw := range database.GetGroupSwitchs(s.db, cmdGr.Group) {
		if sw == "" {
			continue
		}
		url := "/write/switch/" + sw + "/update/settings"
		switchSetup := sd.SwitchConfig{}
		switchSetup.Mac = sw
		switchSetup.Groups = make(map[int]gm.GroupConfig)
		cfg := gm.GroupConfig{}
		cfg.Group = cmdGr.Group
		cfg.Auto = cmdGr.Auto
		cfg.SetpointLeds = cmdGr.SetpointLeds
		cfg.SetpointBlinds = cmdGr.SetpointBlinds
		cfg.SetpointSlatBlinds = cmdGr.SetpointSlats
		cfg.SetpointTempOffset = cmdGr.SetpointTempOffset
		switchSetup.Groups[cmdGr.Group] = cfg
		dump, _ := switchSetup.ToJSON()
		s.server.SendCommand(url, dump)
	}
}
