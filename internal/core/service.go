package core

import (
	"encoding/json"

	pkg "github.com/energieip/common-service-go/pkg/service"
)

//Service describe the link between the object in the building map and the configuration
type Service struct {
	Name               string   `json:"name"`
	Systemd            []string `json:"systemd"` //systemd service
	Version            string   `json:"version"`
	PackageName        string   `json:"packageName"`        //DebianPackageName
	PersistentDataPath string   `json:"persistentDataPath"` // link to store persistent data
	ConfigPath         string   `json:"configPath"`
}

//ServiceDump content
type ServiceDump struct {
	pkg.ServiceStatus
	SwitchMac string `json:"switchMac"`
}

// ToJSON dump ToService struct
func (p Service) ToJSON() (string, error) {
	inrec, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(inrec[:]), err
}

//ToService convert map interface to ToService object
func ToService(val interface{}) (*Service, error) {
	var p Service
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &p)
	return &p, err
}
