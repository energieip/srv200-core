package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/romana/rlog"

	"github.com/tealeg/xlsx"

	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
)

func (api *API) replaceDriver(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	driver := core.ReplaceDriver{}
	err = json.Unmarshal(body, &driver)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}

	savedProject := database.GetProjectByFullMac(api.db, driver.OldFullMac)
	if savedProject == nil {
		oldSubmac := strings.SplitN(driver.OldFullMac, ":", 4)
		oldMac := oldSubmac[len(oldSubmac)-1]
		savedProject = database.GetProjectByMac(api.db, oldMac)
	}

	if savedProject == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Unknow old driver "+driver.OldFullMac, http.StatusInternalServerError)
		return
	}

	event := make(map[string]interface{})
	event["replaceDriver"] = driver
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) installStatus(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	dt := time.Now()
	path := "/tmp/install_status.xlsx"
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var err error

	boldStyle := xlsx.NewStyle()
	boldFont := xlsx.NewFont(12, "Arial")
	boldFont.Bold = true
	boldStyle.Font = *boldFont
	boldStyle.ApplyFont = true

	redStyle := xlsx.NewStyle()
	fontred := xlsx.NewFont(10, "Arial")
	fontred.Color = "FFFF0000"
	redStyle.Font = *fontred
	redStyle.ApplyFont = true

	greenStyle := xlsx.NewStyle()
	fontgreen := xlsx.NewFont(10, "Arial")
	fontgreen.Color = "FF6CC24A"
	greenStyle.Font = *fontgreen
	greenStyle.ApplyFont = true

	file = xlsx.NewFile()

	sheet, err = file.AddSheet("Cables")
	if err != nil {
		rlog.Error(err.Error())
		api.sendError(w, APIErrorDeviceNotFound, "Unable to open new files", http.StatusInternalServerError)
		return
	}
	row = sheet.AddRow()
	cellLabel := row.AddCell()
	cellLabel.Value = "Label"
	cellLabel.SetStyle(boldStyle)
	cellStatus := row.AddCell()
	cellStatus.Value = "Status"
	cellStatus.SetStyle(boldStyle)

	projects := database.GetProjects(api.db)

	for _, project := range projects {
		row = sheet.AddRow()
		cellcable := row.AddCell()
		cellcable.Value = strings.Replace(project.Label, "_", "-", -1)
		cellCableStatus := row.AddCell()
		if project.Mac != nil {
			cellCableStatus.Value = "OK"
			cellCableStatus.SetStyle(greenStyle)
		} else {
			cellCableStatus.Value = "KO"
			cellCableStatus.SetStyle(redStyle)
		}
	}

	sheet2, _ := file.AddSheet("Drivers")

	row = sheet2.AddRow()
	cell := row.AddCell()
	cell.Value = "MAC"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Label"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Type"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Actif"
	cell.SetStyle(boldStyle)

	drivers := database.GetDrivers(api.db)
	for _, driv := range drivers {
		row = sheet2.AddRow()
		cell = row.AddCell()
		cell.Value = driv.Mac

		cell = row.AddCell()
		cell.Value = strings.Replace(driv.Label, "_", "-", -1)

		cell = row.AddCell()
		cell.Value = driv.Type

		cell = row.AddCell()
		if driv.Active {
			cell.Value = "OK"
			cell.SetStyle(greenStyle)
		} else {
			cell.Value = "KO"
			cell.SetStyle(redStyle)
		}
	}

	err = file.Save(path)
	if err != nil {
		rlog.Error(err.Error())
		api.sendError(w, APIErrorDeviceNotFound, "Unable to open new files", http.StatusInternalServerError)
		return
	}

	filename := dt.Format("01-02-2006") + "_install_status.xlsx"
	if err != nil {
		rlog.Error(err)
	}

	fi, err := os.Stat(path)

	// Generate the server headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename+"")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	w.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")

	http.ServeFile(w, req, path)
}

func (api *API) qrcodeGeneration(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	dt := time.Now()
	path := api.dataPath + "/stickers.pdf"
	filename := dt.Format("01-02-2006") + "_cable_qrcodes.pdf"

	fi, err := os.Stat(path)
	if err != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Unable to open file", http.StatusInternalServerError)
		return
	}

	// Generate the server headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename+"")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	w.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")

	http.ServeFile(w, req, path)
}
