import request from '@/utils/request'

export function listProjects() {
  return request({
    url: `/project`,
    method: 'get',
  })
}

export function getProject(id, params) {
  return request({
    url: `/project/${id}`,
    method: 'get',
    params
  })
}

export function getProjectResources(projectId) {
  return request({
    url: `/project/resources`,
    method: 'get',
    params: {project_id: projectId}
  })
}

export function createProject(data) {
  return request({
    url: '/project',
    method: 'post',
    data,
  })
}

export function cloneProject(data) {
  return request({
    url: '/project/clone',
    method: 'post',
    data,
  })
}

export function updateProject(id, data) {
  return request({
    url: `/project/${id}`,
    method: 'put',
    data,
  })
}

export function deleteProject(id, data) {
  return request({
    url: `/project/${id}`,
    method: 'delete',
    data
  })
}

export function projectLabels() {
  return {
    "kubespace.cn/belong-to": "project"
  }
}