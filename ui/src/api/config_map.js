import request from '@/utils/request'

export function listConfigMaps(cluster) {
  return request({
    url: `configmap/${cluster}`,
    method: 'get',
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