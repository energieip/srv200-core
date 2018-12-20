package database

import (
	"github.com/energieip/srv200-coreservice-go/internal/core"
)

func GetIfcs(db Database) map[string]core.IfcInfo {
	res := make(map[string]core.IfcInfo)
	projects := GetProjects(db)
	models := GetModels(db)

	for _, project := range projects {
		model, ok := models[project.ModelName]
		if !ok {
			continue
		}

		res[project.Mac] = core.IfcInfo{
			Label:      project.Label,
			ModelName:  model.Name,
			Mac:        project.Mac,
			Vendor:     model.Vendor,
			URL:        model.URL,
			DeviceType: model.DeviceType,
		}
	}
	return res
}
