package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
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
	driver.NewFullMac = strings.ToUpper(driver.NewFullMac)
	driver.OldFullMac = strings.ToUpper(driver.OldFullMac)

	savedProject := database.GetProjectByMac(api.db, driver.OldFullMac)
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

	sheet3, _ := file.AddSheet("Sensors")

	row = sheet3.AddRow()
	cell = row.AddCell()
	cell.Value = "MAC"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Label"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Switch"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Temperature (°C)"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Hygrometry (%)"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Brightness (Lux)"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Presence"
	cell.SetStyle(boldStyle)

	sensors := database.GetSensorsStatus(api.db)
	for _, driv := range sensors {
		row = sheet3.AddRow()
		cell = row.AddCell()
		cell.Value = driv.Mac

		cell = row.AddCell()
		label := ""
		if driv.Label != nil {
			label = strings.Replace(*driv.Label, "_", "-", -1)
		}
		cell.Value = label

		cell = row.AddCell()
		cell.Value = driv.SwitchMac

		cell = row.AddCell()
		temperature := float32(driv.Temperature) / 10
		cell.Value = fmt.Sprintf("%f", temperature)

		cell = row.AddCell()
		hum := float32(driv.Humidity) / 10
		cell.Value = fmt.Sprintf("%f", hum)

		cell = row.AddCell()
		bright := float32(driv.Brightness) / 10
		cell.Value = fmt.Sprintf("%f", bright)

		cell = row.AddCell()
		if driv.Presence {
			cell.Value = "DETECTED"
		} else {
			cell.Value = "NONE"
		}
	}

	sheet4, _ := file.AddSheet("Nanosenses")

	row = sheet4.AddRow()
	cell = row.AddCell()
	cell.Value = "MAC"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Label"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Temperature (°C)"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "CO2 (ppm)"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "COV (ppm)"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Hygrometry (%)"
	cell.SetStyle(boldStyle)

	nanos := database.GetNanosStatus(api.db)
	for _, driv := range nanos {
		row = sheet4.AddRow()
		cell = row.AddCell()
		cell.Value = driv.Mac

		cell = row.AddCell()
		cell.Value = strings.Replace(driv.Label, "_", "-", -1)

		cell = row.AddCell()
		temperature := float32(driv.Temperature) / 10
		cell.Value = fmt.Sprintf("%f", temperature)

		cell = row.AddCell()
		co2 := float32(driv.CO2) / 10
		cell.Value = fmt.Sprintf("%f", co2)

		cell = row.AddCell()
		cov := float32(driv.COV) / 10
		cell.Value = fmt.Sprintf("%f", cov)

		cell = row.AddCell()
		hum := float32(driv.Hygrometry) / 10
		cell.Value = fmt.Sprintf("%f", hum)
	}

	sheet5, _ := file.AddSheet("Blinds")

	row = sheet5.AddRow()
	cell = row.AddCell()
	cell.Value = "MAC"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Label"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Switch"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Windows 1 Contact"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Windows 2 Contact"
	cell.SetStyle(boldStyle)

	blinds := database.GetBlindsStatus(api.db)
	for _, driv := range blinds {
		row = sheet5.AddRow()
		cell = row.AddCell()
		cell.Value = driv.Mac

		cell = row.AddCell()
		label := ""
		if driv.Label != nil {
			label = strings.Replace(*driv.Label, "_", "-", -1)
		}
		cell.Value = label

		cell = row.AddCell()
		cell.Value = driv.SwitchMac

		cell = row.AddCell()
		if driv.WindowStatus1 {
			cell.Value = "OPENED"
		} else {
			cell.Value = "CLOSED"
		}
		cell = row.AddCell()
		if driv.WindowStatus2 {
			cell.Value = "OPENED"
		} else {
			cell.Value = "CLOSED"
		}
	}

	sheet6, _ := file.AddSheet("HVACs")

	row = sheet6.AddRow()
	cell = row.AddCell()
	cell.Value = "MAC"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Label"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Switch"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Dew Sensor Status"
	cell.SetStyle(boldStyle)

	hvacs := database.GetHvacsStatus(api.db)
	for _, driv := range hvacs {
		row = sheet6.AddRow()
		cell = row.AddCell()
		cell.Value = driv.Mac

		cell = row.AddCell()
		label := ""
		if driv.Label != nil {
			label = strings.Replace(*driv.Label, "_", "-", -1)
		}
		cell.Value = label

		cell = row.AddCell()
		cell.Value = driv.SwitchMac

		cell = row.AddCell()
		if driv.DewSensor1 == 0 {
			cell.Value = "Inactive"
			cell.SetStyle(greenStyle)
		} else {
			cell.Value = "Active"
			cell.SetStyle(redStyle)
		}
	}

	sheet7, _ := file.AddSheet("Switchs")
	row = sheet7.AddRow()
	cell = row.AddCell()
	cell.Value = "MAC"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Label"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Profil"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Baes"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Puls 1"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Puls 2"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Puls 3"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Puls 4"
	cell.SetStyle(boldStyle)

	cell = row.AddCell()
	cell.Value = "Puls 5+"
	cell.SetStyle(boldStyle)

	switchs := database.GetSwitchsDump(api.db)
	for _, driv := range switchs {
		row = sheet7.AddRow()
		cell = row.AddCell()
		cell.Value = driv.Mac

		cell = row.AddCell()
		label := ""
		if driv.Label != nil {
			label = strings.Replace(*driv.Label, "_", "-", -1)
		}
		cell.Value = label

		cell = row.AddCell()
		profil := driv.Profil
		if profil == "" {
			profil = "none"
		}
		cell.Value = profil

		cell = row.AddCell()
		cell.Value = strconv.Itoa(driv.StateBaes)

		cell = row.AddCell()
		cell.Value = strconv.Itoa(driv.StatePuls1)

		cell = row.AddCell()
		cell.Value = strconv.Itoa(driv.StatePuls2)

		cell = row.AddCell()
		cell.Value = strconv.Itoa(driv.StatePuls3)

		cell = row.AddCell()
		cell.Value = strconv.Itoa(driv.StatePuls4)

		cell = row.AddCell()
		cell.Value = strconv.Itoa(driv.StatePuls5)
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

func (api *API) exportDBStart(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	if api.exportDBStatus != "" {
		url := "exportDB/status"
		http.Redirect(w, req, url, 201)
	}

	go func() {
		api.exportDBStatus = "running"
		cmd := exec.Command("rethinkdb", "dump")
		cmd.Dir = "/tmp"
		out, err := cmd.CombinedOutput()
		if err != nil {
			api.exportDBStatus = ""
			rlog.Error("cmd.Run() failed with status " + err.Error() + " : " + string(out))
			api.sendError(w, APIErrorDeviceNotFound, "Unable to open file", http.StatusInternalServerError)
			return
		}
		tempPath := ""
		for _, line := range strings.Split(string(out), "\n") {
			if !strings.HasPrefix(line, "Done") {
				continue
			}
			elts := strings.Split(line, " ")
			tempPath = elts[len(elts)-1]
			break
		}
		api.exportDBPath = tempPath
		api.exportDBStatus = "done"

	}()
	http.Redirect(w, req, req.URL.Path+"/result", 201)
}

func (api *API) exportDB(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	switch api.exportDBStatus {
	case "":
		rlog.Error("No database export running")
		api.sendError(w, APIErrorDeviceNotFound, "Unable to open file", http.StatusInternalServerError)

	case "running":
		http.Redirect(w, req, req.URL.Path, 201)

	case "done":
		fi, err := os.Stat(api.exportDBPath)
		if err != nil {
			rlog.Errorf("cannot access to file %v: %v", api.exportDBPath, err.Error())
			api.sendError(w, APIErrorDeviceNotFound, "Unable to open file", http.StatusInternalServerError)
			return
		}

		// Generate the server headers
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename="+path.Base(api.exportDBPath)+"")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Transfer-Encoding", "binary")
		w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
		w.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")
		api.exportDBStatus = ""
		http.ServeFile(w, req, api.exportDBPath)
	}
}
