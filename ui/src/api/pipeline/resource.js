import request from '@/utils/request'

export function listResources(workspaceId) {
  return request({
    url: `pipeline/resource/${workspaceId}`,
    method: 'get',
  })
}

export function createResource(data) {
  return request({
    url: '/pipeline/resource',
    method: 'post',
    data,
  })
}

export function updateResource(id, data) {
  return request({
    url: `/pipeline/resource/${id}`,
    method: 'put',
    data,
  })
}

export function deleteResource(id) {
  return request({
    url: `/pipeline/resource/${id}`,
    method: 'delete',
  })
}