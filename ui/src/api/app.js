import request from '@/utils/request'

export function listReleases(cluster) {
  return request({
    url: `helm/release/${cluster}`,
    method: 'get',
  })
}

export function getRelease(cluster, params) {
  return request({
    url: `helm/release/${cluster}/get`,
    method: 'get',
    params,
  })
}

export function listApps() {
  return request({
    url: `helm/app/list`,
    method: 'get',
  })
}

export function getApp(name, chart_version) {
  return request({
    url: `helm/app/get`,
    method: 'get',
    params: { name, chart_version }
  })
}

export function createApp(cluster, data) {
  return request({
    url: `helm/release/${cluster}`,
    method: 'post',
    data
  })
}

export function deleteRelease(cluster, params) {
  return request({
    url: `helm/release/${cluster}`,
    method: 'delete',
    params,
  })
}
