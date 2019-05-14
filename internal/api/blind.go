package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/energieip/common-components-go/pkg/dblind"
	"github.com/energieip/common-components-go/pkg/duser"

	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readBlindConfig(w http.ResponseWriter, mac string) {
	driver, _ := database.GetBlindConfig(api.db, mac)
	if driver == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(driver)
}

func (api *API) getBlindSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	api.readBlindConfig(w, params["mac"])
}

func (api *API) setBlindSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}
	setup := dblind.BlindSetup{}
	err = json.Unmarshal(body, &setup)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}

	database.SaveBlindConfig(api.db, setup)
	rlog.Info("Blind config" + setup.Mac + " saved")
	api.readBlindConfig(w, setup.Mac)
}

func (api *API) setBlindConfig(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	cfg := dblind.BlindConf{}
	err = json.Unmarshal(body, &cfg)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}

	if cfg.Group != nil {
		if *cfg.Group < 0 {
			api.sendError(w, APIErrorInvalidValue, "Invalid groupID "+strconv.Itoa(*cfg.Group), http.StatusInternalServerError)
			return
		}
		gr, _ := database.GetGroupConfig(api.db, *cfg.Group)
		if gr == nil {
			api.sendError(w, APIErrorDeviceNotFound, "Group "+strconv.Itoa(*cfg.Group)+" not found", http.StatusInternalServerError)
			return
		}
	}
	event := make(map[string]interface{})
	event["blind"] = cfg
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) sendBlindCommand(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	cmd := core.BlindCmd{}
	err = json.Unmarshal([]byte(body), &cmd)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	dr, _ := database.GetBlindConfig(api.db, cmd.Mac)
	if dr == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+cmd.Mac+" not found", http.StatusInternalServerError)
		return
	}

	if dr.Group == nil {
		group := 0
		dr.Group = &group
	}
	if api.hasEnoughRight(w, req, *dr.Group) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	rlog.Info("Received Blind cmd", cmd)
	event := make(map[string]interface{})
	event["blindCmd"] = cmd
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) getBlindStatus(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	mac := params["mac"]
	dr := database.GetBlindStatus(api.db, mac)
	if dr == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	if api.hasEnoughRight(w, req, dr.Group) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(dr)
}

func (api *API) removeBlindSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	mac := params["mac"]
	res := database.RemoveBlindConfig(api.db, mac)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}
