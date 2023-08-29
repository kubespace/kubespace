package types

import "time"

const (
	ClusterFailed  = "Failed"
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
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name1      string    `gorm:"column:name;size:500;not null;uniqueIndex" json:"name1"`
	Name       string    `gorm:"-" json:"name"`
	KubeConfig string    `gorm:"type:text;column:kubeconfig" json:"-"`
	Token      string    `gorm:"size:255;not null;uniqueIndex" json:"token"`
	Status     string    `gorm:"size:50;" json:"status"`
	CreatedBy  string    `gorm:"size:255;not null;" json:"created_by"`
	Members    []string  `gorm:"-" json:"members"`
	CreateTime time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}
