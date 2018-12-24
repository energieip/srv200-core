package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/energieip/common-group-go/pkg/groupmodel"
	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
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
	eventsAPI       chan map[string]interface{}
	EventsToBackend chan map[string]interface{}
}

//Status
type Status struct {
	Leds    []driverled.Led       `json:"leds"`
	Sensors []driversensor.Sensor `json:"sensors"`
}

type EventLed struct {
	Led   driverled.Led `json:"led"`
	Label string        `json:"label"`
}

type EventSensor struct {
	Sensor driversensor.Sensor `json:"sensor"`
	Label  string              `json:"label"`
}

//EventStatus
type EventStatus struct {
	Leds    []EventLed    `json:"leds"`
	Sensors []EventSensor `json:"sensors"`
}

//DumpLed
type DumpLed struct {
	Ifc    core.IfcInfo       `json:"ifc"`
	Status driverled.Led      `json:"status"`
	Config driverled.LedSetup `json:"config"`
}

//DumpSensor
type DumpSensor struct {
	Ifc    core.IfcInfo             `json:"ifc"`
	Status driversensor.Sensor      `json:"status"`
	Config driversensor.SensorSetup `json:"config"`
}

//Dump
type Dump struct {
	Leds    []DumpLed    `json:"leds"`
	Sensors []DumpSensor `json:"sensors"`
}

//InitAPI start API connection
func InitAPI(db database.Database, eventsAPI chan map[string]interface{}) *API {
	api := API{
		db:              db,
		eventsAPI:       eventsAPI,
		EventsToBackend: make(chan map[string]interface{}),
		clients:         make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
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

func (api *API) sendCommand(w http.ResponseWriter, req *http.Request) {
	//TODO
	api.sendError(w, APIErrorBodyParsing, "Not yet Implemented")
}

func (api *API) getStatus(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	var leds []driverled.Led
	var sensors []driversensor.Sensor
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

		if filterByLabel {
			if _, ok := labels[ifc.Label]; !ok {
				continue
			}
		}
		leds = append(leds, dump)
	}

	cells := database.GetSensorsStatus(api.db)
	cellsConfig := database.GetSensorsConfig(api.db)
	for _, sensor := range cells {
		if filterByMac {
			if _, ok := macs[sensor.Mac]; !ok {
				continue
			}
		}
		cell := DumpSensor{}
		cell.Status = sensor
		config, ok := cellsConfig[sensor.Mac]
		if ok {
			cell.Config = config
		}
		ifc, ok := ifcs[sensor.Mac]
		if ok {
			cell.Ifc = ifc
		}

		if filterByLabel {
			if _, ok := labels[cell.Ifc.Label]; !ok {
				continue
			}
		}
		sensors = append(sensors, cell)
	}

	dump := Dump{
		Leds:    leds,
		Sensors: sensors,
	}

	inrec, _ := json.MarshalIndent(dump, "", "  ")
	w.Write(inrec)
}

type Conf struct {
	Leds    []driverled.LedConf       `json:"leds"`
	Sensors []driversensor.SensorConf `json:"sensors"`
	Groups  []groupmodel.GroupConfig  `json:"groups"`
	Switchs []core.SwitchConfig       `json:"switchs"`
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
	w.Write([]byte(""))
}

func (api *API) webEvents(w http.ResponseWriter, r *http.Request) {
	ws, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		rlog.Error("Error when switching in websocket " + err.Error())
		return
	}
	api.clients[ws] = true

	go func() {
		for {
			select {
			case events := <-api.eventsAPI:
				for eventType, event := range events {
					var leds []EventLed
					var sensors []EventSensor
					res := strings.Split(eventType, ".")
					driver := res[0]
					action := res[1]
					// Convert Type
					switch driver {
					case "sensor":
						sensor, err := driversensor.ToSensor(event)

						if err == nil && sensor != nil {
							label := ""
							project := database.GetProjectByMac(api.db, sensor.Mac)
							if project != nil {
								label = project.Label
							}
							evt := EventSensor{
								Sensor: *sensor,
								Label:  label,
							}
							sensors = append(sensors, evt)
						}
					case "led":
						led, err := driverled.ToLed(event)
						if err == nil && led != nil {
							label := ""
							project := database.GetProjectByMac(api.db, led.Mac)
							if project != nil {
								label = project.Label
							}
							evt := EventLed{
								Led:   *led,
								Label: label,
							}
							leds = append(leds, evt)
						}
					}
					evt := make(map[string]EventStatus)
					evt[action] = EventStatus{
						Leds:    leds,
						Sensors: sensors,
					}

					for client := range api.clients {
						if err := client.WriteJSON(evt); err != nil {
							rlog.Error("Error writing in websocket" + err.Error())
							client.Close()
							delete(api.clients, client)
						}
					}
				}
			}
		}
	}()
}

func (api *API) swagger() {
	router := mux.NewRouter()
	sh := http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("/var/www/swaggerui/")))
	router.PathPrefix("/swaggerui/").Handler(sh)

	//setup API
	router.HandleFunc("/setup/sensor/{mac}", api.getSensorSetup).Methods("GET")
	router.HandleFunc("/setup/sensor/{mac}", api.removeSensorSetup).Methods("DELETE")
	router.HandleFunc("/setup/sensor", api.setSensorSetup).Methods("POST")
	router.HandleFunc("/setup/led/{mac}", api.getLedSetup).Methods("GET")
	router.HandleFunc("/setup/led/{mac}", api.removeLedSetup).Methods("DELETE")
	router.HandleFunc("/setup/led", api.setLedSetup).Methods("POST")
	router.HandleFunc("/setup/group/{groupID}", api.getGroupSetup).Methods("GET")
	router.HandleFunc("/setup/group/{groupID}", api.removeGroupSetup).Methods("DELETE")
	router.HandleFunc("/setup/group", api.setGroupSetup).Methods("POST")
	router.HandleFunc("/setup/switch/{mac}", api.getSwitchSetup).Methods("GET")
	router.HandleFunc("/setup/switch/{mac}", api.removeSwitchSetup).Methods("DELETE")
	router.HandleFunc("/setup/switch", api.setSwitchSetup).Methods("POST")

	//config API
	router.HandleFunc("/config/led", api.setLedConfig).Methods("POST")
	router.HandleFunc("/config/sensor", api.setSensorConfig).Methods("POST")
	router.HandleFunc("/config/group", api.setGroupConfig).Methods("POST")
	router.HandleFunc("/config/switch", api.setSwitchConfig).Methods("POST")
	router.HandleFunc("/configs", api.setConfig).Methods("POST")

	//status API
	router.HandleFunc("/status/sensor/{mac}", api.getSensorStatus).Methods("GET")
	router.HandleFunc("/status/led/{mac}", api.getLedStatus).Methods("GET")
	router.HandleFunc("/status", api.getStatus).Methods("GET")

	//events API
	router.HandleFunc("/events", api.webEvents)

	//command API
	router.HandleFunc("/command/led", api.sendLedCommand).Methods("POST")
	router.HandleFunc("/command/group", api.sendGroupCommand).Methods("POST")
	router.HandleFunc("/commands", api.sendCommand).Methods("POST")

	//project API
	router.HandleFunc("/project/ifcInfo/{label}", api.getIfcInfo).Methods("GET")
	router.HandleFunc("/project/ifcInfo/{label}", api.removeIfcInfo).Methods("DELETE")
	router.HandleFunc("/project/ifcInfo", api.setIfcInfo).Methods("POST")
	router.HandleFunc("/project/model/{modelName}", api.getModelInfo).Methods("GET")
	router.HandleFunc("/project/model/{modelName}", api.removeModelInfo).Methods("DELETE")
	router.HandleFunc("/project/model", api.setModelInfo).Methods("POST")
	router.HandleFunc("/project", api.getIfc).Methods("GET")

	//dump API
	router.HandleFunc("/dump", api.getDump).Methods("GET")

	log.Fatal(http.ListenAndServe(":8888", router))
}
