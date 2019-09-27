package service

import (
	sd "github.com/energieip/common-components-go/pkg/dswitch"
	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/romana/rlog"
)

func (s *CoreService) addNewUser(user duser.UserAccess) {
	err := database.SaveUserConfig(s.db, user)
	if err != nil {
		rlog.Error("Cannot register new UserHash", user.UserHash)
		return
	}
	var switchs []string
	if user.Priviledge == "user" {
		for _, gr := range user.AccessGroups {
			for sw := range database.GetGroupSwitchs(s.db, gr) {
				switchs = append(switchs, sw)
			}
		}
	} else {
		//admin/maintainer should be send to every switch
		for sw := range database.GetSwitchsConfig(s.db) {
			switchs = append(switchs, sw)
		}
	}
	for _, mac := range switchs {
		sw, _ := database.GetSwitchConfig(s.db, mac)
		if sw != nil {
			ip := "0"
			if sw.IP != nil {
				ip = *sw.IP
			}
			url := "/write/switch/" + mac + "/update/settings"
			switchSetup := sd.SwitchConfig{}
			switchSetup.Mac = mac
			switchSetup.IP = ip
			switchSetup.Users = make(map[string]duser.UserAccess)
			switchSetup.Users[user.UserHash] = user
			dump, _ := switchSetup.ToJSON()
			s.server.SendCommand(url, dump)
		}
	}

	rlog.Info("Send new User Access for", user.UserHash)
}

func (s *CoreService) removeUser(user duser.UserAccess) {
	err := database.RemoveUserConfig(s.db, user.UserHash)
	if err != nil {
		rlog.Error("Cannot remove UserHash", user.UserHash)
		return
	}
	var switchs []string
	if user.Priviledge == "user" {
		for _, gr := range user.AccessGroups {
			for sw := range database.GetGroupSwitchs(s.db, gr) {
				switchs = append(switchs, sw)
			}
		}
	} else {
		//admin/maintainer should be send to every switch
		for sw := range database.GetSwitchsConfig(s.db) {
			switchs = append(switchs, sw)
		}
	}

	for _, mac := range switchs {
		sw, _ := database.GetSwitchConfig(s.db, mac)
		if sw != nil {
			ip := "0"
			if sw.IP != nil {
				ip = *sw.IP
			}
			url := "/write/switch/" + mac + "/remove/settings"
			switchSetup := sd.SwitchConfig{}
			switchSetup.Mac = mac
			switchSetup.IP = ip
			switchSetup.Users = make(map[string]duser.UserAccess)
			switchSetup.Users[user.UserHash] = user
			dump, _ := switchSetup.ToJSON()
			s.server.SendCommand(url, dump)
		}
	}
	rlog.Info("Send remove User Access for", user.UserHash)
}
