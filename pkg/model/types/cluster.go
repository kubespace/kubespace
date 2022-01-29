package types

const (
	ClusterPending = "Pending"
	ClusterConnect = "Connect"
)

type ClusterStore struct {
	Common

	Name      string `json:"name"`
	Token     string `json:"token"`
	Status    string `json:"status"`
	CreatedBy string `json:"created_by"`
	Members   string `json:"members"`
}

type Cluster struct {
	Common

	Name      string   `json:"name"`
	Token     string   `json:"token"`
	Status    string   `json:"status"`
	CreatedBy string   `json:"created_by"`
	Members   []string `json:"members"`
}

type Cluster_ struct {

}
