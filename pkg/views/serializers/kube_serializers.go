package serializers

type ListSerializers struct {
	Cluster       string            `json:"cluster" form:"cluster"`
	Name          string            `json:"name" form:"name"`
	Namespace     string            `json:"namespace" form:"namespace"`
	Labels        map[string]string `json:"labels" form:"labels"`
	LabelSelector interface{}       `json:"label_selector" form:"label_selector"`
	CronjobUID    string            `json:"cronjob_uid" form:"cronjob_uid"`
	Names         []string          `json:"names"`
}

type GetSerializers struct {
	Cluster   string `json:"cluster" form:"cluster"`
	Name      string `json:"name" form:"name"`
	Namespace string `json:"namespace" form:"namespace"`
	Output    string `json:"output" form:"output"`
	Kind      string `json:"kind" form:"kind"`
	GetOption string `json:"get_option" form:"get_option"`
}

type GetAppSerializers struct {
	Name         string `json:"name" form:"name"`
	ChartVersion string `json:"chart_version" form:"chart_version"`
}

type CreateAppSerializers struct {
	Name         string                 `json:"name" form:"name"`
	Namespace    string                 `json:"namespace" form:"namespace"`
	ChartVersion string                 `json:"chart_version"`
	ReleaseName  string                 `json:"release_name"`
	Values       map[string]interface{} `json:"values"`
}

type DeleteResource struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type DeleteSerializers struct {
	Resources []DeleteResource `json:"resources"`
}

type UpdateSerializers struct {
	Yaml string `json:"yaml"`
	Kind string `json:"kind"`
}

type UpdateWorkloadSerializers struct {
	Replicas int `json:"replicas" form:"replicas"`
}

type EventListSerializers struct {
	UID       string `json:"uid" form:"uid"`
	Kind      string `json:"kind" form:"kind"`
	Name      string `json:"name" form:"name"`
	Namespace string `json:"namespace" form:"namespace"`
}

type UpdateMapSerializer struct {
	Data map[string]string
}
