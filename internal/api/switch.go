package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/energieip/common-components-go/pkg/dserver"
	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
)

func (api *API) readSwitchConfig(w http.ResponseWriter, mac string) {
	device, _ := database.GetSwitchConfig(api.db, mac)
	if device == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Switch "+mac+" not found", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(device)
}

func (api *API) getSwitchSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	api.readSwitchConfig(w, strings.ToUpper(params["mac"]))
}

func (api *API) setSwitchSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	api.setSwitchConfig(w, req)
}

func (api *API) setSwitchConfig(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	device := dserver.SwitchConfig{}
	err = json.Unmarshal(body, &device)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	if device.Mac != nil {
		mac := strings.ToUpper(*device.Mac)
		device.Mac = &mac
	}
	event := make(map[string]interface{})
	event["switch"] = device
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) removeSwitchSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	mac := strings.ToUpper(params["mac"])
	res := database.RemoveSwitchConfig(api.db, mac)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Switch "+mac+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}
