package network

import (
	"encoding/json"
	"time"

	genericNetwork "github.com/energieip/common-network-go/pkg/network"
	pkg "github.com/energieip/common-service-go/pkg/service"
	sd "github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/romana/rlog"
)

const (
	EventHello = "switchHello"
	EventDump  = "switchDump"

	EventWriteCfg = "switchWriteCfg"
)

//ServerNetwork network object
type ServerNetwork struct {
	Iface     genericNetwork.NetworkInterface
	Events    chan map[string]sd.SwitchStatus
	EventsCfg chan map[string]core.ServerConfig
}

//CreateServerNetwork create network server object
func CreateServerNetwork() (*ServerNetwork, error) {
	serverBroker, err := genericNetwork.NewNetwork(genericNetwork.MQTT)
	if err != nil {
		return nil, err
	}
	serverNet := ServerNetwork{
		Iface:     serverBroker,
		Events:    make(chan map[string]sd.SwitchStatus),
		EventsCfg: make(chan map[string]core.ServerConfig),
	}
	return &serverNet, nil

}

//LocalConnection connect service to server broker
func (net ServerNetwork) LocalConnection(conf pkg.ServiceConfig, clientID string) error {
	cbkServer := make(map[string]func(genericNetwork.Client, genericNetwork.Message))
	cbkServer["/read/switch/+/setup/hello"] = net.onHello
	cbkServer["/read/switch/+/status/dump"] = net.onDump
	cbkServer["/write/server/components/config"] = net.registerConfigs

	confServer := genericNetwork.NetworkConfig{
		IP:               conf.NetworkBroker.IP,
		Port:             conf.NetworkBroker.Port,
		ClientName:       clientID,
		Callbacks:        cbkServer,
		LogLevel:         conf.LogLevel,
		User:             conf.NetworkBroker.Login,
		Password:         conf.NetworkBroker.Password,
		ClientKey:        conf.NetworkBroker.KeyPath,
		ServerCertificat: conf.NetworkBroker.CaPath,
	}

	for {
		rlog.Info("Try to connect to " + conf.NetworkBroker.IP)
		err := net.Iface.Initialize(confServer)
		if err == nil {
			rlog.Info(clientID + " connected to server broker " + conf.NetworkBroker.IP)
			return err
		}
		timer := time.NewTicker(time.Second)
		rlog.Error("Cannot connect to broker " + conf.NetworkBroker.IP + " error: " + err.Error())
		rlog.Error("Try to reconnect " + conf.NetworkBroker.IP + " in 1s")

		select {
		case <-timer.C:
			continue
		}
	}
}

func (net ServerNetwork) onHello(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Debug("Received switch Hello: Received topic: " + msg.Topic() + " payload: " + string(payload))
	var switchStatus sd.SwitchStatus
	err := json.Unmarshal(payload, &switchStatus)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}

	event := make(map[string]sd.SwitchStatus)
	event[EventHello] = switchStatus
	net.Events <- event
}

func (net ServerNetwork) onDump(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Info("Received switch Dump: Received topic: " + msg.Topic() + " payload: " + string(payload))
	var switchStatus sd.SwitchStatus
	err := json.Unmarshal(payload, &switchStatus)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}

	event := make(map[string]sd.SwitchStatus)
	event[EventDump] = switchStatus
	net.Events <- event
}

func (net ServerNetwork) registerConfigs(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Debug("Received registerConfigs: Received topic: " + msg.Topic() + " payload: " + string(payload))
	var servCfg core.ServerConfig
	err := json.Unmarshal(payload, &servCfg)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}

	event := make(map[string]core.ServerConfig)
	event[EventWriteCfg] = servCfg
	net.EventsCfg <- event
}

//Disconnect from server
func (net ServerNetwork) Disconnect() {
	net.Iface.Disconnect()
}

//SendCommand to server
func (net ServerNetwork) SendCommand(topic, content string) error {
	return net.Iface.SendCommand(topic, content)
}
