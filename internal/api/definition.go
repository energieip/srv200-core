package api

import (
	"sync"

	"github.com/energieip/common-components-go/pkg/dnanosense"
	"github.com/energieip/common-components-go/pkg/dwago"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/energieip/common-components-go/pkg/dblind"
	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/dhvac"
	dl "github.com/energieip/common-components-go/pkg/dled"
	ds "github.com/energieip/common-components-go/pkg/dsensor"
	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/energieip/srv200-coreservice-go/internal/history"
	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map"
)

const (
	APIErrorDeviceNotFound = 1
	APIErrorBodyParsing    = 2
	APIErrorDatabase       = 3
	APIErrorInvalidValue   = 4
	APIErrorUnauthorized   = 5
	APIErrorExpiredToken   = 6

	TokenName = "EiPAccessToken"

	TokenExpirationTime = 86400 // in seconds: 1day

	FilterTypeAll    = "all"
	FilterTypeSensor = "sensor"
	FilterTypeLed    = "led"
	FilterTypeBlind  = "blind"
	FilterTypeHvac   = "hvac"
	FilterTypeWago   = "wago"
	FilterTypeNano   = "nanosense"
)

//APIError Message error code
type APIError struct {
	Code    int    `json:"code"` //errorCode
	Message string `json:"message"`
}

type API struct {
	clients         map[*websocket.Conn]duser.UserAccess
	clientsConso    map[*websocket.Conn]duser.UserAccess
	upgrader        websocket.Upgrader
	db              database.Database
	historydb       history.HistoryDb
	eventsAPI       chan map[string]interface{}
	eventsConso     chan core.EventConsumption
	EventsToBackend chan map[string]interface{}
	access          cmap.ConcurrentMap
	apiMutex        sync.Mutex
	certificate     string
	keyfile         string
	apiIP           string
	apiPort         string
	apiPassword     string
	browsingFolder  string
	dataPath        string
	uploadValue     *string
	exportDBPath    string
	exportDBStatus  string
}

type JwtToken struct {
	Token     string `json:"accessToken"`
	TokenType string `json:"tokenType"`
	ExpireIn  int    `json:"expireIn"`
}

type Claims struct {
	jwt.StandardClaims
}

type Credentials struct {
	UserKey string `json:"userKey"`
}

//Status
type Status struct {
	Leds       []dl.Led               `json:"leds"`
	Sensors    []ds.Sensor            `json:"sensors"`
	Blinds     []dblind.Blind         `json:"blinds"`
	Hvacs      []dhvac.Hvac           `json:"hvacs"`
	Wagos      []dwago.Wago           `json:"wagos"`
	Nanosenses []dnanosense.Nanosense `json:"nanosenses"`
}

//DumpBlind
type DumpBlind struct {
	Ifc    core.IfcInfo      `json:"ifc"`
	Status dblind.Blind      `json:"status"`
	Config dblind.BlindSetup `json:"config"`
}

//DumpHvac
type DumpHvac struct {
	Ifc    core.IfcInfo    `json:"ifc"`
	Status dhvac.Hvac      `json:"status"`
	Config dhvac.HvacSetup `json:"config"`
}

//DumpLed
type DumpLed struct {
	Ifc    core.IfcInfo `json:"ifc"`
	Status dl.Led       `json:"status"`
	Config dl.LedSetup  `json:"config"`
}

//DumpSensor
type DumpSensor struct {
	Ifc    core.IfcInfo   `json:"ifc"`
	Status ds.Sensor      `json:"status"`
	Config ds.SensorSetup `json:"config"`
}

//DumpSwitch
type DumpSwitch struct {
	Ifc    core.IfcInfo      `json:"ifc"`
	Status core.SwitchDump   `json:"status"`
	Config core.SwitchConfig `json:"config"`
}

//DumpFrame
type DumpFrame struct {
	Ifc    core.IfcInfo     `json:"ifc"`
	Config core.Frame       `json:"config"`
	Status core.FrameStatus `json:"status"`
}

//DumpWago
type DumpWago struct {
	Ifc    core.IfcInfo    `json:"ifc"`
	Status dwago.Wago      `json:"status"`
	Config dwago.WagoSetup `json:"config"`
}

//DumpNano
type DumpNanosense struct {
	Ifc    core.IfcInfo              `json:"ifc"`
	Status dnanosense.Nanosense      `json:"status"`
	Config dnanosense.NanosenseSetup `json:"config"`
}

//DumpSwitch
type DumpGroup struct {
	Status gm.GroupStatus `json:"status"`
	Config gm.GroupConfig `json:"config"`
}

//Dump
type Dump struct {
	Leds       []DumpLed       `json:"leds"`
	Sensors    []DumpSensor    `json:"sensors"`
	Blinds     []DumpBlind     `json:"blinds"`
	Hvacs      []DumpHvac      `json:"hvacs"`
	Wagos      []DumpWago      `json:"wagos"`
	Switchs    []DumpSwitch    `json:"switchs"`
	Groups     []DumpGroup     `json:"groups"`
	Frames     []DumpFrame     `json:"frames"`
	Nanosenses []DumpNanosense `json:"nanos"`
}

//UserAuthorization
type UserAuthorization struct {
	Priviledges []string `json:"priviledges"`
	AccessGroup []int    `json:"accessGroups"`
}

type LedHist struct {
	Energy float64 `json:"energy"`
	Power  int     `json:"power"`
	Date   string  `json:"date"`
}

type GlobalHistory struct {
	Leds []LedHist `json:"leds"`
}

type APIInfo struct {
	Versions []string `json:"versions"`
}

type APIFunctions struct {
	Functions []string `json:"functions"`
}

type Conf struct {
	Leds    []dl.LedConf        `json:"leds"`
	Sensors []ds.SensorConf     `json:"sensors"`
	Blinds  []dblind.BlindConf  `json:"blinds"`
	Hvacs   []dhvac.HvacConf    `json:"hvacs"`
	Wagos   []dwago.WagoConf    `json:"wagos"`
	Groups  []gm.GroupConfig    `json:"groups"`
	Switchs []core.SwitchConfig `json:"switchs"`
}

type networkError struct {
	s string
}

func (e *networkError) Error() string {
	return e.s
}

// NewError raise an error
func NewError(text string) error {
	return &networkError{text}
}
