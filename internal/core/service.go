package core

import (
	pkg "github.com/energieip/common-components-go/pkg/service"
)

//ServiceDump content
type ServiceDump struct {
	pkg.ServiceStatus
	SwitchMac string `json:"switchMac"`
}
