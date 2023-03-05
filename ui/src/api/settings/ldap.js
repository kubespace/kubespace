import request from '@/utils/request'

export function listLdap() {
  return request({
    url: '/settings/ldap',
    method: 'get',
  })
}

export function createLdap(data) {
  return request({
    url: '/settings/ldap',
    method: 'post',
    data,
  })
}

export function updateLdap(id, data) {
  return request({
    url: `/settings/ldap/${id}`,
    method: 'put',
    data,
  })
}

export function deleteLdap(id) {
  return request({
    url: `/settings/ldap/${id}`,
    method: 'delete',
  })
}