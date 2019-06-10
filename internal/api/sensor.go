package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/energieip/common-components-go/pkg/duser"

	ds "github.com/energieip/common-components-go/pkg/dsensor"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
)

func (api *API) readSensorConfig(w http.ResponseWriter, req *http.Request, mac string) {
	sensor, _ := database.GetSensorConfig(api.db, mac)
	if sensor == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}

	if sensor.Group == nil {
		group := 0
		sensor.Group = &group
	}
	json.NewEncoder(w).Encode(sensor)
}

func (api *API) getSensorSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	api.readSensorConfig(w, req, params["mac"])
}

func (api *API) setSensorSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	sensor := ds.SensorSetup{}
	err = json.Unmarshal(body, &sensor)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}

	event := make(map[string]interface{})
	event["sensorSetup"] = sensor
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) setSensorConfig(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
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
	event := make(map[string]interface{})
	event["sensor"] = sensor
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) getSensorStatus(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	mac := params["mac"]
	sensor := database.GetSensorStatus(api.db, mac)
	if sensor == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}

	if api.hasEnoughRight(w, req, sensor.Group) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(sensor)
}

func (api *API) removeSensorSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(req)
	mac := params["mac"]
	res := database.RemoveSensorConfig(api.db, mac)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}
