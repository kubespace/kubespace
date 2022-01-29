import request from '@/utils/request'

export function listSecret() {
  return request({
    url: '/settings/secret',
    method: 'get',
  })
}

export function createSecret(data) {
  return request({
    url: '/settings/secret',
    method: 'post',
    data,
  })
}

export function updateSecret(id, data) {
  return request({
    url: `/settings/secret/${id}`,
    method: 'put',
    data,
  })
}

export function deleteSecret(id) {
  return request({
    url: `/settings/secret/${id}`,
    method: 'delete',
  })
}