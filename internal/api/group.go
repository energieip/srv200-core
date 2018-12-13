package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/energieip/common-group-go/pkg/groupmodel"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readGroupConfig(w http.ResponseWriter, grID int) {
	group := database.GetGroupConfig(api.db, grID)
	if group == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+strconv.Itoa(grID)+" not found")
		return
	}

	inrec, _ := json.MarshalIndent(group, "", "  ")
	w.Write(inrec)
}

func (api *API) getGroupSetup(w http.ResponseWriter, req *http.Request) {
	api.seDefaultHeader(w)
	params := mux.Vars(req)
	grID, err := strconv.Atoi(params["groupID"])
	if err != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+strconv.Itoa(grID)+" not found")
		return
	}
	api.readGroupConfig(w, grID)
}

func (api *API) setGroupSetup(w http.ResponseWriter, req *http.Request) {
	api.seDefaultHeader(w)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	gr := groupmodel.GroupConfig{}
	err = json.Unmarshal([]byte(body), &gr)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	rlog.Info("Try to save ", gr)
	database.SaveGroupConfig(api.db, gr)

	api.readGroupConfig(w, gr.Group)
}

func (api *API) removeGroupSetup(w http.ResponseWriter, req *http.Request) {
	api.seDefaultHeader(w)
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
	w.Write([]byte(""))
}
