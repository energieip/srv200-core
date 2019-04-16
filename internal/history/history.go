package history

import (
	"encoding/json"
	"time"

	"github.com/energieip/common-components-go/pkg/database"
	"github.com/energieip/common-components-go/pkg/dblind"
	"github.com/energieip/common-components-go/pkg/dhvac"
	dl "github.com/energieip/common-components-go/pkg/dled"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/romana/rlog"
)

const (
	HistoryDB = "history"

	LedsTable    = "leds"
	BlindsTable  = "blinds"
	SwitchsTable = "switchs"
	HvacsTable   = "hvacs"
	TdTable      = "tds"
)

type databaseError struct {
	s string
}

func (e *databaseError) Error() string {
	return e.s
}

// NewError raise an error
func NewError(text string) error {
	return &databaseError{text}
}

type HistoryDb = database.DatabaseInterface

//ConnectDatabase plug datbase
func ConnectDatabase(ip, port string) (*HistoryDb, error) {
	db, err := database.NewDatabase(database.RETHINKDB)
	if err != nil {
		rlog.Error("database err " + err.Error())
		return nil, err
	}

	confDb := database.DatabaseConfig{
		IP:   ip,
		Port: port,
	}
	err = db.Initialize(confDb)
	if err != nil {
		rlog.Error("Cannot connect to database " + err.Error())
		return nil, err
	}

	for _, dbName := range []string{HistoryDB} {
		err = db.CreateDB(dbName)
		if err != nil {
			rlog.Warn("Create DB ", err.Error())
		}

		tableCfg := make(map[string]interface{})
		tableCfg[LedsTable] = dl.Led{}
		tableCfg[SwitchsTable] = core.SwitchDump{}
		tableCfg[BlindsTable] = dblind.Blind{}
		tableCfg[HvacsTable] = dhvac.Hvac{}

		for tableName, objs := range tableCfg {
			err = db.CreateTable(dbName, tableName, &objs)
			if err != nil {
				rlog.Warn("Create table ", err.Error())
			}
		}
	}
	return &db, nil
}

type LedHistory struct {
	Mac    string  `json:"mac"`
	Energy float64 `json:"energy"`
	Power  int     `json:"power"`
	Date   string  `json:"date"`
	Group  int     `json:"group"`
}

type BlindHistory struct {
	Mac    string  `json:"mac"`
	Energy float64 `json:"energy"`
	Power  int     `json:"power"`
	Date   string  `json:"date"`
	Group  int     `json:"group"`
}

type SwitchHistory struct {
	Mac     string  `json:"mac"`
	Energy  float64 `json:"energy"`
	Power   int     `json:"power"`
	Date    string  `json:"date"`
	Cluster int     `json:"cluster"`
}

//ToLedHistory convert map interface to Led object
func ToLedHistory(val interface{}) (*LedHistory, error) {
	var driver LedHistory
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &driver)
	return &driver, err
}

//ToBlindHistory convert map interface to blind object
func ToBlindHistory(val interface{}) (*BlindHistory, error) {
	var driver BlindHistory
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &driver)
	return &driver, err
}

func SaveHistory(db HistoryDb, dbName, tbName string, obj interface{}) error {
	_, err := db.InsertRecord(dbName, tbName, obj)
	return err
}

func SaveLedHistory(db HistoryDb, driver dl.Led) error {
	led := LedHistory{
		Mac:    driver.Mac,
		Energy: driver.Energy,
		Power:  driver.LinePower,
		Group:  driver.Group,
		Date:   time.Now().Format(time.RFC850),
	}
	return SaveHistory(db, HistoryDB, LedsTable, led)
}

func SaveBlindHistory(db HistoryDb, driver dblind.Blind) error {
	blind := BlindHistory{
		Mac: driver.Mac,
		// Energy: driver.Energy,
		Power: driver.LinePower,
		Group: driver.Group,
		Date:  time.Now().Format(time.RFC850),
	}
	return SaveHistory(db, HistoryDB, BlindsTable, blind)
}

func GetBlindsHistory(db HistoryDb) []BlindHistory {
	var history []BlindHistory
	stored, err := db.FetchAllRecords(HistoryDB, BlindsTable)
	if err != nil || stored == nil {
		return history
	}
	for _, l := range stored {
		driver, err := ToBlindHistory(l)
		if err != nil || driver == nil {
			continue
		}
		history = append(history, *driver)
	}
	return history
}

func GetLedsHistory(db HistoryDb) []LedHistory {
	var history []LedHistory
	stored, err := db.FetchAllRecords(HistoryDB, LedsTable)
	if err != nil || stored == nil {
		return history
	}
	for _, l := range stored {
		driver, err := ToLedHistory(l)
		if err != nil || driver == nil {
			continue
		}
		history = append(history, *driver)
	}
	return history
}
