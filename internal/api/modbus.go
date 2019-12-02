package api

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/romana/rlog"

	"github.com/energieip/common-components-go/pkg/duser"
)

func (api *API) modbusTableAPI(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	url := "https://127.0.0.1:8888/v1.0/generateDoc"

	reqTmp, _ := http.NewRequest("POST", url, nil)
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	reqTmp.Close = true
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(reqTmp)

	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		api.sendError(w, APIErrorDeviceNotFound, "Unable to open new files", http.StatusInternalServerError)
		return
	}

	path := "/tmp/tmp_modbus_table.xlsx"
	out, err := os.Create(path)
	if err != nil {
		rlog.Error(err.Error())
		api.sendError(w, APIErrorDeviceNotFound, "Unable to open new files", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, resp.Body)

	dt := time.Now()
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
