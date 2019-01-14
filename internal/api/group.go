package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
)

func (api *API) readGroupConfig(w http.ResponseWriter, grID int) {
	group, _ := database.GetGroupConfig(api.db, grID)
	if group == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+strconv.Itoa(grID)+" not found")
		return
	}

	inrec, _ := json.MarshalIndent(group, "", "  ")
	w.Write(inrec)
}

func (api *API) getGroupSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	grID, err := strconv.Atoi(params["groupID"])
	if err != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+strconv.Itoa(grID)+" not found")
		return
	}
	api.readGroupConfig(w, grID)
}

func (api *API) setGroupSetup(w http.ResponseWriter, req *http.Request) {
	api.setGroupConfig(w, req)
}

func (api *API) setGroupConfig(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	gr := gm.GroupConfig{}
	err = json.Unmarshal(body, &gr)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}
	event := make(map[string]interface{})
	event["group"] = gr
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) sendGroupCommand(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	gr := core.GroupCmd{}
	err = json.Unmarshal(body, &gr)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}
	event := make(map[string]interface{})
	event["groupCmd"] = gr
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) removeGroupSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	grID := params["groupID"]
	i, err := strconv.Atoi(grID)
	if err != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+grID+" not found")
		return
	}
	res := database.RemoveGroupConfig(api.db, i)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+grID+" not found")
		return
	}
	w.Write([]byte("{}"))
}

func (api *API) getGroupStatus(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	grID := params["groupID"]
	i, err := strconv.Atoi(grID)
	if err != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+grID+" not found")
		return
	}
	res := database.GetGroupStatus(api.db, i)
	if res == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+grID+" not found")
		return
	}
	inrec, _ := json.MarshalIndent(res, "", "  ")
	w.Write(inrec)
}
