package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
)

func (api *API) readServiceConfig(w http.ResponseWriter, name string) {
	service := database.GetServiceConfig(api.db, name)
	if service == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Service "+name+" not found")
		return
	}

	inrec, _ := json.MarshalIndent(service, "", "  ")
	w.Write(inrec)
}

func (api *API) getServiceSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	api.readServiceConfig(w, params["name"])
}

func (api *API) setServiceSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	service := pkg.Service{}
	err = json.Unmarshal(body, &service)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	err = database.SaveServiceConfig(api.db, service)
	if err != nil {
		api.sendError(w, APIErrorDatabase, "Service "+service.Name+" cannot be added in database")
		return
	}
	api.readServiceConfig(w, service.Name)
}

func (api *API) removeServiceSetup(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	params := mux.Vars(req)
	name := params["name"]
	res := database.RemoveServiceConfig(api.db, name)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Service "+name+" not found")
		return
	}
	w.Write([]byte("{}"))
}
