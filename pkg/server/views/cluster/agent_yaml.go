package cluster

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
