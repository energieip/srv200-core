package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"io/ioutil"
	"strings"

	"github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/dhvac"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	"github.com/energieip/common-components-go/pkg/pconst"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"

	"github.com/energieip/srv200-coreservice-go/internal/core"
)

type InternalAPI struct {
	db              database.Database
	EventsToBackend chan map[string]interface{}
	certificate     string
	keyfile         string
	apiIP           string
	apiPort         string
	apiPassword     string
	browsingFolder  string
	dataPath        string
}

//InitInternalAPI start API connection
func InitInternalAPI(db database.Database,
	conf pkg.ServiceConfig) *InternalAPI {
	api := InternalAPI{
		db:              db,
		apiIP:           conf.InternalAPI.IP,
		apiPort:         conf.InternalAPI.Port,
		apiPassword:     conf.InternalAPI.Password,
		EventsToBackend: make(chan map[string]interface{}),
		certificate:     conf.InternalAPI.CertPath,
		keyfile:         conf.InternalAPI.KeyPath,
		browsingFolder:  conf.InternalAPI.BrowsingFolder,
		dataPath:        conf.DataPath,
	}
	go api.swagger()
	return &api
}

func (api *InternalAPI) setDefaultHeader(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
}

func (api *InternalAPI) sendError(w http.ResponseWriter, errorCode int, message string, httpStatus int) {
	errCode := APIError{
		Code:    errorCode,
		Message: message,
	}

	inrec, _ := json.MarshalIndent(errCode, "", "  ")
	rlog.Error(errCode.Message)
	http.Error(w, string(inrec), httpStatus)
}

func (api *InternalAPI) getFunctions(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w, req)
	functions := []string{"/versions"}
	apiInfo := APIFunctions{
		Functions: functions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *InternalAPI) getAPIs(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w, req)
	versions := []string{"v1.0"}
	apiInfo := APIInfo{
		Versions: versions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *InternalAPI) getV1Functions(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w, req)
	apiV1 := "/v1.0"
	functions := []string{
		apiV1 + "/command/led", apiV1 + "/command/blind", apiV1 + "/command/hvac",
		apiV1 + "/command/group", apiV1 + "/dump",
	}
	apiInfo := APIFunctions{
		Functions: functions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *InternalAPI) getDump(w http.ResponseWriter, req *http.Request) {
	var leds []DumpLed
	var sensors []DumpSensor
	var switchs []DumpSwitch
	var blinds []DumpBlind
	var hvacs []DumpHvac
	var wagos []DumpWago
	var frames []DumpFrame
	var groups []DumpGroup
	var nanos []DumpNanosense
	macs := make(map[string]bool)
	driversMac := make(map[string]bool)
	labels := make(map[string]bool)
	filterByMac := false
	MacsParam := req.FormValue("macs")
	if MacsParam != "" {
		tempMac := strings.Split(MacsParam, ",")
		for _, v := range tempMac {
			macs[strings.ToUpper(v)] = true
			filterByMac = true
		}
	}

	filterByLabel := false
	LabelsParam := req.FormValue("labels")
	if LabelsParam != "" {
		for _, v := range strings.Split(LabelsParam, ",") {
			labels[v] = true
			filterByLabel = true
		}
	}

	lights := database.GetLedsStatusByLabel(api.db)
	lightsConfig := database.GetLedsConfigByLabel(api.db)
	cells := database.GetSensorsStatusByLabel(api.db)
	cellsConfig := database.GetSensorsConfigByLabel(api.db)
	blds := database.GetBlindsStatusByLabel(api.db)
	bldsConfig := database.GetBlindsConfigByLabel(api.db)
	hvcs := database.GetHvacsStatusByLabel(api.db)
	hvcsConfig := database.GetHvacsConfigByLabel(api.db)
	wags := database.GetWagosStatusByLabel(api.db)
	wagosConfig := database.GetWagosConfigByLabel(api.db)
	switchElts := database.GetSwitchsDumpByLabel(api.db)
	switchEltsConfig := database.GetSwitchsConfigByLabel(api.db)
	frameElts := database.GetFramesDumpByLabel(api.db)
	frameEltsConfig := database.GetFramesConfigByLabel(api.db)
	nans := database.GetNanosStatusByLabel(api.db)
	nanosConfig := database.GetNanosConfigByLabel(api.db)

	ifcs := database.GetIfcs(api.db)
	for _, ifc := range ifcs {
		if filterByLabel {
			if _, ok := labels[ifc.Label]; !ok {
				continue
			}
		}
		if filterByMac {
			if _, ok := macs[ifc.Mac]; !ok {
				continue
			}
		}
		driversMac[ifc.Mac] = true

		switch ifc.DeviceType {
		case pconst.LED:
			dump := DumpLed{}
			led, ok := lights[ifc.Label]
			if ok {
				dump.Status = led
			}
			config, ok := lightsConfig[ifc.Label]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc

			leds = append(leds, dump)
		case pconst.SENSOR:
			dump := DumpSensor{}
			sensor, ok := cells[ifc.Label]
			if ok {
				dump.Status = sensor
			}
			config, ok := cellsConfig[ifc.Label]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			sensors = append(sensors, dump)
		case pconst.BLIND:
			dump := DumpBlind{}
			bld, ok := blds[ifc.Label]
			if ok {
				dump.Status = bld
			}
			config, ok := bldsConfig[ifc.Label]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			blinds = append(blinds, dump)
		case pconst.HVAC:
			dump := DumpHvac{}
			hvac, ok := hvcs[ifc.Label]
			if ok {
				dump.Status = hvac
			}
			config, ok := hvcsConfig[ifc.Label]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			hvacs = append(hvacs, dump)
		case pconst.WAGO:
			dump := DumpWago{}
			wago, ok := wags[ifc.Label]
			if ok {
				dump.Status = wago
			}
			config, ok := wagosConfig[ifc.Label]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			wagos = append(wagos, dump)
		case pconst.SWITCH:
			dump := DumpSwitch{}
			switchElt, ok := switchElts[ifc.Label]
			if ok {
				dump.Status = switchElt
			}
			config, ok := switchEltsConfig[ifc.Label]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			switchs = append(switchs, dump)
		case pconst.FRAME:
			dump := DumpFrame{}
			frameElt, ok := frameElts[ifc.Label]
			if ok {
				dump.Status = frameElt
			}
			config, ok := frameEltsConfig[ifc.Label]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			frames = append(frames, dump)
		case pconst.NANOSENSE:
			dump := DumpNanosense{}
			gr := 0
			nano, ok := nans[ifc.Label]
			if ok {
				dump.Status = nano
				gr = nano.Group
			}
			config, ok := nanosConfig[ifc.Label]
			if ok {
				dump.Config = config
				if gr == 0 {
					gr = config.Group
				}
			}
			dump.Ifc = ifc
			nanos = append(nanos, dump)
		}
	}

	groupsStatus := database.GetGroupsStatus(api.db)
	groupsConfig := database.GetGroupConfigs(api.db, driversMac)

	for _, gr := range groupsConfig {
		dump := DumpGroup{}
		grStatus, ok := groupsStatus[gr.Group]
		if ok {
			dump.Status = grStatus
		}
		dump.Config = gr
		groups = append(groups, dump)
	}

	dump := Dump{
		Leds:       leds,
		Sensors:    sensors,
		Blinds:     blinds,
		Hvacs:      hvacs,
		Wagos:      wagos,
		Switchs:    switchs,
		Groups:     groups,
		Frames:     frames,
		Nanosenses: nanos,
	}

	inrec, _ := json.MarshalIndent(dump, "", "  ")
	w.Write(inrec)
}

func (api *InternalAPI) sendLedCommand(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	led := core.LedCmd{}
	err = json.Unmarshal([]byte(body), &led)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	led.Mac = strings.ToUpper(led.Mac)
	dr, _ := database.GetLedConfig(api.db, led.Mac)
	if dr == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+led.Mac+" not found", http.StatusInternalServerError)
		return
	}

	rlog.Info("Received led cmd", led)
	event := make(map[string]interface{})
	event["ledCmd"] = led
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) sendBlindCommand(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	cmd := core.BlindCmd{}
	err = json.Unmarshal([]byte(body), &cmd)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	cmd.Mac = strings.ToUpper(cmd.Mac)
	dr, _ := database.GetBlindConfig(api.db, cmd.Mac)
	if dr == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+cmd.Mac+" not found", http.StatusInternalServerError)
		return
	}

	rlog.Info("Received Blind cmd", cmd)
	event := make(map[string]interface{})
	event["blindCmd"] = cmd
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) sendHvacCommand(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	cmd := core.HvacCmd{}
	err = json.Unmarshal([]byte(body), &cmd)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	dr, _ := database.GetHvacConfig(api.db, cmd.Mac)
	if dr == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+cmd.Mac+" not found", http.StatusInternalServerError)
		return
	}
	cmd.Mac = strings.ToUpper(cmd.Mac)

	rlog.Info("Received Hvac cmd", cmd)
	event := make(map[string]interface{})
	event["hvacCmd"] = cmd
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) sendGroupCommand(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	gr := core.GroupCmd{}
	err = json.Unmarshal(body, &gr)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}

	event := make(map[string]interface{})
	event["groupCmd"] = gr
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) setLedConfig(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	led := dl.LedConf{}
	err = json.Unmarshal(body, &led)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	led.Mac = strings.ToUpper(led.Mac)

	if led.Group != nil {
		if *led.Group < 0 {
			api.sendError(w, APIErrorInvalidValue, "Invalid groupID "+strconv.Itoa(*led.Group), http.StatusInternalServerError)
			return
		}
		gr, _ := database.GetGroupConfig(api.db, *led.Group)
		if gr == nil {
			name := "Group " + strconv.Itoa(*led.Group)
			group := gm.GroupConfig{
				Group:        *led.Group,
				FriendlyName: &name,
			}
			if led.Mac != "" {
				leds := []string{led.Mac}
				group.Leds = leds
			}
			eventGr := make(map[string]interface{})
			eventGr["group"] = group
			api.EventsToBackend <- eventGr
		}
	}
	event := make(map[string]interface{})
	event["led"] = led
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) setSensorConfig(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	sensor := ds.SensorConf{}
	err = json.Unmarshal(body, &sensor)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	sensor.Mac = strings.ToUpper(sensor.Mac)
	if sensor.Group != nil {
		if *sensor.Group < 0 {
			api.sendError(w, APIErrorInvalidValue, "Invalid groupID "+strconv.Itoa(*sensor.Group), http.StatusInternalServerError)
			return
		}
		gr, _ := database.GetGroupConfig(api.db, *sensor.Group)
		if gr == nil {
			name := "Group " + strconv.Itoa(*sensor.Group)
			group := gm.GroupConfig{
				Group:        *sensor.Group,
				FriendlyName: &name,
			}
			if sensor.Mac != "" {
				sensors := []string{sensor.Mac}
				group.Sensors = sensors
			}
			eventGr := make(map[string]interface{})
			eventGr["group"] = group
			api.EventsToBackend <- eventGr
		}
	}
	event := make(map[string]interface{})
	event["sensor"] = sensor
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) setBlindConfig(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	cfg := dblind.BlindConf{}
	err = json.Unmarshal(body, &cfg)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	cfg.Mac = strings.ToUpper(cfg.Mac)

	if cfg.Group != nil {
		if *cfg.Group < 0 {
			api.sendError(w, APIErrorInvalidValue, "Invalid groupID "+strconv.Itoa(*cfg.Group), http.StatusInternalServerError)
			return
		}
		gr, _ := database.GetGroupConfig(api.db, *cfg.Group)
		if gr == nil {
			name := "Group " + strconv.Itoa(*cfg.Group)
			group := gm.GroupConfig{
				Group:        *cfg.Group,
				FriendlyName: &name,
			}
			if cfg.Mac != "" {
				blinds := []string{cfg.Mac}
				group.Blinds = blinds
			}
			eventGr := make(map[string]interface{})
			eventGr["group"] = group
			api.EventsToBackend <- eventGr
		}
	}
	event := make(map[string]interface{})
	event["blind"] = cfg
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) setHvacConfig(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	cfg := dhvac.HvacConf{}
	err = json.Unmarshal(body, &cfg)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	cfg.Mac = strings.ToUpper(cfg.Mac)

	if cfg.Group != nil {
		if *cfg.Group < 0 {
			api.sendError(w, APIErrorInvalidValue, "Invalid groupID "+strconv.Itoa(*cfg.Group), http.StatusInternalServerError)
			return
		}
		gr, _ := database.GetGroupConfig(api.db, *cfg.Group)
		if gr == nil {
			name := "Group " + strconv.Itoa(*cfg.Group)
			group := gm.GroupConfig{
				Group:        *cfg.Group,
				FriendlyName: &name,
			}
			if cfg.Mac != "" {
				hvacs := []string{cfg.Mac}
				group.Hvacs = hvacs
			}
			eventGr := make(map[string]interface{})
			eventGr["group"] = group
			api.EventsToBackend <- eventGr
		}
	}
	event := make(map[string]interface{})
	event["hvac"] = cfg
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) swagger() {
	router := mux.NewRouter()
	sh := http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("/media/userdata/www/swaggerui/")))
	router.PathPrefix("/swaggerui/").Handler(sh)

	// API v1.0
	apiV1 := "/v1.0"
	router.HandleFunc(apiV1+"/functions", api.getV1Functions).Methods("GET")

	//dump API
	router.HandleFunc(apiV1+"/dump", api.getDump).Methods("GET")

	//config API
	router.HandleFunc(apiV1+"/config/led", api.setLedConfig).Methods("POST")
	router.HandleFunc(apiV1+"/config/sensor", api.setSensorConfig).Methods("POST")
	router.HandleFunc(apiV1+"/config/blind", api.setBlindConfig).Methods("POST")
	router.HandleFunc(apiV1+"/config/hvac", api.setHvacConfig).Methods("POST")

	//command API
	router.HandleFunc(apiV1+"/command/led", api.sendLedCommand).Methods("POST")
	router.HandleFunc(apiV1+"/command/blind", api.sendBlindCommand).Methods("POST")
	router.HandleFunc(apiV1+"/command/hvac", api.sendHvacCommand).Methods("POST")
	router.HandleFunc(apiV1+"/command/group", api.sendGroupCommand).Methods("POST")

	//unversionned API
	router.HandleFunc("/versions", api.getAPIs).Methods("GET")
	router.HandleFunc("/functions", api.getFunctions).Methods("GET")

	if api.browsingFolder != "" {
		sh2 := http.StripPrefix("/", http.FileServer(http.Dir(api.browsingFolder)))
		router.PathPrefix("/").Handler(sh2)
	}

	log.Fatal(http.ListenAndServeTLS(api.apiIP+":"+api.apiPort, api.certificate, api.keyfile, router))
}
