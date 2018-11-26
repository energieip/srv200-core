package network

import (
	"encoding/json"
	"time"

	genericNetwork "github.com/energieip/common-network-go/pkg/network"
	"github.com/energieip/common-switch-go/pkg/deviceswitch"
	"github.com/energieip/srv200-coreservice-go/pkg/config"
	"github.com/romana/rlog"
)

const (
	EventHello = "switchHello"
	EventDump  = "switchDump"
)

//ServerNetwork network object
type ServerNetwork struct {
	Iface  genericNetwork.NetworkInterface
	Events chan map[string]deviceswitch.SwitchStatus
}

//CreateServerNetwork create network server object
func CreateServerNetwork() (*ServerNetwork, error) {
	serverBroker, err := genericNetwork.NewNetwork(genericNetwork.MQTT)
	if err != nil {
		return nil, err
	}
	serverNet := ServerNetwork{
		Iface:  serverBroker,
		Events: make(chan map[string]deviceswitch.SwitchStatus),
	}
	return &serverNet, nil

}

//LocalConnection connect service to server broker
func (net ServerNetwork) LocalConnection(conf config.Configuration, clientID string) error {
	cbkServer := make(map[string]func(genericNetwork.Client, genericNetwork.Message))
	cbkServer["/read/switch/+/setup/hello"] = net.onHello
	cbkServer["/read/switch/+/status/dump"] = net.onDump
	cbkServer["/write/server/components/config"] = net.registerConfigs
	cbkServer["/remove/server/components/config"] = net.removeConfigs
	cbkServer["/write/server/manual/control"] = net.manualControl

	confServer := genericNetwork.NetworkConfig{
		IP:         conf.ServerIP,
		Port:       conf.ServerPort,
		ClientName: clientID,
		Callbacks:  cbkServer,
		LogLevel:   *conf.LogLevel,
	}

	for {
		rlog.Info("Try to connect to " + conf.ServerIP)
		err := net.Iface.Initialize(confServer)
		if err == nil {
			rlog.Info(clientID + " connected to server broker " + conf.ServerIP)
			return err
		}
		timer := time.NewTicker(time.Second)
		rlog.Error("Cannot connect to broker " + conf.ServerIP + " error: " + err.Error())
		rlog.Error("Try to reconnect " + conf.ServerIP + " in 1s")

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
}

func (net ServerNetwork) removeConfigs(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Debug("Received removeConfigs: Received topic: " + msg.Topic() + " payload: " + string(payload))
}

func (net ServerNetwork) manualControl(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Debug("Received manualControl: Received topic: " + msg.Topic() + " payload: " + string(payload))
}

//Disconnect from server
func (net ServerNetwork) Disconnect() {
	net.Iface.Disconnect()
}

//SendCommand to server
func (net ServerNetwork) SendCommand(topic, content string) error {
	return net.Iface.SendCommand(topic, content)
}
