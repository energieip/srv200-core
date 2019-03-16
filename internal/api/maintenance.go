package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
)

func (api *API) replaceDriver(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	driver := core.ReplaceDriver{}
	err = json.Unmarshal(body, &driver)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}

	savedProject := database.GetProjectByFullMac(api.db, driver.OldFullMac)
	if savedProject == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Unknow old driver "+driver.OldFullMac)
		return
	}

	event := make(map[string]interface{})
	event["replaceDriver"] = driver
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}
