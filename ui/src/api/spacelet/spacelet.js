import request from '@/utils/request'

export function listSpacelet() {
  return request({
    url: '/spacelet',
    method: 'get',
  })
}