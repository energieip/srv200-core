package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/energieip/common-led-go/pkg/driverled"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readLedConfig(w http.ResponseWriter, mac string) {
	led := database.GetLedConfig(api.db, mac)
	if led == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}

	inrec, _ := json.MarshalIndent(led, "", "  ")
	w.Write(inrec)
}

func (api *API) getLedSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	api.readLedConfig(w, params["mac"])
}

func (api *API) setLedSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	led := driverled.LedSetup{}
	err = json.Unmarshal([]byte(body), &led)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	rlog.Info("Try to save ", led)
	database.SaveLedConfig(api.db, led)

	api.readLedConfig(w, led.Mac)
}

func (api *API) setLedConfig(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	led := driverled.LedConf{}
	err = json.Unmarshal([]byte(body), &led)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	if led.Group != nil {
		if *led.Group < 0 {
			api.sendError(w, APIErrorInvalidValue, "Invalid groupID "+strconv.Itoa(*led.Group))
			return
		}
		gr := database.GetGroupConfig(api.db, *led.Group)
		if gr == nil {
			api.sendError(w, APIErrorDeviceNotFound, "Group "+strconv.Itoa(*led.Group)+" not found")
			return
		}
	}
	event := make(map[string]interface{})
	event["led"] = led
	api.EventsToBackend <- event
	w.Write([]byte(""))
}

func (api *API) sendLedCommand(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	led := core.LedCmd{}
	err = json.Unmarshal([]byte(body), &led)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}
	rlog.Info("Received led cmd", led)
	event := make(map[string]interface{})
	event["ledCmd"] = led
	api.EventsToBackend <- event
	w.Write([]byte(""))
}

func (api *API) getLedStatus(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	mac := params["mac"]
	led := database.GetLedStatus(api.db, mac)
	if led == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}
	inrec, _ := json.MarshalIndent(led, "", "  ")
	w.Write(inrec)
}

func (api *API) removeLedSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	mac := params["mac"]
	res := database.RemoveLedConfig(api.db, mac)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+mac+" not found")
		return
	}
	w.Write([]byte(""))
}
