import request from '@/utils/request'

export function listStoreApps(params) {
  return request({
    url: `/appstore`,
    method: 'get',
    params
  })
}

export function getStoreApp(id, params) {
  return request({
    url: `/appstore/${id}`,
    method: 'get',
    params
  })
}

export function createStoreApp(data) {
  return request({
    url: `/appstore/create`,
    method: 'post',
    data,
  })
}

export function deleteStoreAppVersion(appId, versionId) {
  return request({
    url: `/appstore/${appId}/${versionId}`,
    method: 'delete',
  })
}