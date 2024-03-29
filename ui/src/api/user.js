import request from '@/utils/request'

export function login(data) {
  return request({
    url: '/user/login',
    method: 'post',
    data
  })
}

export function getInfo(token) {
  return request({
    url: '/user/token',
    method: 'get',
    params: { token }
  })
}

export function logout() {
  return request({
    url: '/user/logout',
    method: 'post'
  })
}

export function adminSet(data) {
  return request({
    url: '/user/admin',
    method: 'post',
    data
  })
}

export function hasAdmin() {
  return request({
    url: '/user/has_admin',
    method: 'get',
  })
}

export function createUser(data) {
  return request({
    url: '/user',
    method: 'post',
    data
  })
}

export function getUser(data) {
  return request({
    url: '/user',
    method: 'get',
    data
  })
}

export function updateUser(pk, data) {
  return request({
    url: `/user/${pk}`,
    method: 'put',
    data
  })
}

export function deleteUser(users) {
  return request({
    url: `/user/delete`,
    method: 'post',
    data: users,
  })
}

export function updatePassword(data) {
  return request({
    url: `/user/update_password`,
    method: 'post',
    data,
  })
}
