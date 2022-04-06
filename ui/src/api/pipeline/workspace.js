import request from '@/utils/request'

export function listWorkspaces() {
  return request({
    url: `pipeline/workspace`,
    method: 'get',
  })
}

export function getWorkspace(id) {
  return request({
    url: `pipeline/workspace/${id}`,
    method: 'get',
  })
}

export function createWorkspace(data) {
  return request({
    url: '/pipeline/workspace',
    method: 'post',
    data,
  })
}

export function updateWorkspace(id, data) {
  return request({
    url: `/pipeline/workspace/${id}`,
    method: 'put',
    data,
  })
}

export function deleteWorkspace(id) {
  return request({
    url: `/pipeline/workspace/${id}`,
    method: 'delete',
  })
}