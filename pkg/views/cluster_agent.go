package views

import (
	"fmt"
	"github.com/gin-gonic/gin"
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
	agentYaml := fmt.Sprintf(clusterAgentYaml, token, serverUrl)
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
  name: osp

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: osp-admin
  namespace: osp

---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: osp-clusterrole-bind
  namespace: osp
subjects:
- kind: ServiceAccount
  name: osp-admin
  namespace: osp
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io

---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: ospagent
  namespace: osp
  labels:
    app: ospagent
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ospagent
  template:
    metadata:
      labels:
        app: ospagent
    spec:
      containers:
      - name: ospagent
        image: openspacee/ospagent:latest
        command:
        - "/ospagent"
        args:
        - --token=%s
        - --server-url=%s
      serviceAccountName: osp-admin
`
