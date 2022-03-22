import request from '@/utils/request'

export function listConfigMaps(cluster, params) {
  return request({
    url: `configmap/${cluster}`,
    method: 'get',
    params
  })
}

export function getConfigMap(cluster, namespace, name, output='') {
  return request({
    url: `configmap/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function updateConfigMap(cluster, namespace, name, yaml) {
  return request({
    url: `configmap/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml, name, namespace }
  })
}

export function deleteConfigMaps(cluster, data) {
  return request({
    url: `configmap/${cluster}/delete`,
    method: 'post',
    data: data
  })
}
