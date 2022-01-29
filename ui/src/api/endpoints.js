import request from '@/utils/request'

export function listEndpoints(cluster, namespace=null, name=null) {
  let params = {}
  if (namespace) params['namespace'] = namespace
  if (name) params['name'] = name
  return request({
    url: `endpoints/${cluster}`,
    method: 'get',
    params: params,
  })
}

export function getEndpoints(cluster, namespace, name, output='') {
  return request({
    url: `endpoints/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteEndpoints(cluster, data) {
  return request({
    url: `endpoints/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateEndpoints(cluster, namespace, name, yaml) {
  return request({
    url: `endpoints/${cluster}/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateEndpointsObj(cluster, namespace, name, data) {
  return request({
    url: `endpoints/${cluster}/${namespace}/${name}/update_obj`,
    method: 'post',
    data: data
  })
}
