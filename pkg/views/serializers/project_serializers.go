package serializers

type ProjectSerializer struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	ClusterId   string `json:"cluster_id" form:"cluster_id"`
	Namespace   string `json:"namespace" form:"namespace"`
	Owner       string `json:"owner" form:"owner"`
}

type ProjectCreateAppSerializer struct {
	ProjectId uint              `json:"project_id" form:"project_id"`
	Name      string            `json:"name" form:"name"`
	Version   string            `json:"version" form:"version"`
	Chart     string            `json:"chart" form:"chart"`
	Templates map[string]string `json:"templates" form:"templates"`
	Values    string            `json:"values" form:"values"`
}
