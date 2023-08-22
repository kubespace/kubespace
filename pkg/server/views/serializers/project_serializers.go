package serializers

type ProjectSerializer struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	ClusterId   string `json:"cluster_id" form:"cluster_id"`
	Namespace   string `json:"namespace" form:"namespace"`
	Owner       string `json:"owner" form:"owner"`
}

type ProjectCloneAppSerializer struct {
	Id uint `json:"id" form:"id"`
}

type ProjectCloneResourceSerializer struct {
	Type string `json:"type" form:"type"`
	Name string `json:"name" form:"name"`
}

type ProjectCloneSerializer struct {
	OriginProjectId uint                              `json:"origin_project_id" form:"origin_project_id"`
	Name            string                            `json:"name" form:"name"`
	Description     string                            `json:"description" form:"description"`
	ClusterId       string                            `json:"cluster_id" form:"cluster_id"`
	Namespace       string                            `json:"namespace" form:"namespace"`
	Owner           string                            `json:"owner" form:"owner"`
	Resources       []*ProjectCloneResourceSerializer `json:"resources" form:"resources"`
	Apps            []*ProjectCloneAppSerializer      `json:"apps" form:"apps"`
}

type ProjectDeleteSerializer struct {
	DelResource bool `json:"del_resource" form:"del_resource"`
}

type CreateAppSerializer struct {
	Scope              string `json:"scope" form:"scope"`
	ScopeId            uint   `json:"scope_id" form:"scope_id"`
	Name               string `json:"name" form:"name"`
	From               string `json:"from" form:"from"`
	Type               string `json:"type" form:"type"`
	Description        string `json:"description" form:"description"`
	VersionDescription string `json:"version_description" form:"version_description"`
	Version            string `json:"version" form:"version"`
	//Chart              string                 `json:"chart" form:"chart"`
	//Templates          map[string]string      `json:"templates" form:"templates"`
	Values     string                 `json:"values" form:"values"`
	ChartFiles map[string]interface{} `json:"chart_files"`
}

type InstallAppSerializer struct {
	AppId        uint   `json:"project_app_id" form:"project_app_id"`
	Values       string `json:"values" form:"values"`
	AppVersionId uint   `json:"app_version_id" form:"app_version_id"`
	Upgrade      bool   `json:"upgrade" form:"upgrade"`
}

type ImportStoreAppSerializers struct {
	Scope        string `json:"scope" form:"scope"`
	ScopeId      uint   `json:"scope_id" form:"scope_id"`
	Namespace    string `json:"namespace" form:"namespace"`
	StoreAppId   uint   `json:"store_app_id" form:"store_app_id"`
	AppVersionId uint   `json:"app_version_id" form:"app_version_id"`
}

type AppListSerializer struct {
	Scope         string `json:"scope" form:"scope"`
	ScopeId       uint   `json:"scope_id" form:"scope_id"`
	Name          string `json:"name" form:"name"`
	Status        string `json:"status" form:"status"`
	WithWorkloads bool   `json:"with_workloads" form:"with_workloads"`
}

type AppVersionListSerializer struct {
	Scope   string `json:"scope" form:"scope"`
	ScopeId uint   `json:"scope_id" form:"scope_id"`
}

type AppVersionGetSerializer struct {
	AppVersionId string `json:"app_version_id" form:"app_version_id"`
}

type AppStoreCreateSerializer struct {
	Name               string `json:"name" form:"name"`
	PackageVersion     string `json:"package_version" form:"package_version"`
	AppVersion         string `json:"app_version" form:"app_version"`
	Description        string `json:"description" form:"description"`
	VersionDescription string `json:"version_description" form:"version_description"`
	Type               string `json:"type" form:"type"`
}

type GetStoreAppSerializer struct {
	WithVersions bool `json:"with_versions" form:"with_versions"`
}

type DuplicateAppSerializer struct {
	Name      string `json:"name" form:"name"`
	AppId     uint   `json:"app_id" form:"app_id"`
	VersionId uint   `json:"version_id" form:"version_id"`
	Scope     string `json:"scope" form:"scope"`
	ScopeId   uint   `json:"scope_id" form:"scope_id"`
}

type ProjectResourcesSerializer struct {
	ProjectId uint `json:"project_id" form:"project_id"`
}

type ImportCustomAppSerializer struct {
	Scope              string `json:"scope" form:"scope"`
	ScopeId            uint   `json:"scope_id" form:"scope_id"`
	Name               string `json:"name" form:"name"`
	PackageVersion     string `json:"package_version" form:"package_version"`
	AppVersion         string `json:"app_version" form:"app_version"`
	Description        string `json:"description" form:"description"`
	VersionDescription string `json:"version_description" form:"version_description"`
	Type               string `json:"type" form:"type"`
}
