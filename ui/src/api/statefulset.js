import request from '@/utils/request'

export function listStatefulSets(cluster) {
  return request({
    url: `statefulset/${cluster}`,
    method: 'get',
  })
}

export function getStatefulSet(cluster, namespace, name, output='') {
  return request({
    url: `statefulset/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteStatefulSets(cluster, data) {
  return request({
    url: `statefulset/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateStatefulSet(cluster, namespace, name, yaml) {
  return request({
    url: `statefulset/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateStatefulSetObj(cluster, namespace, name, data) {
  return request({
    url: `statefulset/${cluster}/update_obj/${namespace}/${name}`,
    method: 'post',
    data: data
  })
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
