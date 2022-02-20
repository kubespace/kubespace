package serializers

type ProjectSerializer struct {
	Name         string    `json:"name" form:"name"`
	Description         string    `json:"description" form:"description"`
	ClusterId         string    `json:"cluster_id" form"cluster_id"`
	Namespace         string    `json:"namespace" form:"namespace"`
	Owner   string    `json:"owner" form:"owner"`
}
