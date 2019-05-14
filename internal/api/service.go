package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/energieip/common-components-go/pkg/duser"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
)

func (api *API) readServiceConfig(w http.ResponseWriter, name string) {
	service := database.GetServiceConfig(api.db, name)
	if service == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Service "+name+" not found", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(service)
}

func (api *API) getServiceSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	api.readServiceConfig(w, params["name"])
}

func (api *API) setServiceSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	service := pkg.Service{}
	err = json.Unmarshal(body, &service)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = database.SaveServiceConfig(api.db, service)
	if err != nil {
		api.sendError(w, APIErrorDatabase, "Service "+service.Name+" cannot be added in database", http.StatusInternalServerError)
		return
	}
	api.readServiceConfig(w, service.Name)
}

func (api *API) removeServiceSetup(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	name := params["name"]
	res := database.RemoveServiceConfig(api.db, name)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Service "+name+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}
