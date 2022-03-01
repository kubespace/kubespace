export { default as Container } from './container'
export { default as PodVolume } from './pod_volume'
export { default as PodNetwork } from './pod_network'
export { default as PodAffinity } from './pod_affinity'
export { default as PodSecurity } from './pod_security'
export { default as Workload } from './workload'
export { default as Service } from './service'

export function kindTemplate(kind) {
  if(kind == 'Workload') return workloadTemplate()
  if(kind == 'Service') return serviceTemplate()
}

export function transferTemplate(template) {
  let tpl = JSON.parse(JSON.stringify(template))
  if(!tpl) return {err: "应用资源为空"}
  if(!tpl.kind) return {err: "应用资源kind为空"}
  if(!tpl.metadata) return {err: "应用资源metadata为空"}
  if(!tpl.metadata.name) return {err: "应用资源名称为空"}
  if(!tpl.spec) return {err: `应用资源${tpl.kind}/${tpl.metadata.name} spec为空`}
  tpl.metadata.labels['kubespace.cn/app'] = tpl.metadata.name
  if(['Deployment', 'StatefulSet'].indexOf(tpl.kind) > -1) return transferWorkload(tpl)
  if(tpl.kind == 'Service') return transferService(tpl)
  return {err: `${tpl.kind}/${tpl.metadata.name}未找到对应的资源类型`}
}

function transferWorkload(tpl) {
  if(!tpl.spec.template.spec.containers || tpl.spec.template.spec.containers.length <= 0) {
    return {err: `应用资源${tpl.kind}/${tpl.metadata.name}容器为空`}
  }
  tpl.spec.selector.matchLabels['kubespace.cn/app'] = tpl.metadata.name
  tpl.spec.template.metadata.labels['kubespace.cn/app'] = tpl.metadata.name
  let err = transferContainer(tpl)
  if(err) return err
  err = transferPodVolume(tpl)
  if(err) return err
  if(tpl.spec.template.spec.hostAliases.length > 0) {
    for(let h of tpl.spec.template.spec.hostAliases) {
      if(!h.hostnames) return {err: `应用资源${tpl.kind}/${tpl.metadata.name}主机别名域名为空`}
      if(!h.ip) return {err: `应用资源${tpl.kind}/${tpl.metadata.name}主机别名ip为空`}
      h.hostnames = [h.hostnames]
    }
  } else {
    delete tpl.spec.template.spec.hostAliases
  }
  return {tpl}
}

function transferContainer(tpl) {
  let initContainers = []
  let containers = []
  for(let c of tpl.spec.template.spec.containers) {
    if(!c.name) {
      return {err: `应用资源${tpl.kind}/${tpl.metadata.name}容器名称为空`}
    }
    if(!c.image) {
      return {err: `应用资源${tpl.kind}/${tpl.metadata.name}中容器镜像为空`}
    }
    let p = transferProbe(c.livenessProbe)
    if(p) {
      c.livenessProbe = p
    } else {
      delete c.livenessProbe
    }
    p = transferProbe(c.readinessProbe)
    if(p) {
      c.readinessProbe = p
    } else {
      delete c.readinessProbe
    }
    if(c.command) {
      try{
        c.command = JSON.parse(c.command)
      }catch(e){
        c.command = [c.command]
      }
    } else {
      c.command = []
    }
    if(c.args) {
      try{
        c.args = JSON.parse(c.args)
      }catch(e){
        c.args = [c.args]
      }
    } else {
      c.args = []
    }
    if(c.securityContext.runAsUser) {
      c.securityContext.runAsUser = parseInt(c.securityContext.runAsUser)
    }
    if(c.securityContext.runAsGroup) {
      c.securityContext.runAsGroup = parseInt(c.securityContext.runAsGroup)
    }
    if(c.init){
      initContainers.push(c)
    } else {
      containers.push(c)
    }
    delete c.init
  }
  tpl.spec.template.spec.containers = containers
  tpl.spec.template.spec.initContainers = initContainers
}

function transferProbe(probe) {
  if(!probe.probe) return 
  let obj = {
    successThreshold: probe.successThreshold, 
    failureThreshold: probe.failureThreshold,
    initialDelaySeconds: probe.initialDelaySeconds, 
    timeoutSeconds: probe.timeoutSeconds, 
    periodSeconds: probe.periodSeconds
  } 
  if(probe.type == 'http' || probe.type == 'https') {
    obj['httpGet'] = {
      path: obj.handle.path,
      port: obj.handle.port,
      scheme: 'HTTP'
    }
    if(probe.type == 'https') obj.httpGet.scheme = 'HTTPS'
  }
  if(probe.type == 'cmd') {
    obj['exec'] = {
      command: probe.handle.command
    }
  }
  if(probe.type == 'tcp') {
    obj['tcpSocket'] = {
      port: probe.handle.command
    }
  }
  return obj
}

function transferService(tpl) {
  return {tpl}
}

export function newContainer() {
  return {
    init: false,
    name: '',
    image: '',
    command: '',
    args: '',
    workingDir: '',
    ports: [],
    env: [],
    resources: {limits: {}, requests: {}},
    livenessProbe: {probe: false, type: 'http', handle: {}, successThreshold: 1, failureThreshold: 3,
                    initialDelaySeconds: 0, timeoutSeconds: 1, periodSeconds: 10},
    readinessProbe: {probe: false, type: 'http', handle: {}, successThreshold: 1, failureThreshold: 3,
                    initialDelaySeconds: 0, timeoutSeconds: 1, periodSeconds: 10},
    imagePullPolicy: '',
    volumeMounts: [],
    stdin: false,
    tty: false,
    securityContext: {seLinuxOptions: {}, capabilities: {add: [], drop: []}},
  }
}

function transferPodVolume(tpl) {
  let vols = []
  for(let v of tpl.spec.template.spec.volumes) {
    if(!v.name) return {err: `应用资源${tpl.kind}/${tpl.metadata.name}中存储卷名称为空`}
    let vol = {
      name: v.name
    }
    if(v.type == 'configMap' || v.type == 'secret') {
      vol[v.type] = {
        items: v[v.type].items,
      }
      if(v[v.type].defaultMode) {
        vol[v.type].defaultMode = parseInt(v[v.type].defaultMode, 8)
      }
      if(v.type == 'configMap') {
        vol[v.type]['name'] = v[v.type].obj.metadata.name
      } else {
        vol[v.type]['secretName'] = v[v.type].obj.metadata.name
      }
    } else {
      vol[v.type] = v[v.type]
    }
    vols.push(vol)
  }
  if(vols.length > 0) tpl.spec.template.spec.volumes = vols
}

export function newPodVolume() {
  return {
    name: '',
    type: 'persistentVolumeClaim',
    persistentVolumeClaim: {},
    glusterfs: {},
    nfs: {},
    secret: {items: [], obj: {keys: []}},
    configMap: {items: [], obj: {keys: []}},
    emptyDir: {},
    hostPath: {}
  }
}

function workloadTemplate() {
  return {
    kind: "Deployment",
    apiVersion: "apps/v1",
    metadata: {
      name: "",
      labels: {},
      namespace: "{{ .Release.Namespace }}"
    },
    spec: {
      replicas: 1,
      selector: {
        matchLabels: {}
      },
      template: {
        metadata: {
          labels: {}
        },
        spec: {
          nodeSelector: [],
          tolerations: [],
          affinity: {nodeAffinity: [], podAffinity: [], podAntiAffinity: []},
          securityContext: {sysctls: [], seLinuxOptions: {}},
          hostAliases: [],
          containers: [newContainer()],
          volumes: [],
        }
      }
    }
  }
}

function serviceTemplate() {
  return {
    kind: "Service",
    apiVersion: "v1",
    metadata: {
      name: "",
      labels: {},
      namespace: "{{ .Release.Namespace }}"
    },
    spec: {
      ports: [],
      selector: {},
      type: 'ClusterIP',
    }
  }
}