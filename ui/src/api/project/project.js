import request from '@/utils/request'

export function listProjects() {
  return request({
    url: `project/workspace`,
    method: 'get',
  })
}

export function getProject(id, params) {
  return request({
    url: `project/workspace/${id}`,
    method: 'get',
    params
  })
}

export function createProject(data) {
  return request({
    url: '/project/workspace',
    method: 'post',
    data,
  })
}

export function updateProject(id, data) {
  return request({
    url: `/project/workspace/${id}`,
    method: 'put',
    data,
  })
}

export function deleteProject(id) {
  return request({
    url: `/project/workspace/${id}`,
    method: 'delete',
  })
}

export function projectLabels() {
  return {
    "kubespace.cn/belong-to": "project"
  }
}