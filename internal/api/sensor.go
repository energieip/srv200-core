package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	gm "github.com/energieip/common-components-go/pkg/dgroup"
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
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	mac := strings.ToUpper(params["mac"])
	api.readSensorConfig(w, req, mac)
}

func (api *API) setSensorSetup(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
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
	sensor.Mac = strings.ToUpper(sensor.Mac)
	event := make(map[string]interface{})
	event["sensorSetup"] = sensor
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) setSensorConfig(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
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

func (api *API) getSensorStatus(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	params := mux.Vars(req)
	mac := params["mac"]
	mac = strings.ToUpper(mac)
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
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(req)
	mac := params["mac"]
	mac = strings.ToUpper(mac)
	res := database.RemoveSensorConfig(api.db, mac)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}
