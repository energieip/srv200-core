package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	gm "github.com/energieip/common-components-go/pkg/dgroup"
	dl "github.com/energieip/common-components-go/pkg/dled"
	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readLedConfig(w http.ResponseWriter, req *http.Request, mac string) {
	led, _ := database.GetLedConfig(api.db, mac)
	if led == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(led)
}

func (api *API) getLedSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(req)
	api.readLedConfig(w, req, params["mac"])
}

func (api *API) setLedSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}
	led := dl.LedSetup{}
	err = json.Unmarshal(body, &led)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}

	event := make(map[string]interface{})
	event["ledSetup"] = led
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) setLedConfig(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	led := dl.LedConf{}
	err = json.Unmarshal(body, &led)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}

	if led.Group != nil {
		if *led.Group < 0 {
			api.sendError(w, APIErrorInvalidValue, "Invalid groupID "+strconv.Itoa(*led.Group), http.StatusInternalServerError)
			return
		}
		gr, _ := database.GetGroupConfig(api.db, *led.Group)
		if gr == nil {
			name := "Group " + strconv.Itoa(*led.Group)
			group := gm.GroupConfig{
				Group:        *led.Group,
				FriendlyName: &name,
			}
			if led.Mac != "" {
				leds := []string{led.Mac}
				group.Leds = leds
			}
			eventGr := make(map[string]interface{})
			eventGr["group"] = group
			api.EventsToBackend <- eventGr
		}
	}
	event := make(map[string]interface{})
	event["led"] = led
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) sendLedCommand(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	led := core.LedCmd{}
	err = json.Unmarshal([]byte(body), &led)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	dr, _ := database.GetLedConfig(api.db, led.Mac)
	if dr == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+led.Mac+" not found", http.StatusInternalServerError)
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
	rlog.Info("Received led cmd", led)
	event := make(map[string]interface{})
	event["ledCmd"] = led
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) getLedStatus(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	mac := params["mac"]
	led := database.GetLedStatus(api.db, mac)
	if led == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	if api.hasEnoughRight(w, req, led.Group) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(led)
}

func (api *API) removeLedSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	mac := params["mac"]
	res := database.RemoveLedConfig(api.db, mac)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}
