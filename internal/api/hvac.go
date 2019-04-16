package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/energieip/common-components-go/pkg/dhvac"

	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readHvacConfig(w http.ResponseWriter, mac string) {
	driver, _ := database.GetHvacConfig(api.db, mac)
	if driver == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}

	inrec, _ := json.MarshalIndent(driver, "", "  ")
	w.Write(inrec)
}

func (api *API) getHvacSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	api.readHvacConfig(w, params["mac"])
}

func (api *API) setHvacSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}
	setup := dhvac.HvacSetup{}
	err = json.Unmarshal(body, &setup)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	database.SaveHvacConfig(api.db, setup)
	rlog.Info("Hvac config" + setup.Mac + " saved")
	api.readHvacConfig(w, setup.Mac)
}

func (api *API) setHvacConfig(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	cfg := dhvac.HvacConf{}
	err = json.Unmarshal(body, &cfg)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	if cfg.Group != nil {
		if *cfg.Group < 0 {
			api.sendError(w, APIErrorInvalidValue, "Invalid groupID "+strconv.Itoa(*cfg.Group))
			return
		}
		gr, _ := database.GetGroupConfig(api.db, *cfg.Group)
		if gr == nil {
			api.sendError(w, APIErrorDeviceNotFound, "Group "+strconv.Itoa(*cfg.Group)+" not found")
			return
		}
	}
	event := make(map[string]interface{})
	event["hvac"] = cfg
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) sendHvacCommand(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	cmd := core.HvacCmd{}
	err = json.Unmarshal([]byte(body), &cmd)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}
	rlog.Info("Received Hvac cmd", cmd)
	event := make(map[string]interface{})
	event["hvacCmd"] = cmd
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) getHvacStatus(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	mac := params["mac"]
	led := database.GetHvacStatus(api.db, mac)
	if led == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}
	inrec, _ := json.MarshalIndent(led, "", "  ")
	w.Write(inrec)
}

func (api *API) removeHvacSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	mac := params["mac"]
	res := database.RemoveHvacConfig(api.db, mac)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}
	w.Write([]byte("{}"))
}
