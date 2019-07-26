package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/common-components-go/pkg/tools"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
)

func (api *API) installDriver(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	driver := core.InstallDriver{}
	err = json.Unmarshal(body, &driver)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	driver.Mac = strings.ToUpper(driver.Mac)
	driver.Device = strings.ToUpper(driver.Device)
	driver.Label = strings.Replace(driver.Label, "-", "_", -1)
	driver.Label = strings.ToUpper(driver.Label)
	savedProject, _ := database.GetProject(api.db, driver.Label)
	if savedProject == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Unknow label "+driver.Label, http.StatusInternalServerError)
		return
	}

	if savedProject.ModelName != nil {
		refModel := tools.Model2Type(*savedProject.ModelName)
		if refModel != driver.Device {
			api.sendError(w, APIErrorDeviceNotFound, "Unexpected Driver, expected "+refModel, http.StatusInternalServerError)
			return
		}
	}
	event := make(map[string]interface{})
	event["installDriver"] = driver
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}
