import request from '@/utils/request'

export function listNamespace(cluster) {
  return request({
    url: `namespace/${cluster}`,
    method: 'get',
  })
}

export function getNamespace(cluster, name, output) {
  return request({
    url: `namespace/${cluster}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteNamespaces(cluster, data) {
  return request({
    url: `namespace/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateNamespace(cluster, name, yaml) {
  return request({
    url: `namespace/${cluster}/update/${name}`,
    method: 'post',
    data: { yaml, name }
  })
}
