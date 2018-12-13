package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readModelInfo(w http.ResponseWriter, modelName string) {
	model := database.GetModel(api.db, modelName)
	if model == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Model "+modelName+" not found")
		return
	}
	inrec, _ := json.MarshalIndent(model, "", "  ")
	w.Write(inrec)
}

func (api *API) getModelInfo(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	label := params["modelName"]
	api.readModelInfo(w, label)
}

func (api *API) removeModelInfo(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	modelName := params["modelName"]
	res := database.RemoveModel(api.db, modelName)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+modelName+" not found")
		return
	}
	w.Write([]byte(""))
}

func (api *API) setModelInfo(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	model := core.Model{}
	err = json.Unmarshal([]byte(body), &model)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	rlog.Info("Try to save ", model)
	database.SaveModel(api.db, model)
	api.readModelInfo(w, model.Name)
}
