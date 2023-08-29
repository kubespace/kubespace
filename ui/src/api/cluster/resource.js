import request from '@/utils/request';
import { Selector } from 'k8s-selector';
import { dateFormat } from '@/utils/utils';
import qs from 'qs'
import { Message } from 'element-ui'
import { Base64 } from 'js-base64';

export const ResType = {
    Namespace: "namespace",
    Cluster: "cluster",
    Pod: "pod",
    Event: "event",
    Node: "node",
    Deployment: "deployment",
    Statefulset: "statefulset",
    Daemonset: "daemonset",
    Job: "job",
    CronJob: "cronjob",
    ConfigMap: "configmap",
    Secret: "secret",
    Hpa: "horizontalPodAutoscaler",
    Service: "service",
    Ingress: "ingress",
    Endpoint: "endpoints",
    NetworkPolicy: "networkpolicy",
    PersistentVolume: "persistentVolume",
    PersistentVolumeClaim: "persistentVolumeClaim",
    StorageClass: "storageclass",
    ServiceAccount: "serviceaccount",
    CRD: "crd",
    CR: "cr",
    Role: "role",
    ClusterRole: "clusterrole",
    RoleBinding: "rolebinding",
    ClusterRoleBinding: "clusterrolebinding",
}

export function listResource(cluster, resType, data, params) {
  return request({
    url: `cluster/${cluster}/${resType}/list`,
    method: 'post',
    data: data,
    params,
  })
}

export function getResource(cluster, resType, namespace, name, output='', params) {
  if(params) {
    params['output'] = output
  } else {
    params = {output}
  }
  return request({
    url: `cluster/${cluster}/${resType}/${namespace?"namespace/"+namespace+"/":""}${name}`,
    method: 'get',
    params
  })
}

export function delResource(cluster, resType, data, params) {
  return request({
    url: `cluster/${cluster}/${resType}/delete`,
    method: 'post',
    data,
    params
  })
}

export function updateResource(cluster, resType, namespace, name, yamlStr, params) {
  return request({
    url: `cluster/${cluster}/${resType}/${namespace?"namespace/"+namespace+"/":""}${name}`,
    method: 'put',
    data: {yaml: yamlStr},
    params
  })
}

export function patchResource(cluster, resType, data, params) {
  return request({
    url: `cluster/${cluster}/${resType}/patch`,
    method: 'post',
    data,
    params
  })
}

export function createResource(cluster, yaml, params) {
  return request({
    url: `cluster/${cluster}/apply`,
    method: 'post',
    data: {yaml, create: true},
    params,
  })
}

export function applyResource(cluster, yaml, params) {
  return request({
    url: `cluster/${cluster}/apply`,
    method: 'post',
    data: {yaml},
    params
  })
}

export function watchResource(sse, cluster, resType, watchFunc, params) {
  let p = qs.stringify(params)
  let url = `/api/v1/cluster/${cluster}/${resType}/watch?${p}`
  let clusterSSE = sse.create({
    url: url,
    includeCredentials: false,
    format: 'plain'
  });
  clusterSSE.on("message", (res) => {
    // console.log(res)
    if(res && res != "\n" && res != "{}") {
      try{
        var data = JSON.parse(res)
      } catch(err) {
          Message.error(res)
          clusterSSE.disconnect()
          return
      }
      if(!data) return
      let d = {"resource": data.Object}
      if (data.Type == "ADDED") d.event="add"
      else if(data.Type == "MODIFIED") d.event = "update"
      else if(data.Type == "DELETED") d.event = "delete"
      watchFunc(d)
    }
  })
  clusterSSE.connect().then(() => {
    console.log('[info] connected', 'system')
  }).catch(() => {
    console.log('[error] failed to connect', 'system')
  })
  clusterSSE.on('error', () => { // eslint-disable-line
    console.log('[error] disconnected, automatically re-attempting connection', 'system')
  })
  return clusterSSE
}

export function buildContainer(container, statuses) {
  if (!container) return {}
  // if (!statuses) return {}
  let c = {
    name: container.name,
    status: 'unknown',
    image: container.image,
    restarts: 0,
    command: container.command ? container.command : [],
    args: container.args ? container.args : [],
    ports: container.ports ? container.ports : [],
    env: container.env ? container.env : [],
    volume_mounts: container.volumeMounts ? container.volumeMounts : [],
    image_pull_policy: container.imagePullPolicy ? container.imagePullPolicy : '',
    resources: container.resources ? container.resources : {},
    start_time: '',
    liveness_probe: container.livenessProbe,
    readiness_probe: container.readinessProbe,
  }
  if(statuses) {
    for (let s of statuses) {
      if (s.name == container.name) {
        c.restarts = s.restartCount
        if (s.state.running) {
          c.status = 'running'
          c.start_time = s.state.running.startedAt
        }
        else if (s.state.terminated) {
          c.status = 'terminated'
          c.start_time = s.state.terminated.startedAt
        }
        else if (s.state.waiting) {
          c.status = 'waiting'
        }
        c.ready = s.ready
        break
      }
    }
  }
  return c
}

export function buildPods(pod) {
  if (!pod) return {}
  let containers = []
  for (let c of pod.spec.containers) {
    let bc = buildContainer(c, pod.status.containerStatuses)
    containers.push(bc)
  }
  let init_containers = []
  if (pod.spec.initContainers) {
    for (let c of pod.spec.initContainers) {
      init_containers.push(buildContainer(c, pod.status.initContainerStatuses))
    }
  }
  let controlled = ''
  let controlled_name = ''
  if (pod.metadata.ownerReferences && pod.metadata.ownerReferences.length > 0) {
    controlled = pod.metadata.ownerReferences[0].kind
    controlled_name = pod.metadata.ownerReferences[0].name
  }
  let p = {
    uid: pod.metadata.uid,
    name: pod.metadata.name,
    namespace: pod.metadata.namespace,
    containers: containers,
    init_containers: init_containers,
    controlled: controlled,
    controlled_name: controlled_name,
    qos: pod.status.qosClass,
    status: pod.status.phase,
    ip: pod.status.podIP,
    created: dateFormat(pod.metadata.creationTimestamp),
    node_name: pod.spec.nodeName,
    resource_version: pod.metadata.resourceVersion,
    labels: pod.metadata.labels,
    annonations: pod.metadata.annotations,
    service_account: pod.spec.serviceAccountName || pod.spec.serviceAccount,
    node_selector: pod.spec.nodeSelector,
    volumes: pod.spec.volumes,
    conditions: pod.status.conditions,
  }
  p['containerNum'] = p.containers.length
  if (p.init_containers){
    p['containerNum'] += p.init_containers.length
  }
  p['restarts'] = 0
  for (let c of p.containers) {
    if (c.restarts > p['restarts']) {
      p['restarts'] = c.restarts
    }
  }
  return p
}

export function resourceFor(r, r_type, f_type) {
  if (r_type in r && f_type in r[r_type]) return r[r_type][f_type]
  return '-'
}

export function containerClass(status) {
  if (status === 'running') return 'running-class'
  if (status === 'terminated') return 'terminate-class'
  if (status === 'waiting') return 'waiting-class'
}

export function podMatch(selector, labels) {
  let s = Selector(selector)
//   console.log(s, selector, labels)
  return s(labels)
}

export function envStr(env) {
  let s = env.name + ': '
  if (env.value) {
    s += env.value
  } else if (env.valueFrom) {
    if (env.valueFrom.configMapKeyRef) {
      s += `configmap(${env.valueFrom.configMapKeyRef.key}:${env.valueFrom.configMapKeyRef.name})`
    } else if (env.valueFrom.fieldRef) {
      s += `fieldRef(${env.valueFrom.fieldRef.apiVersion}:${env.valueFrom.fieldRef.fieldPath})`
    } else if (env.valueFrom.secretKeyRef) {
      s += `secret(${env.valueFrom.secretKeyRef.key}:${env.valueFrom.secretKeyRef.name})`
    } else {
      s += String(env.valueFrom)
    }
  }
  return s
}

export function buildDeployment(deployment) {
  if (!deployment) return {}
  let p = {
    uid: deployment.metadata.uid,
    namespace: deployment.metadata.namespace,
    name: deployment.metadata.name,
    replicas: deployment.spec.replicas,
    status_replicas: deployment.status.replicas || 0,
    ready_replicas: deployment.status.readyReplicas || 0,
    update_replicas: deployment.status.updateReplicas || 0,
    available_replicas: deployment.status.availableReplicas || 0,
    unavailable_replicas: deployment.status.unavailabelReplicas || 0,
    resource_version: deployment.metadata.resourceVersion,
    strategy: deployment.spec.strategy.type,
    conditions: deployment.status.conditions,
    created: deployment.metadata.creationTimestamp,
    label_selector: deployment.spec.selector,
    labels: deployment.metadata.labels,
    annotations: deployment.metadata.annotations,
    volumes: deployment.spec.template.spec.volumes,  
  }
  return p
}

export function buildDaemonSet(daemonset) {
  if (!daemonset) return {}
  let p = {
    uid: daemonset.metadata.uid,
    namespace: daemonset.metadata.namespace,
    name: daemonset.metadata.name,
    desired_number_scheduled: daemonset.status.desiredNumberScheduled || 0,
    number_ready: daemonset.status.numberReady || 0,
    resource_version: daemonset.metadata.resourceVersion,
    strategy: daemonset.spec.updateStrategy.type,
    conditions: daemonset.status.conditions,
    created: daemonset.metadata.creationTimestamp,
    label_selector: daemonset.spec.selector,
    labels: daemonset.metadata.labels,
    annotations: daemonset.metadata.annotations,
    volumes: daemonset.spec.template.spec.volumes,
  }
  return p
}

export function buildJobs(job) {
  if (!job) return
  var conditions = []
  if(job.status.conditions) {
    for (let c of job.status.conditions) {
      if (c.status === "True") {
        conditions.push(c.type)
      }
    }
  }
  let p = {
    uid: job.metadata.uid,
    namespace: job.metadata.namespace,
    name: job.metadata.name,
    completions: job.spec.completions || 0,
    active: job.status.active || 0,
    succeeded: job.status.succeeded || 0,
    failed: job.status.failed || 0,
    resource_version: job.metadata.resourceVersion,
    conditions: conditions,
    node_selector: job.spec.template.spec.nodeSelector,
    created: job.metadata.creationTimestamp,
    label_selector: job.spec.selector,
    volumes: job.spec.template.spec.volumes,
    labels: job.metadata.labels,
    annotations: job.metadata.annotations,
  }
  return p
}

export function buildStatefulSet(statefulset) {
  if (!statefulset) return {}
  let p = {
    uid: statefulset.metadata.uid,
    namespace: statefulset.metadata.namespace,
    name: statefulset.metadata.name,
    replicas: statefulset.spec.replicas,
    status_replicas: statefulset.status.replicas || 0,
    ready_replicas: statefulset.status.readyReplicas || 0,
    resource_version: statefulset.metadata.resourceVersion,
    strategy: statefulset.spec.updateStrategy.type,
    conditions: statefulset.status.conditions,
    created: statefulset.metadata.creationTimestamp,
    label_selector: statefulset.spec.selector,
    labels: statefulset.metadata.labels,
    annotations: statefulset.metadata.annotations,
    volumes: statefulset.spec.template.spec.volumes,
  }
  return p
}

export function transferSecret(secret) {
  let data = {}
  if(["kubernetes.io/tls", "kubernetes.io/basic-auth", "kubernetes.io/dockerconfigjson"].indexOf(secret.type) == -1) {
    for(let d of secret.data) {
      if(!d.key){
        return `配置项key不能为空`
      }
      data[d.key] = Base64.encode(d.value)
    }
    secret.data = data
  } else if(secret.type == 'kubernetes.io/tls') {
    secret.data = {
      'tls.crt': Base64.encode(secret.tls['crt']),
      'tls.key': Base64.encode(secret.tls['key'])
    }
  } else if(secret.type == 'kubernetes.io/basic-auth') {
    secret.data = {
      'username': Base64.encode(secret.userPass['username']),
      'password': Base64.encode(secret.userPass['password'])
    }
  } else if(secret.type == 'kubernetes.io/dockerconfigjson') {
    if(!secret.imagePass.url) {
      return `镜像仓库地址不能为空`
    }
    let auth = {auths: {}}
    auth.auths[secret.imagePass.url] = {
      'username': secret.imagePass.username,
      'password': secret.imagePass.password,
      'email': secret.imagePass.email,
      'auth': Base64.encode(`${secret.imagePass.username}:${secret.imagePass.password}`)
    }
    secret.data = {
      '.dockerconfigjson': Base64.encode(JSON.stringify(auth))
    }
  }
  delete secret.tls
  delete secret.userPass
  delete secret.imagePass
  return
}

export function resolveSecret(secret) {
  secret.tls = {}
  secret.userPass = {}
  secret.imagePass = {}
  let data = []
  if(["kubernetes.io/tls", "kubernetes.io/basic-auth", "kubernetes.io/dockerconfigjson"].indexOf(secret.type) == -1) {
    for(let k in secret.data) {
      data.push({key: k, value: Base64.decode(secret.data[k])})
    }
  } else if(secret.type == 'kubernetes.io/tls') {
    secret.tls['crt'] = Base64.decode(secret.data['tls.crt'])
    secret.tls['key'] = Base64.decode(secret.data['tls.key'])
  } else if(secret.type == 'kubernetes.io/basic-auth') {
    secret.userPass['username'] = Base64.decode(secret.data['username'])
    secret.userPass['password'] = Base64.decode(secret.data['password'])
  } else if(secret.type == 'kubernetes.io/dockerconfigjson') {
    let auths = JSON.parse(Base64.decode(secret.data['.dockerconfigjson']))
    for(let k in auths.auths) {
      secret.imagePass = {
        url: k,
        username: auths.auths[k].username,
        password: auths.auths[k].password,
        email: auths.auths[k].email
      }
    }
  }
  secret.data = data
}

export function buildEvent(event) {
  if (!event) return
  let eventTime = event.lastTimestamp
  if (!eventTime) eventTime = event.firstTimestamp
  if (!eventTime) eventTime = event.metadata.creationTimestamp
  return {
    uid: event.metadata.uid,
    namespace: event.metadata.namespace,
    count: event.spec ? event.spec.count : 1,
    reason: event.reason,
    message: event.message,
    type: event.type,
    object: event.involvedObject,
    source: event.source,
    event_time: eventTime,
    resource_version: event.metadata.resourceVersion,
  }
}