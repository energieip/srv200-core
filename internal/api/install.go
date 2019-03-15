package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (api *API) installDriver(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body")
		return
	}

	driver := core.InstallDriver{}
	err = json.Unmarshal(body, &driver)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error())
		return
	}
	driver.Label = strings.Replace(driver.Label, "-", "_", -1)

	proj := core.Project{
		Label:   driver.Label,
		FullMac: &driver.FullMac,
	}

	submac := strings.SplitN(driver.FullMac, ":", 4)
	mac := submac[len(submac)-1]
	proj.Mac = &mac

	savedProject, _ := database.GetProject(api.db, driver.Label)
	if savedProject == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Unknow label "+driver.Label)
		return
	}
	proj.ModelName = savedProject.ModelName

	if savedProject.ModelName != nil {
		refModel := *savedProject.ModelName
		switch driver.ModelName {
		case "LED":
			if !strings.HasPrefix(refModel, "led") {
				api.sendError(w, APIErrorDeviceNotFound, "Unexpected Driver, expected "+refModel)
				return
			}
		case "BLIND":
			if !strings.HasPrefix(refModel, "bld") {
				api.sendError(w, APIErrorDeviceNotFound, "Unexpected Driver, expected "+refModel)
				return
			}
		case "SENSOR":
			if !strings.HasPrefix(refModel, "mca") {
				api.sendError(w, APIErrorDeviceNotFound, "Unexpected Driver, expected "+refModel)
				return
			}
		default:
			api.sendError(w, APIErrorDeviceNotFound, "Unexpected Driver type "+driver.ModelName)
			return
		}
	}

	err = database.SaveProject(api.db, proj)
	if err != nil {
		api.sendError(w, APIErrorDatabase, "Ifc information "+proj.Label+" cannot be added in database")
		return
	}

	rlog.Info("Driver " + driver.FullMac + " associate to " + proj.Label)
	api.readBim(w, proj.Label)
}
