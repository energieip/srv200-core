package api

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/romana/rlog"
)

func (api *API) modbusTableAPI(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	dt := time.Now()
	path := "/tmp/modbus_table.xlsx"

	filename := dt.Format("01-02-2006") + "_modbus_table.xlsx"

	fi, err := os.Stat(path)
	if err != nil {
		rlog.Error(err.Error())
		api.sendError(w, APIErrorDeviceNotFound, "Unable to open new files", http.StatusInternalServerError)
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
