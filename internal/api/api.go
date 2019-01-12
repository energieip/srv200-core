package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

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
)

//APIError Message error code
type APIError struct {
	Code    int    `json:"code"` //errorCode
	Message string `json:"message"`
}

type API struct {
	clients         map[*websocket.Conn]bool
	upgrader        websocket.Upgrader
	db              database.Database
	eventsAPI       chan map[string]core.EventStatus
	EventsToBackend chan map[string]interface{}
	apiMutex        sync.Mutex
	installMode     *bool
	certificate     string
	keyfile         string
}

//Status
type Status struct {
	Leds    []dl.Led    `json:"leds"`
	Sensors []ds.Sensor `json:"sensors"`
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
	Switchs []DumpSwitch `json:"switchs"`
}

//InitAPI start API connection
func InitAPI(db database.Database, eventsAPI chan map[string]core.EventStatus, installMode *bool, conf pkg.ServiceConfig) *API {
	api := API{
		db:              db,
		eventsAPI:       eventsAPI,
		EventsToBackend: make(chan map[string]interface{}),
		clients:         make(map[*websocket.Conn]bool),
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

	status := Status{
		Leds:    leds,
		Sensors: sensors,
	}

	inrec, _ := json.MarshalIndent(status, "", "  ")
	w.Write(inrec)
}

func (api *API) getDump(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	var leds []DumpLed
	var sensors []DumpSensor
	var switchs []DumpSwitch
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
		Switchs: switchs,
	}

	inrec, _ := json.MarshalIndent(dump, "", "  ")
	w.Write(inrec)
}

type Conf struct {
	Leds    []dl.LedConf        `json:"leds"`
	Sensors []ds.SensorConf     `json:"sensors"`
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
		apiV1 + "/config/led", apiV1 + "/config/sensor", apiV1 + "/config/group",
		apiV1 + "/config/switch", apiV1 + "/configs", apiV1 + "/status", apiV1 + "/events",
		apiV1 + "/command/led", apiV1 + "/command/group", apiV1 + "/project/ifcInfo",
		apiV1 + "/project/model", apiV1 + "/project/bim", apiV1 + "/project", apiV1 + "/dump",
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
	router := mux.NewRouter()
	sh := http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("/var/www/swaggerui/")))
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
	router.HandleFunc(apiV1+"/setup/group/{groupID}", api.getGroupSetup).Methods("GET")
	router.HandleFunc(apiV1+"/setup/group/{groupID}", api.removeGroupSetup).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/group", api.setGroupSetup).Methods("POST")
	router.HandleFunc(apiV1+"/setup/switch/{mac}", api.getSwitchSetup).Methods("GET")
	router.HandleFunc(apiV1+"/setup/switch/{mac}", api.removeSwitchSetup).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/switch", api.setSwitchSetup).Methods("POST")
	router.HandleFunc(apiV1+"/setup/installMode", api.getInstallMode).Methods("GET")
	router.HandleFunc(apiV1+"/setup/installMode", api.setInstallMode).Methods("POST")

	//config API
	router.HandleFunc(apiV1+"/config/led", api.setLedConfig).Methods("POST")
	router.HandleFunc(apiV1+"/config/sensor", api.setSensorConfig).Methods("POST")
	router.HandleFunc(apiV1+"/config/group", api.setGroupConfig).Methods("POST")
	router.HandleFunc(apiV1+"/config/switch", api.setSwitchConfig).Methods("POST")
	router.HandleFunc(apiV1+"/configs", api.setConfig).Methods("POST")

	//status API
	router.HandleFunc(apiV1+"/status/sensor/{mac}", api.getSensorStatus).Methods("GET")
	router.HandleFunc(apiV1+"/status/led/{mac}", api.getLedStatus).Methods("GET")
	router.HandleFunc(apiV1+"/status/group/{groupID}", api.getGroupStatus).Methods("GET")
	router.HandleFunc(apiV1+"/status", api.getStatus).Methods("GET")

	//events API
	router.HandleFunc(apiV1+"/events", api.webEvents)

	//command API
	router.HandleFunc(apiV1+"/command/led", api.sendLedCommand).Methods("POST")
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

	//dump API
	router.HandleFunc(apiV1+"/dump", api.getDump).Methods("GET")

	//unversionned API
	router.HandleFunc("/versions", api.getAPIs).Methods("GET")
	router.HandleFunc("/functions", api.getFunctions).Methods("GET")

	log.Fatal(http.ListenAndServeTLS(":8888", api.certificate, api.keyfile, router))
}
