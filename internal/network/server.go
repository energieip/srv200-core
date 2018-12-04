package network

import (
	"encoding/json"
	"strconv"
	"time"

	genericNetwork "github.com/energieip/common-network-go/pkg/network"
	pkg "github.com/energieip/common-service-go/pkg/service"
	"github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/romana/rlog"
)

const (
	EventHello = "switchHello"
	EventDump  = "switchDump"

	EventRemoveCfg = "switchRemoveCfg"
	EventWriteCfg  = "switchWriteCfg"
	EventManualCfg = "switchManualCfg"
)

//ServerNetwork network object
type ServerNetwork struct {
	Iface             genericNetwork.NetworkInterface
	Events            chan map[string]deviceswitch.SwitchStatus
	EventsCfg         chan map[string]core.ServerConfig
	EventsInstallMode chan bool //installation mode
	EventsCmd         chan core.ServerCmd
}

//CreateServerNetwork create network server object
func CreateServerNetwork() (*ServerNetwork, error) {
	serverBroker, err := genericNetwork.NewNetwork(genericNetwork.MQTT)
	if err != nil {
		return nil, err
	}
	serverNet := ServerNetwork{
		Iface:             serverBroker,
		Events:            make(chan map[string]deviceswitch.SwitchStatus),
		EventsCfg:         make(chan map[string]core.ServerConfig),
		EventsInstallMode: make(chan bool),
		EventsCmd:         make(chan core.ServerCmd),
	}
	return &serverNet, nil

}

//LocalConnection connect service to server broker
func (net ServerNetwork) LocalConnection(conf pkg.ServiceConfig, clientID string) error {
	cbkServer := make(map[string]func(genericNetwork.Client, genericNetwork.Message))
	cbkServer["/read/switch/+/setup/hello"] = net.onHello
	cbkServer["/read/switch/+/status/dump"] = net.onDump
	cbkServer["/write/server/components/config"] = net.registerConfigs
	cbkServer["/remove/server/components/config"] = net.removeConfigs
	cbkServer["/write/server/manual/control"] = net.manualControl
	cbkServer["/write/server/install/mode"] = net.installMode

	confServer := genericNetwork.NetworkConfig{
		IP:         conf.NetworkBroker.IP,
		Port:       conf.NetworkBroker.Port,
		ClientName: clientID,
		Callbacks:  cbkServer,
		LogLevel:   conf.LogLevel,
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
	var switchStatus deviceswitch.SwitchStatus
	err := json.Unmarshal(payload, &switchStatus)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}

	event := make(map[string]deviceswitch.SwitchStatus)
	event[EventHello] = switchStatus
	net.Events <- event
}

func (net ServerNetwork) onDump(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Debug("Received switch Dump: Received topic: " + msg.Topic() + " payload: " + string(payload))
	var switchStatus deviceswitch.SwitchStatus
	err := json.Unmarshal(payload, &switchStatus)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}

	event := make(map[string]deviceswitch.SwitchStatus)
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

func (net ServerNetwork) removeConfigs(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Debug("Received removeConfigs: Received topic: " + msg.Topic() + " payload: " + string(payload))
}

func (net ServerNetwork) manualControl(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Debug("Received manualControl: Received topic: " + msg.Topic() + " payload: " + string(payload))
	var servCmd core.ServerCmd
	err := json.Unmarshal(payload, &servCmd)
	if err != nil {
		rlog.Error("Cannot parse command ", err.Error())
		return
	}
	net.EventsCmd <- servCmd
}

func (net ServerNetwork) installMode(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := string(msg.Payload())
	rlog.Info("Received installMode: Received topic: " + msg.Topic() + " payload: " + payload)
	mode, err := strconv.ParseBool(payload)
	if err != nil {
		rlog.Error("Cannot parse installation Mode ", err.Error())
		return
	}
	net.EventsInstallMode <- mode
}

//Disconnect from server
func (net ServerNetwork) Disconnect() {
	net.Iface.Disconnect()
}

//SendCommand to server
func (net ServerNetwork) SendCommand(topic, content string) error {
	return net.Iface.SendCommand(topic, content)
}
