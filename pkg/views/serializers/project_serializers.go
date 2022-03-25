package serializers

type ProjectSerializer struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	ClusterId   string `json:"cluster_id" form:"cluster_id"`
	Namespace   string `json:"namespace" form:"namespace"`
	Owner       string `json:"owner" form:"owner"`
}

type ProjectCreateAppSerializer struct {
	ProjectId   uint              `json:"project_id" form:"project_id"`
	Name        string            `json:"name" form:"name"`
	Description string            `json:"description" form:"description"`
	Version     string            `json:"version" form:"version"`
	Chart       string            `json:"chart" form:"chart"`
	Templates   map[string]string `json:"templates" form:"templates"`
	Values      string            `json:"values" form:"values"`
}

type ProjectInstallAppSerializer struct {
	ProjectAppId uint   `json:"project_app_id" form:"project_app_id"`
	Values       string `json:"values" form:"values"`
	AppVersionId uint   `json:"app_version_id" form:"app_version_id"`
	Upgrade      bool   `json:"upgrade" form:"upgrade"`
}

type ImportStoreAppSerializers struct {
	ProjectId    uint `json:"project_id" form:"project_id"`
	StoreAppId   uint `json:"store_app_id" form:"store_app_id"`
	AppVersionId uint `json:"app_version_id" form:"app_version_id"`
}

type ProjectAppListSerializer struct {
	ProjectId     uint   `json:"project_id" form:"project_id"`
	Name          string `json:"name" form:"name"`
	Status        string `json:"status" form:"status"`
	WithWorkloads bool   `json:"with_workloads" form:"with_workloads"`
}

type ProjectAppVersionListSerializer struct {
	Scope   string `json:"scope" form:"scope"`
	ScopeId uint   `json:"scope_id" form:"scope_id"`
}

type ProjectAppVersionGetSerializer struct {
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
	AppId     uint   `json:"app_id" form:"app_id"`
	VersionId uint   `json:"version_id" form:"version_id"`
	Scope     string `json:"scope" form:"scope"`
	ScopeId   uint   `json:"scope_id" form:"scope_id"`
}
