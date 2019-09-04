package api

import (
	"sync"

	jwt "github.com/dgrijalva/jwt-go"
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
	importDBStatus  string
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
