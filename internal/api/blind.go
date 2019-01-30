package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/energieip/common-components-go/pkg/dblind"

	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readBlindConfig(w http.ResponseWriter, mac string) {
	driver, _ := database.GetBlindConfig(api.db, mac)
	if driver == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}

	inrec, _ := json.MarshalIndent(driver, "", "  ")
	w.Write(inrec)
}

func (api *API) getBlindSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	api.readBlindConfig(w, params["mac"])
}

func (api *API) setBlindSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}
	setup := dblind.BlindSetup{}
	err = json.Unmarshal(body, &setup)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	database.SaveBlindConfig(api.db, setup)
	rlog.Info("Blind config" + setup.Mac + " saved")
	api.readBlindConfig(w, setup.Mac)
}

func (api *API) setBlindConfig(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	cfg := dblind.BlindConf{}
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
	event["blind"] = cfg
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) sendBlindCommand(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	cmd := core.BlindCmd{}
	err = json.Unmarshal([]byte(body), &cmd)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}
	rlog.Info("Received Blind cmd", cmd)
	event := make(map[string]interface{})
	event["blindCmd"] = cmd
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) getBlindStatus(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	mac := params["mac"]
	led := database.GetBlindStatus(api.db, mac)
	if led == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}
	inrec, _ := json.MarshalIndent(led, "", "  ")
	w.Write(inrec)
}

func (api *API) removeBlindSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	mac := params["mac"]
	res := database.RemoveBlindConfig(api.db, mac)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}
	w.Write([]byte("{}"))
}
