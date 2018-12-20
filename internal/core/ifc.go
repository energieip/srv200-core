package core

//IfcInfo ifc component description
type IfcInfo struct {
	Label      string `json:"label"` //cable label
	ModelName  string `json:"modelName"`
	Mac        string `json:"mac"` //device Mac address
	Vendor     string `json:"vendor"`
	URL        string `json:"url"`
	DeviceType string `json:"deviceType"`
}
