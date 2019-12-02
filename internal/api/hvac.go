package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/dhvac"
	"github.com/energieip/common-components-go/pkg/dserver"
	"github.com/energieip/common-components-go/pkg/duser"

	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readHvacConfig(w http.ResponseWriter, mac string) {
	driver, _ := database.GetHvacConfig(api.db, mac)
	if driver == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(driver)
}

func (api *API) getHvacSetup(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	mac := strings.ToUpper(params["mac"])
	api.readHvacConfig(w, mac)
}

func (api *API) setHvacSetup(w http.ResponseWriter, req *http.Request) {
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
	setup := dhvac.HvacSetup{}
	err = json.Unmarshal(body, &setup)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	setup.Mac = strings.ToUpper(setup.Mac)
	event := make(map[string]interface{})
	event["hvacSetup"] = setup
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) setHvacConfig(w http.ResponseWriter, req *http.Request) {
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

	cfg := dhvac.HvacConf{}
	err = json.Unmarshal(body, &cfg)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	cfg.Mac = strings.ToUpper(cfg.Mac)

	if cfg.Group != nil {
		if *cfg.Group < 0 {
			api.sendError(w, APIErrorInvalidValue, "Invalid groupID "+strconv.Itoa(*cfg.Group), http.StatusInternalServerError)
			return
		}
		gr, _ := database.GetGroupConfig(api.db, *cfg.Group)
		if gr == nil {
			name := "Group " + strconv.Itoa(*cfg.Group)
			group := gm.GroupConfig{
				Group:        *cfg.Group,
				FriendlyName: &name,
			}
			if cfg.Mac != "" {
				hvacs := []string{cfg.Mac}
				group.Hvacs = hvacs
			}
			eventGr := make(map[string]interface{})
			eventGr["group"] = group
			api.EventsToBackend <- eventGr
		}
	}
	event := make(map[string]interface{})
	event["hvac"] = cfg
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) sendHvacCommand(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	cmd := dserver.HvacCmd{}
	err = json.Unmarshal([]byte(body), &cmd)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	cmd.Mac = strings.ToUpper(cmd.Mac)
	dr, _ := database.GetHvacConfig(api.db, cmd.Mac)
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
	rlog.Info("Received Hvac cmd", cmd)
	event := make(map[string]interface{})
	event["hvacCmd"] = cmd
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) getHvacStatus(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	params := mux.Vars(req)
	mac := params["mac"]
	mac = strings.ToUpper(mac)
	dr := database.GetHvacStatus(api.db, mac)
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

func (api *API) removeHvacSetup(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	mac := params["mac"]
	mac = strings.ToUpper(mac)
	res := database.RemoveHvacConfig(api.db, mac)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}
