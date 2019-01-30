package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	ds "github.com/energieip/common-components-go/pkg/dsensor"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readSensorConfig(w http.ResponseWriter, mac string) {
	sensor, _ := database.GetSensorConfig(api.db, mac)
	if sensor == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}

	inrec, _ := json.MarshalIndent(sensor, "", "  ")
	w.Write(inrec)
}

func (api *API) getSensorSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	api.readSensorConfig(w, params["mac"])
}

func (api *API) setSensorSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	sensor := ds.SensorSetup{}
	err = json.Unmarshal(body, &sensor)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	database.SaveSensorConfig(api.db, sensor)
	rlog.Info("Save sensor configuration ", sensor.Mac)
	api.readSensorConfig(w, sensor.Mac)
}

func (api *API) setSensorConfig(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	sensor := ds.SensorConf{}
	err = json.Unmarshal(body, &sensor)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}
	event := make(map[string]interface{})
	event["sensor"] = sensor
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) getSensorStatus(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	mac := params["mac"]
	sensor := database.GetSensorStatus(api.db, mac)
	if sensor == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}
	inrec, _ := json.MarshalIndent(sensor, "", "  ")
	w.Write(inrec)
}

func (api *API) removeSensorSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	mac := params["mac"]
	res := database.RemoveSensorConfig(api.db, mac)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}
	w.Write([]byte("{}"))
}
