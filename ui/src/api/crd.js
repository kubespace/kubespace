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

export function listCrs(cluster, params) {
  return request({
    url: `crd/${cluster}/cr`,
    method: 'get',
    params
  })
}

export function getCr(cluster, name, params) {
  return request({
    url: `crd/${cluster}/cr/${name}`,
    method: 'get',
    params
  })
}

export function deleteCr(cluster, name, params) {
  return request({
    url: `crd/${cluster}/cr/${name}`,
    method: 'delete',
    data: params
  })
}
