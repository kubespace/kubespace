import request from '@/utils/request'

export function listDeployments(cluster) {
  return request({
    url: `deployment/${cluster}`,
    method: 'get',
  })
}

export function getDeployment(cluster, namespace, name, output='') {
  return request({
    url: `deployment/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteDeployments(cluster, data) {
  return request({
    url: `deployment/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateDeployment(cluster, namespace, name, yaml) {
  return request({
    url: `deployment/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateDeploymentObj(cluster, namespace, name, data) {
  return request({
    url: `deployment/${cluster}/update_obj/${namespace}/${name}`,
    method: 'post',
    data: data
  })
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
