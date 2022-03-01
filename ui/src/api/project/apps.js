import request from '@/utils/request'

export function listApps(params) {
  return request({
    url: `project/apps`,
    method: 'get',
    params
  })
}

export function listAppVersions(params) {
  return request({
    url: `project/apps/versions`,
    method: 'get',
    params
  })
}

export function createApp(data) {
  return request({
    url: '/project/apps',
    method: 'post',
    data,
  })
}

export function installApp(data) {
  return request({
    url: '/project/apps/install',
    method: 'post',
    data,
  })
}

export function destroyApp(data) {
  return request({
    url: '/project/apps/destroy',
    method: 'post',
    data,
  })
}

export function updateApp(id, data) {
  return request({
    url: `/project/apps/${id}`,
    method: 'put',
    data,
  })
}

export function deleteApp(id) {
  return request({
    url: `/project/apps/${id}`,
    method: 'delete',
  })
}