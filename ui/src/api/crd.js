import request from '@/utils/request'

export function listCrds(cluster) {
  return request({
    url: `crd/${cluster}`,
    method: 'get',
  })
}

export function getCrd(cluster, name, output='') {
  return request({
    url: `crd/${cluster}/${name}`,
    method: 'get',
    params: { output }
  })
}
