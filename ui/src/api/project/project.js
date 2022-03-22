import request from '@/utils/request'

export function listProjects() {
  return request({
    url: `project/workspace`,
    method: 'get',
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
    url: `/pipeline/workspace/${id}`,
    method: 'put',
    data,
  })
}

export function deleteProject(id) {
  return request({
    url: `/pipeline/workspace/${id}`,
    method: 'delete',
  })
}

export function projectLabels() {
  return {
    "kubespace.cn/belong-to": "project"
  }
}