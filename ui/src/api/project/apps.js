import request from '@/utils/request'

export function listApps(params) {
  return request({
    url: `project/apps`,
    method: 'get',
    params
  })
}

export function listAppStatus(params) {
  return request({
    url: `project/apps/status`,
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

export function getApp(id) {
  return request({
    url: `project/apps/${id}`,
    method: 'get',
  })
}

export function deleteApp(id) {
  return request({
    url: `project/apps/${id}`,
    method: 'delete',
  })
}

export function getAppVersion(id) {
  return request({
    url: `project/apps/version/${id}`,
    method: 'get',
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

export function importStoreApp(data) {
  return request({
    url: `/project/apps/import_storeapp`,
    method: 'post',
    data,
  })
}

export function importCustomApp(data) {
  return request({
    url: `/project/apps/import_custom_app`,
    method: 'post',
    data,
  })
}

export function duplicateApp(data) {
  return request({
    url: `/project/apps/duplicate_app`,
    method: 'post',
    data,
  })
}

export function deleteAppVersion(id) {
  return request({
    url: `project/apps/version/${id}`,
    method: 'delete',
  })
}

export function downloadChart(path) {
  return request({
    url: `project/apps/download`,
    method: 'get',
    params: {path}
  })
}

export function getAppChartFiles(versionId) {
  return request({
    url: `project/apps/version/${versionId}/chartfiles`,
    method: 'get'
  })
}