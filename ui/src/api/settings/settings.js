import request from '@/utils/request'

export function globalSettings() {
  return request({
    url: '/settings/global',
    method: 'get',
  })
}