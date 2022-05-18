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
	agentYaml := fmt.Sprintf(clusterAgentYaml, conf.AppConfig.AgentRepository, conf.AppConfig.AgentVersion, token, serverUrl)
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
apiVersion: v1
kind: Namespace
metadata:
  name: kubespace

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubespace-agent
  namespace: kubespace

---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: kubespace-agent
  namespace: kubespace
subjects:
- kind: ServiceAccount
  name: kubespace-agent
  namespace: kubespace
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io

---

kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kubespace-agent
  namespace: kubespace
subjects:
- kind: ServiceAccount
  name: kubespace-agent
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
    kubespace-app: kubespace-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/instance: kubespace
      app.kubernetes.io/name: kubespace
      kubespace-app: kubespace-agent
  template:
    metadata:
      labels:
        app.kubernetes.io/instance: kubespace
        app.kubernetes.io/name: kubespace
        kubespace-app: kubespace-agent
    spec:
      containers:
      - name: kubespace-agent
        image: %s:%s
        command:
        - "/agent"
        args:
        - --token=%s
        - --server-url=%s
        env:
        - name: TZ
          value: Asia/Shanghai
      serviceAccountName: kubespace-agent
`
