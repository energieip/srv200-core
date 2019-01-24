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

func (api *API) readBim(w http.ResponseWriter, label string) {
	project, _ := database.GetProject(api.db, label)
	if project == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Could not found information on device "+label)
		return
	}
	inrec, _ := json.MarshalIndent(*project, "", "  ")
	w.Write(inrec)
}

func (api *API) getBim(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	label := params["label"]
	api.readBim(w, label)
}

func (api *API) removeBim(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	label := params["label"]
	res := database.RemoveProject(api.db, label)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+label+" not found")
		return
	}
	w.Write([]byte("{}"))
}

func (api *API) setBim(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	proj := core.Project{}
	err = json.Unmarshal(body, &proj)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	rlog.Info("Try to save IfcInfo", proj)
	err = database.SaveProject(api.db, proj)
	if err != nil {
		api.sendError(w, APIErrorDatabase, "Ifc information "+proj.Label+" cannot be added in database")
		return
	}
	api.readBim(w, proj.Label)
}

func (api *API) readIfcInfo(w http.ResponseWriter, label string) {
	project, _ := database.GetProject(api.db, label)
	if project == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Could not found information on device "+label)
		return
	}
	info := core.IfcInfo{
		Label: label,
	}
	if project.ModelName != nil {
		model := database.GetModel(api.db, *project.ModelName)
		info.ModelName = model.Name
		info.Vendor = model.Vendor
		info.URL = model.URL
		info.DeviceType = model.DeviceType
	}
	if project.Mac != nil {
		info.Mac = *project.Mac
	}

	inrec, _ := json.MarshalIndent(info, "", "  ")
	w.Write(inrec)
}

func (api *API) getIfcInfo(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	label := params["label"]
	api.readIfcInfo(w, label)
}

func (api *API) removeIfcInfo(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	label := params["label"]
	res := database.RemoveProject(api.db, label)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+label+" not found")
		return
	}
	w.Write([]byte("{}"))
}

func (api *API) setIfcInfo(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	ifcInfo := core.IfcInfo{}
	err = json.Unmarshal(body, &ifcInfo)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	rlog.Info("Try to save IfcInfo", ifcInfo)
	model := core.Model{
		Name:       ifcInfo.ModelName,
		Vendor:     ifcInfo.Vendor,
		URL:        ifcInfo.URL,
		DeviceType: ifcInfo.DeviceType,
	}
	err = database.SaveModel(api.db, model)
	if err != nil {
		api.sendError(w, APIErrorDatabase, "Ifc information "+ifcInfo.Label+" cannot be added in database")
		return
	}

	proj := core.Project{
		Label:     ifcInfo.Label,
		ModelName: &ifcInfo.ModelName,
		Mac:       &ifcInfo.Mac,
	}
	err = database.SaveProject(api.db, proj)
	if err != nil {
		api.sendError(w, APIErrorDatabase, "Ifc information "+ifcInfo.Label+" cannot be added in database")
		return
	}
	api.readIfcInfo(w, ifcInfo.Label)
}

func (api *API) getIfc(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	var infos []core.IfcInfo

	ifcs := database.GetIfcs(api.db)
	for _, info := range ifcs {
		infos = append(infos, info)
	}
	inrec, _ := json.MarshalIndent(infos, "", "  ")
	w.Write(inrec)
}
