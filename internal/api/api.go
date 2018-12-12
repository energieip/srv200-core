package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/energieip/common-sensor-go/pkg/driversensor"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/romana/rlog"
)

type API struct {
	clients   map[*websocket.Conn]bool
	upgrader  websocket.Upgrader
	db        database.Database
	eventsAPI chan map[string]interface{}
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

type ModelInfo struct {
	Label     string `json:"label"` //cable label
	ModelName string `json:"modelName"`
	Mac       string `json:"mac"` //device Mac address
	Vendor    string `json:"vendor"`
	URL       string `json:"url"`
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

func (api *API) getLed(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(req)
	led := database.GetLedStatus(api.db, params["mac"])
	inrec, err := json.MarshalIndent(led, "", "  ")
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

func (api *API) getSensor(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(req)
	sensor := database.GetSensorStatus(api.db, params["mac"])
	inrec, err := json.MarshalIndent(sensor, "", "  ")
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
		rlog.Error("Error when switchin in websocket " + err.Error())
		return
	}
	api.clients[ws] = true

	go func() {
		for {
			select {
			case events := <-api.eventsAPI:
				for eventType, event := range events {
					for client := range api.clients {
						evt := make(map[string]interface{})
						dev := make(map[string]interface{})
						dev["Sensor"] = event
						evt[eventType] = dev
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
	router.HandleFunc("/leds", api.getLeds).Methods("GET")
	router.HandleFunc("/led/{mac}", api.getLed).Methods("GET")
	router.HandleFunc("/sensors", api.getSensors).Methods("GET")
	router.HandleFunc("/sensor/{mac}", api.getSensor).Methods("GET")
	router.HandleFunc("/sensor/{mac}", api.setSensor).Methods("POST")
	router.HandleFunc("/modelInfo/{label}", api.getModelInfo).Methods("GET")
	router.HandleFunc("/events", api.webEvents)
	log.Fatal(http.ListenAndServe(":8888", router))
}
