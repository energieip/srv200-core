package service

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/energieip/common-sensor-go/pkg/driversensor"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (s *CoreService) getLeds(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	leds := database.GetLedsStatus(s.db)
	inrec, err := json.MarshalIndent(leds, "", "  ")
	if err != nil {
		rlog.Error("Could not parse input format " + err.Error())
		http.Error(w, "Error Parsing json",
			http.StatusInternalServerError)
		return
	}
	w.Write(inrec)
}

func (s *CoreService) getLed(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(req)
	led := database.GetLedStatus(s.db, params["mac"])
	inrec, err := json.MarshalIndent(led, "", "  ")
	if err != nil {
		rlog.Error("Could not parse input format " + err.Error())
		http.Error(w, "Error Parsing json",
			http.StatusInternalServerError)
		return
	}
	w.Write(inrec)
}

func (s *CoreService) getSensors(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	sensors := database.GetSensorsStatus(s.db)
	inrec, err := json.MarshalIndent(sensors, "", "  ")
	if err != nil {
		rlog.Error("Could not parse input format " + err.Error())
		http.Error(w, "Error Parsing json",
			http.StatusInternalServerError)
		return
	}
	w.Write(inrec)
}

func (s *CoreService) getSensor(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(req)
	sensor := database.GetSensorStatus(s.db, params["mac"])
	inrec, err := json.MarshalIndent(sensor, "", "  ")
	if err != nil {
		rlog.Error("Could not parse input format " + err.Error())
		http.Error(w, "Error Parsing json",
			http.StatusInternalServerError)
		return
	}
	w.Write(inrec)
}

func (s *CoreService) setSensor(w http.ResponseWriter, req *http.Request) {
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
		database.SaveSensorConfig(s.db, config)
		json.NewEncoder(w).Encode(nil)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func (s *CoreService) swagger() {
	router := mux.NewRouter()
	sh := http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("/var/www/swaggerui/")))
	router.PathPrefix("/swaggerui/").Handler(sh)
	router.HandleFunc("/leds", s.getLeds).Methods("GET")
	router.HandleFunc("/led/{mac}", s.getLed).Methods("GET")
	router.HandleFunc("/sensors", s.getSensors).Methods("GET")
	router.HandleFunc("/sensor/{mac}", s.getSensor).Methods("GET")
	router.HandleFunc("/sensor/{mac}", s.setSensor).Methods("POST")
	log.Fatal(http.ListenAndServe(":8888", router))
}
