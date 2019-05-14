package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readModelInfo(w http.ResponseWriter, modelName string) {
	model := database.GetModel(api.db, modelName)
	if model == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Model "+modelName+" not found", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(model)
}

func (api *API) getModelInfo(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	label := params["modelName"]
	api.readModelInfo(w, label)
}

func (api *API) removeModelInfo(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	modelName := params["modelName"]
	res := database.RemoveModel(api.db, modelName)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+modelName+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}

func (api *API) setModelInfo(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	model := core.Model{}
	err = json.Unmarshal(body, &model)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}

	database.SaveModel(api.db, model)
	rlog.Info("Model " + model.Name + " saved")
	api.readModelInfo(w, model.Name)
}
