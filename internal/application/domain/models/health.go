package models

type ApplicationInfo struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Host         string `json:"host"`
	Port         string `json:"port"`
	Protocol     string `json:"protocol"`
	Architecture string `json:"architecture"`
	Duration     int    `json:"duration"`
	Path         string `json:"path"`
}

type ApplicationHealth struct {
	host map[string]*ApplicationInfo
}
