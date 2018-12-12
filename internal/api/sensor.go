package api

import (
	"encoding/json"
	"net/http"

	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"
)

func (api *API) getSensorSetup(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(req)
	sensor := database.GetSensorConfig(api.db, params["mac"])
	if sensor == nil {
		errCode := APIError{
			Code:    APIErrorDeviceNotFound,
			Message: "Device " + params["mac"] + "not found",
		}

		inrec, _ := json.MarshalIndent(errCode, "", "  ")
		rlog.Error(errCode.Message)
		http.Error(w, string(inrec),
			http.StatusInternalServerError)
		return
	}

	inrec, _ := json.MarshalIndent(sensor, "", "  ")
	w.Write(inrec)
}
