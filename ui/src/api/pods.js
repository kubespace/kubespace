import request from '@/utils/request';
import { Selector } from 'k8s-selector';

export function podMatch(selector, labels) {
  let s = Selector(selector)
  console.log(s, selector, labels)
  return s(labels)
}

export function listPods(cluster, label_selector=null, names=null) {
  let data = {}
  if (label_selector) data['label_selector'] = label_selector
  if (names) data['names'] = names
  return request({
    url: `pods/${cluster}/list`,
    method: 'post',
    data,
  })
}

export function getPod(cluster, namespace, name, output='') {
  return request({
    url: `pods/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deletePods(cluster, data) {
  return request({
    url: `pods/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updatePod(cluster, namespace, name, yaml) {
  return request({
    url: `pods/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function containerClass(status) {
  if (status === 'running') return 'running-class'
  if (status === 'terminated') return 'terminate-class'
  if (status === 'waiting') return 'waiting-class'
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
    created: pod.metadata.creationTimestamp,
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
