import request from '@/utils/request'

export function listSpacelet() {
  return request({
    url: '/spacelet',
    method: 'get',
  })
}

export function deleteSpacelet(id) {
  return request({
    url: `/spacelet/${id}`,
    method: 'delete',
  })
}