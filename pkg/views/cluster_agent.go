package views

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kubespace/kubespace/pkg/conf"
	"github.com/kubespace/kubespace/pkg/model"
	"k8s.io/klog"
	"net/http"
)

type ClusterAgent struct {
	models *model.Models
}

func NewClusterAgent(models *model.Models) *ClusterAgent {
	return &ClusterAgent{
		models: models,
	}
}

func (a *ClusterAgent) AgentYaml(c *gin.Context) {
	token := c.Param("token")
	serverUrl := a.resolveHost(c.Request)
	klog.Info("server url: ", serverUrl)
	agentYaml := fmt.Sprintf(conf.AppConfig.AgentVersion, clusterAgentYaml, token, serverUrl)
	c.String(200, agentYaml)
}

func (a *ClusterAgent) resolveHost(r *http.Request) (host string) {
	switch {
	case r.Header.Get("X-Host") != "":
		return r.Header.Get("X-Host")
	case r.Host != "":
		return r.Host
	case r.URL.Host != "":
		return r.URL.Host
	default:
		return r.Host
	}
}

var clusterAgentYaml = `
---

apiVersion: v1
kind: Namespace
metadata:
  name: kubespace

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubespace
  namespace: kubespace

---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: kubespace-admin
  namespace: kubespace
subjects:
- kind: ServiceAccount
  name: kubespace
  namespace: kubespace
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io

---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubespace-agent
  namespace: kubespace
  labels:
    kubespace-app: agent
spec:
  replicas: 1
  selector:
    matchLabels:
      kubespace-app: agent
  template:
    metadata:
      labels:
        kubespace-app: agent
    spec:
      containers:
      - name: agent
        image: kubespace/agent:%s
        command:
        - "/agent"
        args:
        - --token=%s
        - --server-url=%s
      serviceAccountName: kubespace
`
