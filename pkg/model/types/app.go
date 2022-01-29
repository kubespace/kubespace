package types

type App struct {
	Name         string `json:"name"`
	ChartVersion string `json:"chart_version"`
	AppVersion   string `json:"app_version"`
	Icon         string `json:"icon"`
	Description  string `json:"description"`
	Chart        string `json:"chart"`
}
