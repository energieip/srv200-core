package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) readBim(w http.ResponseWriter, label string) {
	label = strings.Replace(label, "-", "_", -1)
	project, _ := database.GetProject(api.db, label)
	if project == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Could not found information on device "+label, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(*project)
}

func (api *API) getBim(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	label := params["label"]
	api.readBim(w, label)
}

func (api *API) removeBim(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	label := params["label"]
	label = strings.Replace(label, "-", "_", -1)
	res := database.RemoveProject(api.db, label)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+label+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}

func (api *API) setBim(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	proj := core.Project{}
	err = json.Unmarshal(body, &proj)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	if proj.ModelName != nil {
		model := strings.ToUpper(*proj.ModelName)
		proj.ModelName = &model
	}
	if proj.Mac != nil {
		mac := strings.ToUpper(*proj.Mac)
		proj.Mac = &mac
	}
	proj.Label = strings.Replace(proj.Label, "-", "_", -1)
	err = database.SaveProject(api.db, proj)
	if err != nil {
		api.sendError(w, APIErrorDatabase, "Ifc information "+proj.Label+" cannot be added in database", http.StatusInternalServerError)
		return
	}

	rlog.Info("IfcInfo for " + proj.Label + " saved")
	api.readBim(w, proj.Label)
}

func (api *API) readIfcInfo(w http.ResponseWriter, label string) {
	label = strings.Replace(label, "-", "_", -1)
	project, _ := database.GetProject(api.db, label)
	if project == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Could not found information on device "+label, http.StatusInternalServerError)
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
	json.NewEncoder(w).Encode(info)
}

func (api *API) getIfcInfo(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	label := params["label"]
	label = strings.Replace(label, "-", "_", -1)
	api.readIfcInfo(w, label)
}

func (api *API) removeIfcInfo(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	label := params["label"]
	label = strings.Replace(label, "-", "_", -1)
	res := database.RemoveProject(api.db, label)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+label+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}

func (api *API) setIfcInfo(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	ifcInfo := core.IfcInfo{}
	err = json.Unmarshal(body, &ifcInfo)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	ifcInfo.Mac = strings.ToUpper(ifcInfo.Mac)
	ifcInfo.DeviceType = strings.ToUpper(ifcInfo.DeviceType)
	ifcInfo.ModelName = strings.ToUpper(ifcInfo.ModelName)

	model := core.Model{
		Name:       ifcInfo.ModelName,
		Vendor:     ifcInfo.Vendor,
		URL:        ifcInfo.URL,
		DeviceType: ifcInfo.DeviceType,
	}
	err = database.SaveModel(api.db, model)
	if err != nil {
		api.sendError(w, APIErrorDatabase, "Ifc information "+ifcInfo.Label+" cannot be added in database", http.StatusInternalServerError)
		return
	}

	proj := core.Project{
		Label:     ifcInfo.Label,
		ModelName: &ifcInfo.ModelName,
		Mac:       &ifcInfo.Mac,
	}
	err = database.SaveProject(api.db, proj)
	if err != nil {
		api.sendError(w, APIErrorDatabase, "Ifc information "+ifcInfo.Label+" cannot be added in database", http.StatusInternalServerError)
		return
	}
	rlog.Info("IfcInfo for " + proj.Label + " saved")
	api.readIfcInfo(w, ifcInfo.Label)
}

func (api *API) getIfc(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	var infos []core.IfcInfo

	ifcs := database.GetIfcs(api.db)
	for _, info := range ifcs {
		infos = append(infos, info)
	}
	if ifcs != nil {
		inrec, _ := json.MarshalIndent(infos, "", "  ")
		w.Write(inrec)
	} else {
		w.Write([]byte("[]"))
	}
}
