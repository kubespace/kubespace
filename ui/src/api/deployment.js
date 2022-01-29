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
