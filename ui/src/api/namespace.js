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
