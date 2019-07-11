package database

import (
	"github.com/energieip/common-components-go/pkg/pconst"
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

//SaveFrame dump project in database
func SaveFrame(db Database, cfg core.Frame) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = cfg.Label

	fr, dbID := GetFrame(db, cfg.Label)
	if fr == nil || dbID == "" {
		_, err := db.InsertRecord(pconst.DbConfig, pconst.TbFrames, cfg)
		return err
	}
	return db.UpdateRecord(pconst.DbConfig, pconst.TbFrames, dbID, cfg)
}

//RemoveFrame remove project entry in database
func RemoveFrame(db Database, label string) error {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	return db.DeleteRecord(pconst.DbConfig, pconst.TbFrames, criteria)
}

//GetFrame return the project configuration
func GetFrame(db Database, label string) (*core.Frame, string) {
	criteria := make(map[string]interface{})
	criteria["Label"] = label
	stored, err := db.GetRecord(pconst.DbConfig, pconst.TbFrames, criteria)
	if err != nil || stored == nil {
		return nil, ""
	}
	var dbID string
	m := stored.(map[string]interface{})
	id, ok := m["id"]
	if ok {
		dbID = id.(string)
	}
	project, err := core.ToFrame(stored)
	if err != nil {
		return nil, dbID
	}
	return project, dbID
}

//GetFrames return the frame configuration
func GetFrames(db Database) []core.Frame {
	var projects []core.Frame
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbFrames)
	if err != nil || stored == nil {
		return nil
	}
	for _, st := range stored {
		project, err := core.ToFrame(st)
		if err != nil || project == nil {
			continue
		}
		projects = append(projects, *project)
	}
	return projects
}

//GetFramesConfigByLabel return the frame configuration
func GetFramesConfigByLabel(db Database) map[string]core.Frame {
	projects := make(map[string]core.Frame)
	stored, err := db.FetchAllRecords(pconst.DbConfig, pconst.TbFrames)
	if err != nil || stored == nil {
		return nil
	}
	for _, st := range stored {
		project, err := core.ToFrame(st)
		if err != nil || project == nil {
			continue
		}
		projects[project.Label] = *project
	}
	return projects
}

//GetFramesDumpByLabel return the switch status list
func GetFramesDumpByLabel(db Database) map[string]core.FrameStatus {
	frames := make(map[string]core.FrameStatus)

	configs := GetFramesConfigByLabel(db)
	for _, fr := range configs {
		ledsPower := int64(0)
		blindsPower := int64(0)
		hvacsPower := int64(0)
		totalPower := int64(0)
		ledsEnergy := int64(0)
		blindsEnergy := int64(0)
		hvacsEnergy := int64(0)
		totalEnergy := int64(0)
		baes := 0
		profil := "none"
		puls1 := 1
		puls2 := 1
		puls3 := 1
		puls4 := 1
		puls5 := 1

		clusters := GetSwitchStatusCluster(db, fr.Cluster)
		for _, cl := range clusters {
			ledsPower += cl.LedsPower
			blindsPower += cl.BlindsPower
			hvacsPower += cl.HvacsPower
			ledsEnergy += cl.LedsEnergy
			blindsEnergy += cl.BlindsEnergy
			hvacsEnergy += cl.HvacsEnergy
			totalPower += cl.TotalPower
			totalEnergy += cl.TotalEnergy
			if cl.StateBaes != 0 {
				baes = cl.StateBaes
			}
			if cl.Profil == "puls" {
				profil = "puls"
				if cl.StatePuls1 == 0 {
					puls1 = cl.StatePuls1
				}
				if cl.StatePuls2 == 0 {
					puls2 = cl.StatePuls2
				}
				if cl.StatePuls3 == 0 {
					puls3 = cl.StatePuls3
				}
				if cl.StatePuls4 == 0 {
					puls4 = cl.StatePuls4
				}
				if cl.StatePuls5 == 0 {
					puls5 = cl.StatePuls5
				}
			}
		}

		frame := core.FrameStatus{
			Label:        fr.Label,
			FriendlyName: fr.FriendlyName,
			Cluster:      fr.Cluster,
			LedsPower:    ledsPower,
			BlindsPower:  blindsPower,
			HvacsPower:   hvacsPower,
			TotalPower:   totalPower,
			LedsEnergy:   ledsEnergy,
			BlindsEnergy: blindsEnergy,
			HvacsEnergy:  hvacsEnergy,
			TotalEnergy:  totalEnergy,
			StateBaes:    baes,
			Profil:       profil,
			StatePuls1:   puls1,
			StatePuls2:   puls2,
			StatePuls3:   puls3,
			StatePuls4:   puls4,
			StatePuls5:   puls5,
		}
		frames[fr.Label] = frame
	}
	return frames
}
