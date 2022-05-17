import request from '@/utils/request'

export function listDaemonSets(cluster) {
  return request({
    url: `daemonset/${cluster}`,
    method: 'get',
  })
}

export function getDaemonSet(cluster, namespace, name, output='') {
  return request({
    url: `daemonset/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteDaemonSets(cluster, data) {
  return request({
    url: `daemonset/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateDaemonSet(cluster, namespace, name, yaml) {
  return request({
    url: `daemonset/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateDaemonSetObj(cluster, namespace, name, data) {
  return request({
    url: `daemonset/${cluster}/${namespace}/${name}/update_obj`,
    method: 'post',
    data: data
  })
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
