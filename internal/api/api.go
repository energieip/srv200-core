package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/energieip/srv200-coreservice-go/internal/history"

	"github.com/energieip/common-components-go/pkg/dblind"
	"github.com/energieip/common-components-go/pkg/dhvac"

	gm "github.com/energieip/common-components-go/pkg/dgroup"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/romana/rlog"
)

const (
	APIErrorDeviceNotFound = 1
	APIErrorBodyParsing    = 2
	APIErrorDatabase       = 3
	APIErrorInvalidValue   = 4

	FilterTypeAll    = "all"
	FilterTypeSensor = "sensor"
	FilterTypeLed    = "led"
	FilterTypeBlind  = "blind"
	FilterTypeHvac   = "hvac"
)

//APIError Message error code
type APIError struct {
	Code    int    `json:"code"` //errorCode
	Message string `json:"message"`
}

type API struct {
	clients         map[*websocket.Conn]bool
	clientsConso    map[*websocket.Conn]bool
	upgrader        websocket.Upgrader
	db              database.Database
	historydb       history.HistoryDb
	eventsAPI       chan map[string]interface{}
	eventsConso     chan core.EventConsumption
	EventsToBackend chan map[string]interface{}
	apiMutex        sync.Mutex
	installMode     *bool
	certificate     string
	keyfile         string
}

//Status
type Status struct {
	Leds    []dl.Led       `json:"leds"`
	Sensors []ds.Sensor    `json:"sensors"`
	Blind   []dblind.Blind `json:"blinds"`
	Hvac    []dhvac.Hvac   `json:"hvacs"`
}

//DumpBlind
type DumpBlind struct {
	Ifc    core.IfcInfo      `json:"ifc"`
	Status dblind.Blind      `json:"status"`
	Config dblind.BlindSetup `json:"config"`
}

//DumpHvac
type DumpHvac struct {
	Ifc    core.IfcInfo    `json:"ifc"`
	Status dhvac.Hvac      `json:"status"`
	Config dhvac.HvacSetup `json:"config"`
}

//DumpLed
type DumpLed struct {
	Ifc    core.IfcInfo `json:"ifc"`
	Status dl.Led       `json:"status"`
	Config dl.LedSetup  `json:"config"`
}

//DumpSensor
type DumpSensor struct {
	Ifc    core.IfcInfo   `json:"ifc"`
	Status ds.Sensor      `json:"status"`
	Config ds.SensorSetup `json:"config"`
}

//DumpSwitch
type DumpSwitch struct {
	Ifc    core.IfcInfo      `json:"ifc"`
	Status core.SwitchDump   `json:"status"`
	Config core.SwitchConfig `json:"config"`
}

//Dump
type Dump struct {
	Leds    []DumpLed    `json:"leds"`
	Sensors []DumpSensor `json:"sensors"`
	Blinds  []DumpBlind  `json:"blinds"`
	Hvacs   []DumpHvac   `json:"hvacs"`
	Switchs []DumpSwitch `json:"switchs"`
}

//InitAPI start API connection
func InitAPI(db database.Database, historydb history.HistoryDb, eventsAPI chan map[string]interface{},
	eventsConso chan core.EventConsumption, installMode *bool, conf pkg.ServiceConfig) *API {
	api := API{
		db:              db,
		historydb:       historydb,
		eventsAPI:       eventsAPI,
		eventsConso:     eventsConso,
		EventsToBackend: make(chan map[string]interface{}),
		clients:         make(map[*websocket.Conn]bool),
		clientsConso:    make(map[*websocket.Conn]bool),
		installMode:     installMode,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		certificate: conf.Certificate,
		keyfile:     conf.Key,
	}
	go api.swagger()
	return &api
}

func (api *API) setDefaultHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
}

func (api *API) sendError(w http.ResponseWriter, errorCode int, message string) {
	errCode := APIError{
		Code:    APIErrorDeviceNotFound,
		Message: message,
	}

	inrec, _ := json.MarshalIndent(errCode, "", "  ")
	rlog.Error(errCode.Message)
	http.Error(w, string(inrec),
		http.StatusInternalServerError)
}

type InstallModeStruct struct {
	InstallMode bool `json:"installMode"`
}

func (api *API) setInstallMode(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	inputMode := InstallModeStruct{}
	err = json.Unmarshal(body, &inputMode)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}
	*api.installMode = inputMode.InstallMode

	status := InstallModeStruct{
		InstallMode: *api.installMode,
	}

	inrec, _ := json.MarshalIndent(status, "", "  ")
	w.Write(inrec)
}

func (api *API) getInstallMode(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	status := InstallModeStruct{
		InstallMode: *api.installMode,
	}

	inrec, _ := json.MarshalIndent(status, "", "  ")
	w.Write(inrec)
}

func (api *API) getStatus(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	var leds []dl.Led
	var sensors []ds.Sensor
	var blinds []dblind.Blind
	var hvacs []dhvac.Hvac
	var grID *int
	var isConfig *bool
	driverType := req.FormValue("type")
	if driverType == "" {
		driverType = FilterTypeAll
	}

	groupID := req.FormValue("groupID")
	if groupID != "" {
		i, err := strconv.Atoi(groupID)
		if err == nil {
			grID = &i
		}
	}

	isConfigured := req.FormValue("isConfigured")
	if isConfigured != "" {
		b, err := strconv.ParseBool(isConfigured)
		if err == nil {
			isConfig = &b
		}
	}

	if driverType == FilterTypeAll || driverType == FilterTypeLed {
		lights := database.GetLedsStatus(api.db)
		for _, led := range lights {
			if grID == nil || *grID == led.Group {
				if isConfig == nil || *isConfig == led.IsConfigured {
					leds = append(leds, led)
				}
			}
		}
	}

	if driverType == FilterTypeAll || driverType == FilterTypeSensor {
		cells := database.GetSensorsStatus(api.db)
		for _, sensor := range cells {
			if grID == nil || *grID == sensor.Group {
				if isConfig == nil || *isConfig == sensor.IsConfigured {
					sensors = append(sensors, sensor)
				}
			}
		}
	}

	if driverType == FilterTypeAll || driverType == FilterTypeBlind {
		drivers := database.GetBlindsStatus(api.db)
		for _, driver := range drivers {
			if grID == nil || *grID == driver.Group {
				if isConfig == nil || *isConfig == driver.IsConfigured {
					blinds = append(blinds, driver)
				}
			}
		}
	}

	if driverType == FilterTypeAll || driverType == FilterTypeHvac {
		drivers := database.GetHvacsStatus(api.db)
		for _, driver := range drivers {
			if grID == nil || *grID == driver.Group {
				if isConfig == nil || *isConfig == driver.IsConfigured {
					hvacs = append(hvacs, driver)
				}
			}
		}
	}

	status := Status{
		Leds:    leds,
		Sensors: sensors,
		Blind:   blinds,
		Hvac:    hvacs,
	}

	inrec, _ := json.MarshalIndent(status, "", "  ")
	w.Write(inrec)
}

func (api *API) getDump(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	var leds []DumpLed
	var sensors []DumpSensor
	var switchs []DumpSwitch
	var blinds []DumpBlind
	var hvacs []DumpHvac
	macs := make(map[string]bool)
	labels := make(map[string]bool)
	filterByMac := false
	MacsParam := req.FormValue("macs")
	if MacsParam != "" {
		tempMac := strings.Split(MacsParam, ",")

		for _, v := range tempMac {
			macs[v] = true
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

	lights := database.GetLedsStatus(api.db)
	lightsConfig := database.GetLedsConfig(api.db)
	cells := database.GetSensorsStatus(api.db)
	cellsConfig := database.GetSensorsConfig(api.db)
	blds := database.GetBlindsStatus(api.db)
	bldsConfig := database.GetBlindsConfig(api.db)
	hvcs := database.GetHvacsStatus(api.db)
	hvcsConfig := database.GetHvacsConfig(api.db)
	switchElts := database.GetSwitchsDump(api.db)
	switchEltsConfig := database.GetSwitchsConfig(api.db)

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

		switch ifc.DeviceType {
		case "led":
			dump := DumpLed{}
			led, ok := lights[ifc.Mac]
			if ok {
				dump.Status = led
			}
			config, ok := lightsConfig[ifc.Mac]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc

			leds = append(leds, dump)
		case "sensor":
			dump := DumpSensor{}
			sensor, ok := cells[ifc.Mac]
			if ok {
				dump.Status = sensor
			}
			config, ok := cellsConfig[ifc.Mac]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			sensors = append(sensors, dump)
		case "blind":
			dump := DumpBlind{}
			bld, ok := blds[ifc.Mac]
			if ok {
				dump.Status = bld
			}
			config, ok := bldsConfig[ifc.Mac]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			blinds = append(blinds, dump)
		case "hvac":
			dump := DumpHvac{}
			hvc, ok := hvcs[ifc.Mac]
			if ok {
				dump.Status = hvc
			}
			config, ok := hvcsConfig[ifc.Mac]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			hvacs = append(hvacs, dump)
		case "switch":
			dump := DumpSwitch{}
			switchElt, ok := switchElts[ifc.Mac]
			if ok {
				dump.Status = switchElt
			}
			config, ok := switchEltsConfig[ifc.Mac]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			switchs = append(switchs, dump)
		}
	}

	dump := Dump{
		Leds:    leds,
		Sensors: sensors,
		Blinds:  blinds,
		Hvacs:   hvacs,
		Switchs: switchs,
	}

	inrec, _ := json.MarshalIndent(dump, "", "  ")
	w.Write(inrec)
}

type LedHist struct {
	Energy float64 `json:"energy"`
	Power  int     `json:"power"`
	Date   string  `json:"date"`
}

type GlobalHistory struct {
	Leds []LedHist `json:"leds"`
}

func (api *API) getHistory(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)

	var leds []LedHist

	ledHistories := make(map[string]LedHist)
	ledDrivers := history.GetLedsHistory(api.historydb)
	for _, l := range ledDrivers {
		val, ok := ledHistories[l.Date]
		if !ok {
			ledHistories[l.Date] = LedHist{
				Energy: l.Energy,
				Power:  l.Power,
				Date:   l.Date,
			}
		} else {
			val.Energy += l.Energy
			val.Power += l.Power
			ledHistories[l.Date] = val
		}
	}
	for _, hist := range ledHistories {
		leds = append(leds, hist)
	}

	dump := GlobalHistory{
		Leds: leds,
	}

	inrec, _ := json.MarshalIndent(dump, "", "  ")
	w.Write(inrec)
}

type Conf struct {
	Leds    []dl.LedConf        `json:"leds"`
	Sensors []ds.SensorConf     `json:"sensors"`
	Blinds  []dblind.BlindConf  `json:"blinds"`
	Hvacs   []dhvac.HvacConf    `json:"hvacs"`
	Groups  []gm.GroupConfig    `json:"groups"`
	Switchs []core.SwitchConfig `json:"switchs"`
}

func (api *API) setConfig(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	config := Conf{}
	err = json.Unmarshal([]byte(body), &config)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}
	event := make(map[string]interface{})
	for _, led := range config.Leds {
		event["led"] = led
	}
	for _, sensor := range config.Sensors {
		event["sensor"] = sensor
	}
	for _, group := range config.Groups {
		event["group"] = group
	}
	for _, sw := range config.Switchs {
		event["switch"] = sw
	}
	for _, sw := range config.Blinds {
		event["blind"] = sw
	}
	for _, sw := range config.Hvacs {
		event["hvac"] = sw
	}
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) webEvents(w http.ResponseWriter, r *http.Request) {
	ws, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		rlog.Error("Error when switching in websocket " + err.Error())
		return
	}
	api.clients[ws] = true

}

func (api *API) consumptionEvents(w http.ResponseWriter, r *http.Request) {
	ws, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		rlog.Error("Error when switching in consumption websocket " + err.Error())
		return
	}
	api.clientsConso[ws] = true
}

func (api *API) websocketEvents() {
	for {
		select {
		case event := <-api.eventsAPI:
			api.apiMutex.Lock()
			for client := range api.clients {
				if err := client.WriteJSON(event); err != nil {
					rlog.Error("Error writing in websocket" + err.Error())
					client.Close()
					delete(api.clients, client)
				}
			}
			api.apiMutex.Unlock()
		}
	}
}

func (api *API) websocketConsumptions() {
	for {
		select {
		case event := <-api.eventsConso:
			api.apiMutex.Lock()
			for client := range api.clientsConso {
				if err := client.WriteJSON(event); err != nil {
					rlog.Error("Error writing in websocket" + err.Error())
					client.Close()
					delete(api.clientsConso, client)
				}
			}
			api.apiMutex.Unlock()
		}
	}
}

type APIInfo struct {
	Versions []string `json:"versions"`
}

type APIFunctions struct {
	Functions []string `json:"functions"`
}

func (api *API) getAPIs(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	versions := []string{"v1.0"}
	apiInfo := APIInfo{
		Versions: versions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *API) getV1Functions(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	apiV1 := "/v1.0"
	functions := []string{apiV1 + "/setup/sensor", apiV1 + "/setup/led",
		apiV1 + "/setup/group", apiV1 + "/setup/switch", apiV1 + "/setup/installMode",
		apiV1 + "/setup/service", apiV1 + "/setup/blind", apiV1 + "/setup/hvac",
		apiV1 + "/config/led", apiV1 + "/config/sensor", apiV1 + "/config/blind", apiV1 + "/config/hvac",
		apiV1 + "/config/group", apiV1 + "/config/switch", apiV1 + "/configs",
		apiV1 + "/status", apiV1 + "/events", apiV1 + "/events/consumption", apiV1 + "/history",
		apiV1 + "/command/led", apiV1 + "/command/blind", apiV1 + "/command/hvac", apiV1 + "/command/group", apiV1 + "/project/ifcInfo",
		apiV1 + "/project/model", apiV1 + "/project/bim", apiV1 + "/project", apiV1 + "/dump",
		apiV1 + "/status/sensor", apiV1 + "/status/group", apiV1 + "/status/led", apiV1 + "/status/blind", apiV1 + "/status/hvac",
		apiV1 + "/status/groups" + apiV1 + "/maintenance/driver", apiV1 + "/commissioning/install",
	}
	apiInfo := APIFunctions{
		Functions: functions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *API) getFunctions(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	functions := []string{"/versions"}
	apiInfo := APIFunctions{
		Functions: functions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *API) swagger() {
	go api.websocketConsumptions()
	go api.websocketEvents()
	router := mux.NewRouter()
	sh := http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("/media/userdata/www/swaggerui/")))
	router.PathPrefix("/swaggerui/").Handler(sh)

	// API v1.0
	apiV1 := "/v1.0"
	router.HandleFunc(apiV1+"/functions", api.getV1Functions).Methods("GET")

	//setup API
	router.HandleFunc(apiV1+"/setup/sensor/{mac}", api.getSensorSetup).Methods("GET")
	router.HandleFunc(apiV1+"/setup/sensor/{mac}", api.removeSensorSetup).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/sensor", api.setSensorSetup).Methods("POST")
	router.HandleFunc(apiV1+"/setup/led/{mac}", api.getLedSetup).Methods("GET")
	router.HandleFunc(apiV1+"/setup/led/{mac}", api.removeLedSetup).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/led", api.setLedSetup).Methods("POST")
	router.HandleFunc(apiV1+"/setup/blind/{mac}", api.getBlindSetup).Methods("GET")
	router.HandleFunc(apiV1+"/setup/blind/{mac}", api.removeBlindSetup).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/blind", api.setBlindSetup).Methods("POST")
	router.HandleFunc(apiV1+"/setup/hvac/{mac}", api.getHvacSetup).Methods("GET")
	router.HandleFunc(apiV1+"/setup/hvac/{mac}", api.removeHvacSetup).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/hvac", api.setHvacSetup).Methods("POST")
	router.HandleFunc(apiV1+"/setup/group/{groupID}", api.getGroupSetup).Methods("GET")
	router.HandleFunc(apiV1+"/setup/group/{groupID}", api.removeGroupSetup).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/group", api.setGroupSetup).Methods("POST")
	router.HandleFunc(apiV1+"/setup/switch/{mac}", api.getSwitchSetup).Methods("GET")
	router.HandleFunc(apiV1+"/setup/switch/{mac}", api.removeSwitchSetup).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/switch", api.setSwitchSetup).Methods("POST")
	router.HandleFunc(apiV1+"/setup/installMode", api.getInstallMode).Methods("GET")
	router.HandleFunc(apiV1+"/setup/installMode", api.setInstallMode).Methods("POST")
	router.HandleFunc(apiV1+"/setup/service/{name}", api.getServiceSetup).Methods("GET")
	router.HandleFunc(apiV1+"/setup/service/{name}", api.removeServiceSetup).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/service", api.setServiceSetup).Methods("POST")

	//config API
	router.HandleFunc(apiV1+"/config/led", api.setLedConfig).Methods("POST")
	router.HandleFunc(apiV1+"/config/sensor", api.setSensorConfig).Methods("POST")
	router.HandleFunc(apiV1+"/config/blind", api.setBlindConfig).Methods("POST")
	router.HandleFunc(apiV1+"/config/hvac", api.setHvacConfig).Methods("POST")
	router.HandleFunc(apiV1+"/config/group", api.setGroupConfig).Methods("POST")
	router.HandleFunc(apiV1+"/config/switch", api.setSwitchConfig).Methods("POST")
	router.HandleFunc(apiV1+"/configs", api.setConfig).Methods("POST")

	//status API
	router.HandleFunc(apiV1+"/status/sensor/{mac}", api.getSensorStatus).Methods("GET")
	router.HandleFunc(apiV1+"/status/blind/{mac}", api.getBlindStatus).Methods("GET")
	router.HandleFunc(apiV1+"/status/hvac/{mac}", api.getHvacStatus).Methods("GET")
	router.HandleFunc(apiV1+"/status/led/{mac}", api.getLedStatus).Methods("GET")
	router.HandleFunc(apiV1+"/status/group/{groupID}", api.getGroupStatus).Methods("GET")
	router.HandleFunc(apiV1+"/status/groups", api.getGroupsStatus).Methods("GET")
	router.HandleFunc(apiV1+"/status", api.getStatus).Methods("GET")

	//events API
	router.HandleFunc(apiV1+"/events", api.webEvents)
	router.HandleFunc(apiV1+"/events/consumption", api.consumptionEvents)

	//command API
	router.HandleFunc(apiV1+"/command/led", api.sendLedCommand).Methods("POST")
	router.HandleFunc(apiV1+"/command/blind", api.sendBlindCommand).Methods("POST")
	router.HandleFunc(apiV1+"/command/hvac", api.sendHvacCommand).Methods("POST")
	router.HandleFunc(apiV1+"/command/group", api.sendGroupCommand).Methods("POST")

	//project API
	router.HandleFunc(apiV1+"/project/ifcInfo/{label}", api.getIfcInfo).Methods("GET")
	router.HandleFunc(apiV1+"/project/ifcInfo/{label}", api.removeIfcInfo).Methods("DELETE")
	router.HandleFunc(apiV1+"/project/ifcInfo", api.setIfcInfo).Methods("POST")
	router.HandleFunc(apiV1+"/project/model/{modelName}", api.getModelInfo).Methods("GET")
	router.HandleFunc(apiV1+"/project/model/{modelName}", api.removeModelInfo).Methods("DELETE")
	router.HandleFunc(apiV1+"/project/model", api.setModelInfo).Methods("POST")
	router.HandleFunc(apiV1+"/project/bim/{label}", api.getBim).Methods("GET")
	router.HandleFunc(apiV1+"/project/bim/{label}", api.removeBim).Methods("DELETE")
	router.HandleFunc(apiV1+"/project/bim", api.setBim).Methods("POST")
	router.HandleFunc(apiV1+"/project", api.getIfc).Methods("GET")

	//Maintenance API
	router.HandleFunc(apiV1+"/maintenance/driver", api.replaceDriver).Methods("POST")

	//Install API
	router.HandleFunc(apiV1+"/commissioning/install", api.installDriver).Methods("POST")

	//dump API
	router.HandleFunc(apiV1+"/dump", api.getDump).Methods("GET")

	//History API
	router.HandleFunc(apiV1+"/history", api.getHistory).Methods("GET")

	//unversionned API
	router.HandleFunc("/versions", api.getAPIs).Methods("GET")
	router.HandleFunc("/functions", api.getFunctions).Methods("GET")

	log.Fatal(http.ListenAndServeTLS(":8888", api.certificate, api.keyfile, router))
}
