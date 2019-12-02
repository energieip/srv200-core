package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/common-components-go/pkg/dwago"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
)

func (api *API) readWagoConfig(w http.ResponseWriter, req *http.Request, mac string) {
	led, _ := database.GetWagoConfig(api.db, mac)
	if led == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(led)
}

func (api *API) getWagoSetup(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(req)
	api.readLedConfig(w, req, strings.ToUpper(params["mac"]))
}

func (api *API) setWagoSetup(w http.ResponseWriter, req *http.Request) {
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
	wago := dwago.WagoSetup{}
	err = json.Unmarshal(body, &wago)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	wago.Mac = strings.ToUpper(wago.Mac)
	event := make(map[string]interface{})
	event["wagoSetup"] = wago
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) setWagoConfig(w http.ResponseWriter, req *http.Request) {
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

	wago := dwago.WagoConf{}
	err = json.Unmarshal(body, &wago)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	wago.Mac = strings.ToUpper(wago.Mac)
	event := make(map[string]interface{})
	event["wago"] = wago
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) getWagoStatus(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	params := mux.Vars(req)
	mac := strings.ToUpper(params["mac"])
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	wago := database.GetWagoStatus(api.db, mac)
	if wago == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(wago)
}

func (api *API) removeWagoSetup(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	mac := strings.ToUpper(params["mac"])
	res := database.RemoveWagoConfig(api.db, mac)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}
