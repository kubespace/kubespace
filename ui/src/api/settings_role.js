import request from '@/utils/request'

export function getPermissions() {
  return request({
    url: '/settings_role/permissions',
    method: 'get',
  })
}

export function getRoles(data) {
  return request({
    url: '/settings_role/',
    method: 'get',
    data
  })
}

export function createRole(role) {
  return request({
    url: '/settings_role/',
    method: 'post',
    data: role
  })
}

export function updateRole(name, data) {
  return request({
    url: `/settings_role/${name}`,
    method: 'put',
    data
  })
}

export function deleteRoles(data) {
  return request({
    url: '/settings_role/delete',
    method: 'post',
    data
  })
}


