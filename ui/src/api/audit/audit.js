import request from '@/utils/request'

export function listAuditOperate(params) {
  return request({
    url: '/audit',
    method: 'get',
    params
  })
}