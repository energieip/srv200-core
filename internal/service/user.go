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
	for _, gr := range user.AccessGroups {
		for sw := range database.GetGroupSwitchs(s.db, gr) {
			url := "/write/switch/" + sw + "/update/settings"
			switchSetup := sd.SwitchConfig{}
			switchSetup.Mac = sw
			switchSetup.Users = make(map[string]duser.UserAccess)
			switchSetup.Users[user.UserHash] = user
			dump, _ := switchSetup.ToJSON()
			s.server.SendCommand(url, dump)
		}
	}
	rlog.Info("Send new User Access for", user.UserHash)
}
