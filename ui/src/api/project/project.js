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

export function getProjectResources(projectId) {
  return request({
    url: `project/workspace/resources`,
    method: 'get',
    params: {project_id: projectId}
  })
}

export function createProject(data) {
  return request({
    url: '/project/workspace',
    method: 'post',
    data,
  })
}

export function cloneProject(data) {
  return request({
    url: '/project/workspace/clone',
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

export function deleteProject(id, data) {
  return request({
    url: `/project/workspace/${id}`,
    method: 'delete',
    data
  })
}

export function projectLabels() {
  return {
    "kubespace.cn/belong-to": "project"
  }
}