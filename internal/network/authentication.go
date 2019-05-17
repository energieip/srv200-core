package network

import (
	"encoding/json"
	"time"

	"github.com/energieip/common-components-go/pkg/duser"
	genericNetwork "github.com/energieip/common-components-go/pkg/network"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/romana/rlog"
)

const (
	EventNewUser    = "addUser"
	EventRemoveUser = "removeUser"
)

//AuthNetwork network object
type AuthNetwork struct {
	Iface  genericNetwork.NetworkInterface
	Events chan map[string]duser.UserAccess
}

//CreateAuthNetwork create network server object
func CreateAuthNetwork() (*AuthNetwork, error) {
	serverBroker, err := genericNetwork.NewNetwork(genericNetwork.MQTT)
	if err != nil {
		return nil, err
	}
	serverNet := AuthNetwork{
		Iface:  serverBroker,
		Events: make(chan map[string]duser.UserAccess),
	}
	return &serverNet, nil
}

//LocalConnection connect service to server broker
func (net AuthNetwork) LocalConnection(conf pkg.ServiceConfig, clientID string) error {
	cbkServer := make(map[string]func(genericNetwork.Client, genericNetwork.Message))
	cbkServer[EventNewUser] = net.onNewUser
	cbkServer[EventRemoveUser] = net.onRemoveUser

	confServer := genericNetwork.NetworkConfig{
		IP:         conf.AuthBroker.IP,
		Port:       conf.AuthBroker.Port,
		ClientName: clientID + "2",
		Callbacks:  cbkServer,
		LogLevel:   conf.LogLevel,
		User:       conf.AuthBroker.Login,
		Password:   conf.AuthBroker.Password,
		CaPath:     conf.AuthBroker.CaPath,
	}

	for {
		rlog.Info("Try to connect to " + conf.AuthBroker.IP)
		err := net.Iface.Initialize(confServer)
		if err == nil {
			rlog.Info(clientID + " connected to server broker " + conf.AuthBroker.IP)
			return err
		}
		timer := time.NewTicker(time.Second)
		rlog.Error("Cannot connect to broker " + conf.AuthBroker.IP + " error: " + err.Error())
		rlog.Error("Try to reconnect " + conf.AuthBroker.IP + " in 1s")

		select {
		case <-timer.C:
			continue
		}
	}
}

func (net AuthNetwork) onNewUser(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Info(msg.Topic() + " : " + string(payload))
	var user duser.UserAccess
	err := json.Unmarshal(payload, &user)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}

	event := make(map[string]duser.UserAccess)
	event[EventNewUser] = user
	net.Events <- event
}

func (net AuthNetwork) onRemoveUser(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Info(msg.Topic() + " : " + string(payload))
	var user duser.UserAccess
	err := json.Unmarshal(payload, &user)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}

	event := make(map[string]duser.UserAccess)
	event[EventRemoveUser] = user
	net.Events <- event
}

//Disconnect from server
func (net AuthNetwork) Disconnect() {
	net.Iface.Disconnect()
}

//SendCommand to server
func (net AuthNetwork) SendCommand(topic, content string) error {
	err := net.Iface.SendCommand(topic, content)
	if err != nil {
		rlog.Error("Cannot send : " + content + " on: " + topic + " Error: " + err.Error())
	} else {
		rlog.Info("Sent : " + content + " on: " + topic)
	}
	return err
}
