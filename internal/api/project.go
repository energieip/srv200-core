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

//IfcInfo ifc component description
type IfcInfo struct {
	Label     string `json:"label"` //cable label
	ModelName string `json:"modelName"`
	Mac       string `json:"mac"` //device Mac address
	Vendor    string `json:"vendor"`
	URL       string `json:"url"`
}

func (api *API) readIfcInfo(w http.ResponseWriter, label string) {
	project := database.GetProject(api.db, label)
	if project == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Could not found information on device "+label)
		return
	}
	model := database.GetModel(api.db, project.ModelName)
	info := IfcInfo{
		Label:     label,
		ModelName: model.Name,
		Mac:       project.Mac,
		Vendor:    model.Vendor,
		URL:       model.URL,
	}
	inrec, _ := json.MarshalIndent(info, "", "  ")
	w.Write(inrec)
}

func (api *API) getIfcInfo(w http.ResponseWriter, req *http.Request) {
	api.seDefaultHeader(w)
	params := mux.Vars(req)
	label := params["label"]
	api.readIfcInfo(w, label)
}

func (api *API) removeIfcInfo(w http.ResponseWriter, req *http.Request) {
	api.seDefaultHeader(w)
	params := mux.Vars(req)
	label := params["label"]
	res := database.RemoveModel(api.db, label)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+label+" not found")
		return
	}
	res = database.RemoveProject(api.db, label)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+label+" not found")
		return
	}
	w.Write([]byte(""))
}

func (api *API) setIfcInfo(w http.ResponseWriter, req *http.Request) {
	api.seDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	ifcInfo := IfcInfo{}
	err = json.Unmarshal([]byte(body), &ifcInfo)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	rlog.Info("Try to save ", ifcInfo)
	model := core.Model{
		Name:   ifcInfo.ModelName,
		Vendor: ifcInfo.Vendor,
		URL:    ifcInfo.URL,
	}
	err = database.SaveModel(api.db, model)
	if err != nil {
		api.sendError(w, APIErrorDatabase, "Ifc information "+ifcInfo.Label+" cannot be added in database")
		return
	}

	proj := core.Project{
		Label:     ifcInfo.Label,
		ModelName: ifcInfo.ModelName,
		Mac:       ifcInfo.Mac,
	}
	err = database.SaveProject(api.db, proj)
	if err != nil {
		api.sendError(w, APIErrorDatabase, "Ifc information "+ifcInfo.Label+" cannot be added in database")
		return
	}
	api.readIfcInfo(w, ifcInfo.Label)
}
