package api

import (
	"encoding/json"
	"log"
	"net/http"

	"io/ioutil"
	"strings"

	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/mux"
	"github.com/romana/rlog"

	"github.com/energieip/srv200-coreservice-go/internal/core"
)

type InternalAPI struct {
	db              database.Database
	EventsToBackend chan map[string]interface{}
	certificate     string
	keyfile         string
	apiIP           string
	apiPort         string
	apiPassword     string
	browsingFolder  string
	dataPath        string
}

//InitInternalAPI start API connection
func InitInternalAPI(db database.Database,
	conf pkg.ServiceConfig) *InternalAPI {
	api := InternalAPI{
		db:              db,
		apiIP:           conf.InternalAPI.IP,
		apiPort:         conf.InternalAPI.Port,
		apiPassword:     conf.InternalAPI.Password,
		EventsToBackend: make(chan map[string]interface{}),
		certificate:     conf.InternalAPI.CertPath,
		keyfile:         conf.InternalAPI.KeyPath,
		browsingFolder:  conf.InternalAPI.BrowsingFolder,
		dataPath:        conf.DataPath,
	}
	go api.swagger()
	return &api
}

func (api *InternalAPI) setDefaultHeader(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
}

func (api *InternalAPI) sendError(w http.ResponseWriter, errorCode int, message string, httpStatus int) {
	errCode := APIError{
		Code:    errorCode,
		Message: message,
	}

	inrec, _ := json.MarshalIndent(errCode, "", "  ")
	rlog.Error(errCode.Message)
	http.Error(w, string(inrec), httpStatus)
}

func (api *InternalAPI) getFunctions(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w, req)
	functions := []string{"/versions"}
	apiInfo := APIFunctions{
		Functions: functions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *InternalAPI) getAPIs(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w, req)
	versions := []string{"v1.0"}
	apiInfo := APIInfo{
		Versions: versions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *InternalAPI) getV1Functions(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w, req)
	apiV1 := "/v1.0"
	functions := []string{
		apiV1 + "/command/led", apiV1 + "/command/blind", apiV1 + "/command/hvac",
		apiV1 + "/command/group", apiV1 + "/dump",
	}
	apiInfo := APIFunctions{
		Functions: functions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *InternalAPI) getDump(w http.ResponseWriter, req *http.Request) {
	var leds []DumpLed
	var sensors []DumpSensor
	var switchs []DumpSwitch
	var blinds []DumpBlind
	var hvacs []DumpHvac
	var wagos []DumpWago
	var groups []DumpGroup
	macs := make(map[string]bool)
	driversMac := make(map[string]bool)
	labels := make(map[string]bool)
	filterByMac := false
	MacsParam := req.FormValue("macs")
	if MacsParam != "" {
		tempMac := strings.Split(MacsParam, ",")

		for _, v := range tempMac {
			macs[v] = true
			filterByMac = true
		}
	}

	filterByLabel := false
	LabelsParam := req.FormValue("labels")
	if LabelsParam != "" {
		for _, v := range strings.Split(LabelsParam, ",") {
			labels[v] = true
			filterByLabel = true
		}
	}

	lights := database.GetLedsStatusByLabel(api.db)
	lightsConfig := database.GetLedsConfigByLabel(api.db)
	cells := database.GetSensorsStatusByLabel(api.db)
	cellsConfig := database.GetSensorsConfigByLabel(api.db)
	blds := database.GetBlindsStatusByLabel(api.db)
	bldsConfig := database.GetBlindsConfigByLabel(api.db)
	hvcs := database.GetHvacsStatusByLabel(api.db)
	hvcsConfig := database.GetHvacsConfigByLabel(api.db)
	wags := database.GetWagosStatusByLabel(api.db)
	wagosConfig := database.GetWagosConfigByLabel(api.db)
	switchElts := database.GetSwitchsDumpByLabel(api.db)
	switchEltsConfig := database.GetSwitchsConfigByLabel(api.db)

	ifcs := database.GetIfcs(api.db)
	for _, ifc := range ifcs {
		if filterByLabel {
			if _, ok := labels[ifc.Label]; !ok {
				continue
			}
		}
		if filterByMac {
			if _, ok := macs[ifc.Mac]; !ok {
				continue
			}
		}
		driversMac[ifc.Mac] = true

		switch ifc.DeviceType {
		case "led":
			dump := DumpLed{}
			led, ok := lights[ifc.Label]
			gr := 0
			if ok {
				dump.Status = led
				gr = led.Group
			}
			config, ok := lightsConfig[ifc.Label]
			if ok {
				dump.Config = config
				if gr == 0 && config.Group != nil {
					gr = *config.Group
				}
			}
			dump.Ifc = ifc

			leds = append(leds, dump)
		case "sensor":
			dump := DumpSensor{}
			gr := 0
			sensor, ok := cells[ifc.Label]
			if ok {
				dump.Status = sensor
				gr = sensor.Group
			}
			config, ok := cellsConfig[ifc.Label]
			if ok {
				dump.Config = config
				if gr == 0 && config.Group != nil {
					gr = *config.Group
				}
			}
			dump.Ifc = ifc
			sensors = append(sensors, dump)
		case "blind":
			dump := DumpBlind{}
			gr := 0
			bld, ok := blds[ifc.Label]
			if ok {
				dump.Status = bld
				gr = bld.Group
			}
			config, ok := bldsConfig[ifc.Label]
			if ok {
				dump.Config = config
				if gr == 0 && config.Group != nil {
					gr = *config.Group
				}
			}
			dump.Ifc = ifc
			blinds = append(blinds, dump)
		case "hvac":
			dump := DumpHvac{}
			gr := 0
			hvac, ok := hvcs[ifc.Label]
			if ok {
				dump.Status = hvac
				gr = hvac.Group
			}
			config, ok := hvcsConfig[ifc.Label]
			if ok {
				dump.Config = config
				if gr == 0 && config.Group != nil {
					gr = *config.Group
				}
			}
			dump.Ifc = ifc
			hvacs = append(hvacs, dump)
		case "wago":
			dump := DumpWago{}
			wago, ok := wags[ifc.Label]
			if ok {
				dump.Status = wago
			}
			config, ok := wagosConfig[ifc.Label]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			wagos = append(wagos, dump)
		case "switch":
			dump := DumpSwitch{}
			switchElt, ok := switchElts[ifc.Label]
			if ok {
				dump.Status = switchElt
			}
			config, ok := switchEltsConfig[ifc.Label]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			switchs = append(switchs, dump)
		}
	}

	groupsStatus := database.GetGroupsStatus(api.db)
	groupsConfig := database.GetGroupConfigs(api.db, driversMac)

	for _, gr := range groupsConfig {
		dump := DumpGroup{}
		grStatus, ok := groupsStatus[gr.Group]
		if ok {
			dump.Status = grStatus
		}
		dump.Config = gr
		groups = append(groups, dump)
	}

	dump := Dump{
		Leds:    leds,
		Sensors: sensors,
		Blinds:  blinds,
		Hvacs:   hvacs,
		Wagos:   wagos,
		Switchs: switchs,
		Groups:  groups,
	}

	inrec, _ := json.MarshalIndent(dump, "", "  ")
	w.Write(inrec)
}

func (api *InternalAPI) sendLedCommand(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	led := core.LedCmd{}
	err = json.Unmarshal([]byte(body), &led)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	dr, _ := database.GetLedConfig(api.db, led.Mac)
	if dr == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+led.Mac+" not found", http.StatusInternalServerError)
		return
	}

	rlog.Info("Received led cmd", led)
	event := make(map[string]interface{})
	event["ledCmd"] = led
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) sendBlindCommand(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	cmd := core.BlindCmd{}
	err = json.Unmarshal([]byte(body), &cmd)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	dr, _ := database.GetBlindConfig(api.db, cmd.Mac)
	if dr == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+cmd.Mac+" not found", http.StatusInternalServerError)
		return
	}

	rlog.Info("Received Blind cmd", cmd)
	event := make(map[string]interface{})
	event["blindCmd"] = cmd
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) sendHvacCommand(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	cmd := core.HvacCmd{}
	err = json.Unmarshal([]byte(body), &cmd)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	dr, _ := database.GetHvacConfig(api.db, cmd.Mac)
	if dr == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Device "+cmd.Mac+" not found", http.StatusInternalServerError)
		return
	}

	rlog.Info("Received Hvac cmd", cmd)
	event := make(map[string]interface{})
	event["hvacCmd"] = cmd
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) sendGroupCommand(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	gr := core.GroupCmd{}
	err = json.Unmarshal(body, &gr)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}

	event := make(map[string]interface{})
	event["groupCmd"] = gr
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *InternalAPI) swagger() {
	router := mux.NewRouter()
	sh := http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("/media/userdata/www/swaggerui/")))
	router.PathPrefix("/swaggerui/").Handler(sh)

	// API v1.0
	apiV1 := "/v1.0"
	router.HandleFunc(apiV1+"/functions", api.getV1Functions).Methods("GET")

	//dump API
	router.HandleFunc(apiV1+"/dump", api.getDump).Methods("GET")

	//command API
	router.HandleFunc(apiV1+"/command/led", api.sendLedCommand).Methods("POST")
	router.HandleFunc(apiV1+"/command/blind", api.sendBlindCommand).Methods("POST")
	router.HandleFunc(apiV1+"/command/hvac", api.sendHvacCommand).Methods("POST")
	router.HandleFunc(apiV1+"/command/group", api.sendGroupCommand).Methods("POST")

	//unversionned API
	router.HandleFunc("/versions", api.getAPIs).Methods("GET")
	router.HandleFunc("/functions", api.getFunctions).Methods("GET")

	if api.browsingFolder != "" {
		sh2 := http.StripPrefix("/", http.FileServer(http.Dir(api.browsingFolder)))
		router.PathPrefix("/").Handler(sh2)
	}

	log.Fatal(http.ListenAndServeTLS(api.apiIP+":"+api.apiPort, api.certificate, api.keyfile, router))
}
