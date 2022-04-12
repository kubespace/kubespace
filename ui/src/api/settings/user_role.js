import request from '@/utils/request'

export function listUserRole(params) {
  return request({
    url: '/user_role',
    method: 'get',
    params
  })
}

export function updateUserRole(data) {
  return request({
    url: `/user_role`,
    method: 'post',
    data,
  })
}

export function deleteUserRole(id) {
  return request({
    url: `/user_role/${id}`,
    method: 'delete',
  })
}