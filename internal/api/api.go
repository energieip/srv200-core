package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/energieip/common-components-go/pkg/dnanosense"

	"github.com/energieip/common-components-go/pkg/pconst"

	"github.com/energieip/common-components-go/pkg/dwago"

	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/common-components-go/pkg/tools"
	"github.com/mitchellh/mapstructure"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/energieip/srv200-coreservice-go/internal/history"

	"github.com/energieip/common-components-go/pkg/dblind"
	"github.com/energieip/common-components-go/pkg/dhvac"

	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/romana/rlog"
)

//InitAPI start API connection
func InitAPI(db database.Database, historydb history.HistoryDb, eventsAPI chan map[string]interface{},
	eventsConso chan core.EventConsumption, uploadValue *string, conf pkg.ServiceConfig) *API {
	api := API{
		db:              db,
		apiIP:           conf.ExternalAPI.IP,
		apiPort:         conf.ExternalAPI.Port,
		apiPassword:     conf.ExternalAPI.Password,
		access:          cmap.New(),
		historydb:       historydb,
		eventsAPI:       eventsAPI,
		eventsConso:     eventsConso,
		EventsToBackend: make(chan map[string]interface{}),
		clients:         make(map[*websocket.Conn]duser.UserAccess),
		clientsConso:    make(map[*websocket.Conn]duser.UserAccess),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		certificate:    conf.ExternalAPI.CertPath,
		keyfile:        conf.ExternalAPI.KeyPath,
		browsingFolder: conf.ExternalAPI.BrowsingFolder,
		dataPath:       conf.DataPath,
		uploadValue:    uploadValue,
	}
	go api.swagger()
	return &api
}

func (api *API) verification(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValue := ""
		tokenCookie, err := r.Cookie(TokenName)

		if err != nil || tokenCookie == nil {
			//Check header
			authorizationHeader := r.Header.Get("Authorization")
			if authorizationHeader != "" {
				bearerToken := strings.Split(authorizationHeader, " ")
				if len(bearerToken) > 1 {
					tokenValue = bearerToken[1]
				} else {
					tokenValue = authorizationHeader
				}
			}
		} else {
			tokenValue = tokenCookie.Value
		}
		api.setDefaultHeader(w, r)

		if tokenValue == "" {
			api.sendError(w, APIErrorUnauthorized, "Unauthorized access", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenValue, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error")
			}
			return []byte(api.apiPassword), nil
		})

		switch err.(type) {
		case nil:
			_, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				if ok {
					_, ok = api.access.Get(tokenValue)
					// case token deprecated cleanup needed
					api.access.Remove(tokenValue)
				}
				api.sendError(w, APIErrorUnauthorized, "Unauthorized access", http.StatusUnauthorized)
				return
			}

			//check in map
			user, ok := api.access.Get(tokenValue)
			if !ok || user == nil {
				api.sendError(w, APIErrorExpiredToken, "Invalid Token", http.StatusUnauthorized)
				return
			}
			userAccess, _ := duser.ToUserAccess(user)
			context.Set(r, "decoded", *userAccess)
			context.Set(r, "token", tokenValue)
			next(w, r)

		case *jwt.ValidationError:
			vErr := err.(*jwt.ValidationError)

			switch vErr.Errors {
			case jwt.ValidationErrorExpired:
				api.sendError(w, APIErrorExpiredToken, "Expired Token", http.StatusUnauthorized)
				return

			default:
				api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
				return
			}

		default:
			api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
			return
		}
	})
}

func (api *API) setDefaultHeader(w http.ResponseWriter, req *http.Request) {
	header := "https://" + req.Host
	w.Header().Set("Access-Control-Allow-Origin", header)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, *")
	w.Header().Set("Content-Type", "application/json")
}

func (api *API) sendError(w http.ResponseWriter, errorCode int, message string, httpStatus int) {
	errCode := APIError{
		Code:    errorCode,
		Message: message,
	}

	inrec, _ := json.MarshalIndent(errCode, "", "  ")
	rlog.Error(errCode.Message)
	http.Error(w, string(inrec), httpStatus)
}

func (api *API) getStatus(w http.ResponseWriter, req *http.Request) {
	decoded := context.Get(req, "decoded")
	var auth duser.UserAccess
	mapstructure.Decode(decoded.(duser.UserAccess), &auth)

	var leds []dl.Led
	var sensors []ds.Sensor
	var blinds []dblind.Blind
	var hvacs []dhvac.Hvac
	var wagos []dwago.Wago
	var nanos []dnanosense.Nanosense
	var grID *int
	var isConfig *bool
	driverType := req.FormValue("type")
	if driverType == "" {
		driverType = FilterTypeAll
	}
	driverType = strings.ToLower(driverType)

	groupID := req.FormValue("groupID")
	if groupID != "" {
		i, err := strconv.Atoi(groupID)
		if err == nil {
			grID = &i
		}
	}

	isConfigured := req.FormValue("isConfigured")
	if isConfigured != "" {
		b, err := strconv.ParseBool(isConfigured)
		if err == nil {
			isConfig = &b
		}
	}

	if driverType == FilterTypeAll || driverType == FilterTypeLed {
		lights := database.GetLedsStatus(api.db)
		for _, led := range lights {
			if grID == nil || *grID == led.Group {
				if isConfig == nil || *isConfig == led.IsConfigured {
					if auth.Priviledge == duser.PriviledgeUser {
						if !tools.IntInSlice(led.Group, auth.AccessGroups) {
							continue
						}
					}
					leds = append(leds, led)
				}
			}
		}
	}

	if driverType == FilterTypeAll || driverType == FilterTypeSensor {
		cells := database.GetSensorsStatus(api.db)
		for _, sensor := range cells {
			if grID == nil || *grID == sensor.Group {
				if isConfig == nil || *isConfig == sensor.IsConfigured {
					if auth.Priviledge == duser.PriviledgeUser {
						if !tools.IntInSlice(sensor.Group, auth.AccessGroups) {
							continue
						}
					}
					sensors = append(sensors, sensor)
				}
			}
		}
	}

	if driverType == FilterTypeAll || driverType == FilterTypeBlind {
		drivers := database.GetBlindsStatus(api.db)
		for _, driver := range drivers {
			if grID == nil || *grID == driver.Group {
				if isConfig == nil || *isConfig == driver.IsConfigured {
					if auth.Priviledge == duser.PriviledgeUser {
						if !tools.IntInSlice(driver.Group, auth.AccessGroups) {
							continue
						}
					}
					blinds = append(blinds, driver)
				}
			}
		}
	}

	if driverType == FilterTypeAll || driverType == FilterTypeHvac {
		drivers := database.GetHvacsStatus(api.db)
		for _, driver := range drivers {
			if grID == nil || *grID == driver.Group {
				if isConfig == nil || *isConfig == driver.IsConfigured {
					if auth.Priviledge == duser.PriviledgeUser {
						if !tools.IntInSlice(driver.Group, auth.AccessGroups) {
							continue
						}
					}
					hvacs = append(hvacs, driver)
				}
			}
		}
	}

	if driverType == FilterTypeAll || driverType == FilterTypeWago {
		drivers := database.GetWagosStatus(api.db)
		for _, driver := range drivers {
			if isConfig == nil || *isConfig == driver.IsConfigured {
				if auth.Priviledge != duser.PriviledgeUser {
					continue
				}
				wagos = append(wagos, driver)
			}
		}
	}

	if driverType == FilterTypeAll || driverType == FilterTypeNano {
		drivers := database.GetNanosStatus(api.db)
		for _, driver := range drivers {
			if auth.Priviledge != duser.PriviledgeUser {
				if !tools.IntInSlice(driver.Group, auth.AccessGroups) {
					continue
				}

				nanos = append(nanos, driver)
			}
		}
	}

	status := Status{
		Leds:       leds,
		Sensors:    sensors,
		Blinds:     blinds,
		Hvacs:      hvacs,
		Wagos:      wagos,
		Nanosenses: nanos,
	}

	inrec, _ := json.MarshalIndent(status, "", "  ")
	w.Write(inrec)
}

func (api *API) getDump(w http.ResponseWriter, req *http.Request) {
	decoded := context.Get(req, "decoded")
	var auth duser.UserAccess
	mapstructure.Decode(decoded.(duser.UserAccess), &auth)

	var leds []DumpLed
	var sensors []DumpSensor
	var switchs []DumpSwitch
	var frames []DumpFrame
	var blinds []DumpBlind
	var hvacs []DumpHvac
	var wagos []DumpWago
	var groups []DumpGroup
	var nanos []DumpNanosense
	macs := make(map[string]bool)
	driversMac := make(map[string]bool)
	labels := make(map[string]bool)
	filterByMac := false
	MacsParam := req.FormValue("macs")
	if MacsParam != "" {
		tempMac := strings.Split(MacsParam, ",")

		for _, v := range tempMac {
			macs[strings.ToUpper(v)] = true
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
	frameElts := database.GetFramesDumpByLabel(api.db)
	frameEltsConfig := database.GetFramesConfigByLabel(api.db)
	nans := database.GetNanosStatusByLabel(api.db)
	nanosConfig := database.GetNanosConfigByLabel(api.db)

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
		case pconst.LED:
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
			if auth.Priviledge == duser.PriviledgeUser {
				if !tools.IntInSlice(gr, auth.AccessGroups) {
					continue
				}
			}
			dump.Ifc = ifc

			leds = append(leds, dump)
		case pconst.SENSOR:
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
			if auth.Priviledge == duser.PriviledgeUser {
				if !tools.IntInSlice(gr, auth.AccessGroups) {
					continue
				}
			}
			dump.Ifc = ifc
			sensors = append(sensors, dump)
		case pconst.BLIND:
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
			if auth.Priviledge == duser.PriviledgeUser {
				if !tools.IntInSlice(gr, auth.AccessGroups) {
					continue
				}
			}
			dump.Ifc = ifc
			blinds = append(blinds, dump)
		case pconst.HVAC:
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
			if auth.Priviledge == duser.PriviledgeUser {
				if !tools.IntInSlice(gr, auth.AccessGroups) {
					continue
				}
			}
			dump.Ifc = ifc
			hvacs = append(hvacs, dump)
		case pconst.WAGO:
			dump := DumpWago{}
			wago, ok := wags[ifc.Label]
			if ok {
				dump.Status = wago
			}
			config, ok := wagosConfig[ifc.Label]
			if ok {
				dump.Config = config
			}
			if auth.Priviledge == duser.PriviledgeUser {
				continue
			}
			dump.Ifc = ifc
			wagos = append(wagos, dump)
		case pconst.SWITCH:
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
			if auth.Priviledge == duser.PriviledgeUser {
				continue
			}
			switchs = append(switchs, dump)
		case pconst.FRAME:
			dump := DumpFrame{}
			frameElt, ok := frameElts[ifc.Label]
			if ok {
				dump.Status = frameElt
			}
			config, ok := frameEltsConfig[ifc.Label]
			if ok {
				dump.Config = config
			}
			dump.Ifc = ifc
			if auth.Priviledge == duser.PriviledgeUser {
				continue
			}
			frames = append(frames, dump)
		case pconst.NANOSENSE:
			dump := DumpNanosense{}
			gr := 0
			nano, ok := nans[ifc.Label]
			if ok {
				dump.Status = nano
				gr = nano.Group
			}
			config, ok := nanosConfig[ifc.Label]
			if ok {
				dump.Config = config
				if gr == 0 {
					gr = config.Group
				}
			}
			if auth.Priviledge == duser.PriviledgeUser {
				if !tools.IntInSlice(gr, auth.AccessGroups) {
					continue
				}
			}
			dump.Ifc = ifc
			nanos = append(nanos, dump)
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
		Leds:       leds,
		Sensors:    sensors,
		Blinds:     blinds,
		Hvacs:      hvacs,
		Wagos:      wagos,
		Switchs:    switchs,
		Frames:     frames,
		Groups:     groups,
		Nanosenses: nanos,
	}

	inrec, _ := json.MarshalIndent(dump, "", "  ")
	w.Write(inrec)
}

func (api *API) getHistory(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	var leds []LedHist
	ledHistories := make(map[string]LedHist)
	ledDrivers := history.GetLedsHistory(api.historydb)
	for _, l := range ledDrivers {
		val, ok := ledHistories[l.Date]
		if !ok {
			ledHistories[l.Date] = LedHist{
				Energy: l.Energy,
				Power:  l.Power,
				Date:   l.Date,
			}
		} else {
			val.Energy += l.Energy
			val.Power += l.Power
			ledHistories[l.Date] = val
		}
	}
	for _, hist := range ledHistories {
		leds = append(leds, hist)
	}

	dump := GlobalHistory{
		Leds: leds,
	}

	inrec, _ := json.MarshalIndent(dump, "", "  ")
	w.Write(inrec)
}

func (api *API) setConfig(w http.ResponseWriter, req *http.Request) {
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	config := Conf{}
	err = json.Unmarshal([]byte(body), &config)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	event := make(map[string]interface{})
	for _, led := range config.Leds {
		event["led"] = led
	}
	for _, sensor := range config.Sensors {
		event["sensor"] = sensor
	}
	for _, group := range config.Groups {
		event["group"] = group
	}
	for _, sw := range config.Switchs {
		event["switch"] = sw
	}
	for _, sw := range config.Blinds {
		event["blind"] = sw
	}
	for _, sw := range config.Hvacs {
		event["hvac"] = sw
	}
	for _, sw := range config.Wagos {
		event["wago"] = sw
	}
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) webEvents(w http.ResponseWriter, r *http.Request) {
	ws, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		rlog.Error("Error when switching in websocket " + err.Error())
		return
	}
	decoded := context.Get(r, "decoded")
	var auth duser.UserAccess
	mapstructure.Decode(decoded.(duser.UserAccess), &auth)
	api.clients[ws] = auth

}

func (api *API) consumptionEvents(w http.ResponseWriter, r *http.Request) {
	ws, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		rlog.Error("Error when switching in consumption websocket " + err.Error())
		return
	}
	decoded := context.Get(r, "decoded")
	var auth duser.UserAccess
	mapstructure.Decode(decoded.(duser.UserAccess), &auth)
	api.clientsConso[ws] = auth
}

func (api *API) websocketEvents() {
	for {
		select {
		case event := <-api.eventsAPI:
			api.apiMutex.Lock()
			for client, auth := range api.clients {
				var res map[string]interface{}
				if auth.Priviledge == duser.PriviledgeUser {
					res = make(map[string]interface{})
					for evtType, e := range event {
						evt, _ := core.ToEventStatus(e)
						if evt == nil {
							continue
						}
						newEvt := core.EventStatus{
							Leds:    []core.EventLed{},
							Sensors: []core.EventSensor{},
							Groups:  []gm.GroupStatus{},
							Blinds:  []core.EventBlind{},
							Wagos:   []core.EventWago{},
							Nanos:   []core.EventNano{},
							Hvacs:   []core.EventHvac{},
							Switchs: []core.EventSwitch{},
						}

						for _, bld := range evt.Blinds {
							if !tools.IntInSlice(bld.Blind.Group, auth.AccessGroups) {
								continue
							}
							newEvt.Blinds = append(newEvt.Blinds, bld)
						}

						for _, led := range evt.Leds {
							if !tools.IntInSlice(led.Led.Group, auth.AccessGroups) {
								continue
							}
							newEvt.Leds = append(newEvt.Leds, led)
						}

						for _, hvac := range evt.Hvacs {
							if !tools.IntInSlice(hvac.Hvac.Group, auth.AccessGroups) {
								continue
							}
							newEvt.Hvacs = append(newEvt.Hvacs, hvac)
						}

						for _, sensor := range evt.Sensors {
							if !tools.IntInSlice(sensor.Sensor.Group, auth.AccessGroups) {
								continue
							}
							newEvt.Sensors = append(newEvt.Sensors, sensor)
						}

						for _, wago := range evt.Wagos {
							if !tools.StringInSlice(auth.Priviledge, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) {
								continue
							}
							newEvt.Wagos = append(newEvt.Wagos, wago)
						}

						for _, sw := range evt.Switchs {
							if !tools.StringInSlice(auth.Priviledge, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) {
								continue
							}
							newEvt.Switchs = append(newEvt.Switchs, sw)
						}

						for _, nano := range evt.Nanos {
							if !tools.IntInSlice(nano.Nano.Group, auth.AccessGroups) {
								continue
							}
							newEvt.Nanos = append(newEvt.Nanos, nano)
						}

						for _, group := range evt.Groups {
							if !tools.IntInSlice(group.Group, auth.AccessGroups) {
								continue
							}
							newEvt.Groups = append(newEvt.Groups, group)
						}

						res[evtType] = newEvt
					}
				} else {
					res = event
				}

				if err := client.WriteJSON(res); err != nil {
					rlog.Error("Error writing in websocket" + err.Error())
					client.Close()
					delete(api.clients, client)
				}
			}
			api.apiMutex.Unlock()
		}
	}
}

func (api *API) websocketConsumptions() {
	for {
		select {
		case event := <-api.eventsConso:
			api.apiMutex.Lock()
			for client := range api.clientsConso {
				if err := client.WriteJSON(event); err != nil {
					rlog.Error("Error writing in websocket" + err.Error())
					client.Close()
					delete(api.clientsConso, client)
				}
			}
			api.apiMutex.Unlock()
		}
	}
}

func (api *API) getAPIs(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w, req)
	versions := []string{"v1.0"}
	apiInfo := APIInfo{
		Versions: versions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *API) getV1Functions(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w, req)
	apiV1 := "/v1.0"
	functions := []string{apiV1 + "/setup/sensor", apiV1 + "/setup/led",
		apiV1 + "/setup/group", apiV1 + "/setup/switch", apiV1 + "/setup/installMode",
		apiV1 + "/setup/service", apiV1 + "/setup/blind", apiV1 + "/setup/hvac", apiV1 + "/setup/wago",
		apiV1 + "/config/led", apiV1 + "/config/sensor", apiV1 + "/config/blind", apiV1 + "/config/hvac",
		apiV1 + "/config/group", apiV1 + "/config/switch", apiV1 + "/config/wago", apiV1 + "/configs",
		apiV1 + "/status", apiV1 + "/events", apiV1 + "/events/consumption", apiV1 + "/history",
		apiV1 + "/command/led", apiV1 + "/command/blind", apiV1 + "/command/hvac", apiV1 + "/command/group", apiV1 + "/project/ifcInfo",
		apiV1 + "/project/model", apiV1 + "/project/bim", apiV1 + "/project", apiV1 + "/dump",
		apiV1 + "/status/sensor", apiV1 + "/status/group", apiV1 + "/status/led", apiV1 + "/status/blind", apiV1 + "/status/hvac",
		apiV1 + "/status/groups", apiV1 + "/status/wago", apiV1 + "/maintenance/driver", apiV1 + "/commissioning/install",
		apiV1 + "/install/status", apiV1 + "/install/stickers",
		apiV1 + "/user/info", apiV1 + "/user/login", apiV1 + "/map/upload", apiV1 + "/map/upload/status",
	}
	apiInfo := APIFunctions{
		Functions: functions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *API) getFunctions(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w, req)
	functions := []string{"/versions"}
	apiInfo := APIFunctions{
		Functions: functions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *API) hasAccessMode(w http.ResponseWriter, req *http.Request, modes []string) error {
	decoded := context.Get(req, "decoded")
	var auth duser.UserAccess
	mapstructure.Decode(decoded.(duser.UserAccess), &auth)
	if !tools.StringInSlice(auth.Priviledge, modes) {
		return NewError("Unauthorized Access")
	}
	return nil
}

func (api *API) hasEnoughRight(w http.ResponseWriter, req *http.Request, group int) error {
	decoded := context.Get(req, "decoded")
	var auth duser.UserAccess
	mapstructure.Decode(decoded.(duser.UserAccess), &auth)
	if auth.Priviledge == duser.PriviledgeUser {
		if !tools.IntInSlice(group, auth.AccessGroups) {
			return NewError("Unauthorized Access")
		}
		return nil
	}
	if !tools.StringInSlice(auth.Priviledge, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) {
		return NewError("Unauthorized Access")
	}
	return nil
}

func (api *API) swagger() {
	go api.websocketConsumptions()
	go api.websocketEvents()
	router := mux.NewRouter()
	sh := http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("/media/userdata/www/swaggerui/")))
	router.PathPrefix("/swaggerui/").Handler(sh)

	// API v1.0
	apiV1 := "/v1.0"
	router.HandleFunc(apiV1+"/functions", api.getV1Functions).Methods("GET")

	// Auth
	router.HandleFunc(apiV1+"/user/login", api.createToken).Methods("POST")
	router.HandleFunc(apiV1+"/user/info", api.verification(api.getUserInfo)).Methods("GET")
	router.HandleFunc(apiV1+"/user/logout", api.verification(api.logout)).Methods("POST")

	//setup API
	router.HandleFunc(apiV1+"/setup/sensor/{mac}", api.verification(api.getSensorSetup)).Methods("GET")
	router.HandleFunc(apiV1+"/setup/sensor/{mac}", api.verification(api.removeSensorSetup)).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/sensor", api.verification(api.setSensorSetup)).Methods("POST")
	router.HandleFunc(apiV1+"/setup/led/{mac}", api.verification(api.getLedSetup)).Methods("GET")
	router.HandleFunc(apiV1+"/setup/led/{mac}", api.verification(api.removeLedSetup)).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/led", api.verification(api.setLedSetup)).Methods("POST")
	router.HandleFunc(apiV1+"/setup/blind/{mac}", api.verification(api.getBlindSetup)).Methods("GET")
	router.HandleFunc(apiV1+"/setup/blind/{mac}", api.verification(api.removeBlindSetup)).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/blind", api.verification(api.setBlindSetup)).Methods("POST")
	router.HandleFunc(apiV1+"/setup/hvac/{mac}", api.verification(api.getHvacSetup)).Methods("GET")
	router.HandleFunc(apiV1+"/setup/hvac/{mac}", api.verification(api.removeHvacSetup)).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/hvac", api.verification(api.setHvacSetup)).Methods("POST")
	router.HandleFunc(apiV1+"/setup/wago/{mac}", api.verification(api.getWagoSetup)).Methods("GET")
	router.HandleFunc(apiV1+"/setup/wago/{mac}", api.verification(api.removeWagoSetup)).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/wago", api.verification(api.setWagoSetup)).Methods("POST")
	router.HandleFunc(apiV1+"/setup/group/{groupID}", api.verification(api.getGroupSetup)).Methods("GET")
	router.HandleFunc(apiV1+"/setup/group/{groupID}", api.verification(api.removeGroupSetup)).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/group", api.verification(api.setGroupSetup)).Methods("POST")
	router.HandleFunc(apiV1+"/setup/switch/{mac}", api.verification(api.getSwitchSetup)).Methods("GET")
	router.HandleFunc(apiV1+"/setup/switch/{mac}", api.verification(api.removeSwitchSetup)).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/switch", api.verification(api.setSwitchSetup)).Methods("POST")
	router.HandleFunc(apiV1+"/setup/service/{name}", api.verification(api.getServiceSetup)).Methods("GET")
	router.HandleFunc(apiV1+"/setup/service/{name}", api.verification(api.removeServiceSetup)).Methods("DELETE")
	router.HandleFunc(apiV1+"/setup/service", api.verification(api.setServiceSetup)).Methods("POST")

	//config API
	router.HandleFunc(apiV1+"/config/led", api.verification(api.setLedConfig)).Methods("POST")
	router.HandleFunc(apiV1+"/config/sensor", api.verification(api.setSensorConfig)).Methods("POST")
	router.HandleFunc(apiV1+"/config/blind", api.verification(api.setBlindConfig)).Methods("POST")
	router.HandleFunc(apiV1+"/config/hvac", api.verification(api.setHvacConfig)).Methods("POST")
	router.HandleFunc(apiV1+"/config/wago", api.verification(api.setWagoConfig)).Methods("POST")
	router.HandleFunc(apiV1+"/config/nanosense", api.verification(api.setNanoConfig)).Methods("POST")
	router.HandleFunc(apiV1+"/config/group", api.verification(api.setGroupConfig)).Methods("POST")
	router.HandleFunc(apiV1+"/config/switch", api.verification(api.setSwitchConfig)).Methods("POST")
	router.HandleFunc(apiV1+"/configs", api.verification(api.setConfig)).Methods("POST")

	//status API
	router.HandleFunc(apiV1+"/status/sensor/{mac}", api.verification(api.getSensorStatus)).Methods("GET")
	router.HandleFunc(apiV1+"/status/blind/{mac}", api.verification(api.getBlindStatus)).Methods("GET")
	router.HandleFunc(apiV1+"/status/hvac/{mac}", api.verification(api.getHvacStatus)).Methods("GET")
	router.HandleFunc(apiV1+"/status/led/{mac}", api.verification(api.getLedStatus)).Methods("GET")
	router.HandleFunc(apiV1+"/status/wago/{mac}", api.verification(api.getWagoStatus)).Methods("GET")
	router.HandleFunc(apiV1+"/status/group/{groupID}", api.verification(api.getGroupStatus)).Methods("GET")
	router.HandleFunc(apiV1+"/status/groups", api.verification(api.getGroupsStatus)).Methods("GET")
	router.HandleFunc(apiV1+"/status", api.verification(api.getStatus)).Methods("GET")

	//events API
	router.HandleFunc(apiV1+"/events", api.verification(api.webEvents))
	router.HandleFunc(apiV1+"/events/consumption", api.verification(api.consumptionEvents))

	//command API
	router.HandleFunc(apiV1+"/command/led", api.verification(api.sendLedCommand)).Methods("POST")
	router.HandleFunc(apiV1+"/command/blind", api.verification(api.sendBlindCommand)).Methods("POST")
	router.HandleFunc(apiV1+"/command/hvac", api.verification(api.sendHvacCommand)).Methods("POST")
	router.HandleFunc(apiV1+"/command/group", api.verification(api.sendGroupCommand)).Methods("POST")

	//project API
	router.HandleFunc(apiV1+"/project/ifcInfo/{label}", api.verification(api.getIfcInfo)).Methods("GET")
	router.HandleFunc(apiV1+"/project/ifcInfo/{label}", api.verification(api.removeIfcInfo)).Methods("DELETE")
	router.HandleFunc(apiV1+"/project/ifcInfo", api.verification(api.setIfcInfo)).Methods("POST")
	router.HandleFunc(apiV1+"/project/model/{modelName}", api.verification(api.getModelInfo)).Methods("GET")
	router.HandleFunc(apiV1+"/project/model/{modelName}", api.verification(api.removeModelInfo)).Methods("DELETE")
	router.HandleFunc(apiV1+"/project/model", api.verification(api.setModelInfo)).Methods("POST")
	router.HandleFunc(apiV1+"/project/bim/{label}", api.verification(api.getBim)).Methods("GET")
	router.HandleFunc(apiV1+"/project/bim/{label}", api.verification(api.removeBim)).Methods("DELETE")
	router.HandleFunc(apiV1+"/project/bim", api.verification(api.setBim)).Methods("POST")
	router.HandleFunc(apiV1+"/project", api.verification(api.getIfc)).Methods("GET")

	//map API
	router.HandleFunc(apiV1+"/map/upload", api.verification(api.uploadHandler)).Methods("POST")
	router.HandleFunc(apiV1+"/map/upload/status", api.verification(api.uploadStatus)).Methods("GET")

	//Maintenance API
	router.HandleFunc(apiV1+"/maintenance/driver", api.verification(api.replaceDriver)).Methods("POST")
	router.HandleFunc(apiV1+"/install/status", api.verification(api.installStatus)).Methods("GET")
	router.HandleFunc(apiV1+"/install/stickers", api.verification(api.qrcodeGeneration)).Methods("GET")
	router.HandleFunc(apiV1+"/maintenance/exportDB", api.verification(api.exportDBStart)).Methods("GET")
	router.HandleFunc(apiV1+"/maintenance/exportDB/result", api.verification(api.exportDB)).Methods("GET")

	//Install API
	router.HandleFunc(apiV1+"/commissioning/install", api.verification(api.installDriver)).Methods("POST")

	//dump API
	router.HandleFunc(apiV1+"/dump", api.verification(api.getDump)).Methods("GET")

	//History API
	router.HandleFunc(apiV1+"/history", api.verification(api.getHistory)).Methods("GET")

	//unversionned API
	router.HandleFunc("/versions", api.getAPIs).Methods("GET")
	router.HandleFunc("/functions", api.getFunctions).Methods("GET")

	if api.browsingFolder != "" {
		sh2 := http.StripPrefix("/", http.FileServer(http.Dir(api.browsingFolder)))
		router.PathPrefix("/").Handler(sh2)
	}

	log.Fatal(http.ListenAndServeTLS(api.apiIP+":"+api.apiPort, api.certificate, api.keyfile, router))
}
