package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/common-sensor-go/pkg/driversensor"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/romana/rlog"
)

const (
	APIErrorDeviceNotFound = 1
	APIErrorBodyParsing    = 2
)

//APIError Message error code
type APIError struct {
	Code    int    `json:"code"` //errorCode
	Message string `json:"message"`
}

type API struct {
	clients   map[*websocket.Conn]bool
	upgrader  websocket.Upgrader
	db        database.Database
	eventsAPI chan map[string]interface{}
}

type ModelInfo struct {
	Label     string `json:"label"` //cable label
	ModelName string `json:"modelName"`
	Mac       string `json:"mac"` //device Mac address
	Vendor    string `json:"vendor"`
	URL       string `json:"url"`
}

//Status
type Status struct {
	Leds    []driverled.Led       `json:"leds"`
	Sensors []driversensor.Sensor `json:"sensors"`
}

//InitAPI start API connection
func InitAPI(db database.Database, eventsAPI chan map[string]interface{}) *API {
	api := API{
		db:        db,
		eventsAPI: eventsAPI,
		clients:   make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	go api.swagger()
	return &api
}

func (api *API) seDefaultHeader(w http.ResponseWriter) {
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

func (api *API) getLeds(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	leds := database.GetLedsStatus(api.db)
	inrec, err := json.MarshalIndent(leds, "", "  ")
	if err != nil {
		rlog.Error("Could not parse input format " + err.Error())
		http.Error(w, "Error Parsing json",
			http.StatusInternalServerError)
		return
	}
	w.Write(inrec)
}

func (api *API) getModelInfo(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(req)
	label := params["label"]
	project := database.GetProject(api.db, label)
	if project == nil {
		rlog.Error("Could not parse input format")
		http.Error(w, "Error Parsing json",
			http.StatusInternalServerError)
		return
	}
	model := database.GetModel(api.db, project.ModelName)
	info := ModelInfo{
		Label:     label,
		ModelName: model.Name,
		Mac:       project.Mac,
		Vendor:    model.Vendor,
		URL:       model.URL,
	}
	inrec, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		rlog.Error("Could not parse input format " + err.Error())
		http.Error(w, "Error Parsing json",
			http.StatusInternalServerError)
		return
	}
	w.Write(inrec)
}

func (api *API) getSensors(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	sensors := database.GetSensorsStatus(api.db)
	inrec, err := json.MarshalIndent(sensors, "", "  ")
	if err != nil {
		rlog.Error("Could not parse input format " + err.Error())
		http.Error(w, "Error Parsing json",
			http.StatusInternalServerError)
		return
	}
	w.Write(inrec)
}

func (api *API) setSensor(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	if req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}
		sensor := driversensor.Sensor{}
		err = json.Unmarshal([]byte(body), &sensor)
		if err != nil {
			rlog.Error("Could not parse input format " + err.Error())
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}
		config := driversensor.SensorSetup{
			Mac:          sensor.Mac,
			Group:        &sensor.Group,
			FriendlyName: &sensor.FriendlyName,
		}
		rlog.Info("Try to save ", config)
		database.SaveSensorConfig(api.db, config)
		json.NewEncoder(w).Encode(nil)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
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
					var leds []driverled.Led
					var sensors []driversensor.Sensor
					// Convert Type
					sensor, err := driversensor.ToSensor(event)
					if err == nil && sensor != nil {
						sensors = append(sensors, *sensor)
					} else {
						led, err := driverled.ToLed(event)
						if err == nil && led != nil {
							leds = append(leds, *led)
						}
					}
					evt := make(map[string]Status)
					evt[eventType] = Status{
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
	router.HandleFunc("/setup/sensor", api.setSensorSetup).Methods("POST")
	router.HandleFunc("/setup/led/{mac}", api.getLedSetup).Methods("GET")
	router.HandleFunc("/setup/led", api.setLedSetup).Methods("POST")
	router.HandleFunc("/setup/group/{groupID}", api.getGroupSetup).Methods("GET")
	router.HandleFunc("/setup/group", api.setGroupSetup).Methods("POST")
	router.HandleFunc("/setup/switch/{mac}", api.getSwitchSetup).Methods("GET")
	router.HandleFunc("/setup/switch", api.setSwitchSetup).Methods("POST")

	//status API
	router.HandleFunc("/status/sensor/{mac}", api.getSensorStatus).Methods("GET")
	router.HandleFunc("/status/led/{mac}", api.getLedStatus).Methods("GET")
	router.HandleFunc("/events", api.webEvents)

	router.HandleFunc("/leds", api.getLeds).Methods("GET")
	router.HandleFunc("/sensors", api.getSensors).Methods("GET")
	router.HandleFunc("/sensor/{mac}", api.setSensor).Methods("POST")
	router.HandleFunc("/modelInfo/{label}", api.getModelInfo).Methods("GET")
	log.Fatal(http.ListenAndServe(":8888", router))
}
