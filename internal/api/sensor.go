package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/energieip/common-sensor-go/pkg/driversensor"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readSensorConfig(w http.ResponseWriter, mac string) {
	sensor := database.GetSensorConfig(api.db, mac)
	if sensor == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}

	inrec, _ := json.MarshalIndent(sensor, "", "  ")
	w.Write(inrec)
}

func (api *API) getSensorSetup(w http.ResponseWriter, req *http.Request) {
	api.seDefaultHeader(w)
	params := mux.Vars(req)
	api.readSensorConfig(w, params["mac"])
}

func (api *API) setSensorSetup(w http.ResponseWriter, req *http.Request) {
	api.seDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	sensor := driversensor.SensorSetup{}
	err = json.Unmarshal([]byte(body), &sensor)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	rlog.Info("Try to save ", sensor)
	database.SaveSensorConfig(api.db, sensor)

	api.readSensorConfig(w, sensor.Mac)
}

func (api *API) getSensorStatus(w http.ResponseWriter, req *http.Request) {
	api.seDefaultHeader(w)
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
