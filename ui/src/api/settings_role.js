import request from '@/utils/request'
import store from '@/store'
import router from '@/router'

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

export function hasPermission(scope, object, operation) {
  let userInfo = store.getters.userInfo;
  if(userInfo.is_super) return true;

  for (var perm of userInfo.permissions) {
    if(perm.scope == scope && perm.object == object && perm.operations.indexOf(operation) >= 0){
      return true
    }
  }
  return false
}

export function podOpPerm(operation) {
  return hasPermission('cluster', 'pod', operation)
}

export function getPerm() {
  let meta = router.app._route.meta
  if(!meta) {
    return false
  }
  return hasPermission(meta.group, meta.object, 'get')
}

export function createPerm() {
  let meta = router.app._route.meta
  if(!meta) {
    return false
  }
  return hasPermission(meta.group, meta.object, 'create')
}

export function updatePerm() {
  let meta = router.app._route.meta
  if(!meta) {
    return false
  }
  return hasPermission(meta.group, meta.object, 'update')
}

export function deletePerm() {
  let meta = router.app._route.meta
  if(!meta) {
    return false
  }
  return hasPermission(meta.group, meta.object, 'delete')
}
