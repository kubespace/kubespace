import request from '@/utils/request'
import store from '@/store'
import router from '@/router'

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
  console.log(router.app._route)
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

export function hasScopePermission(scope, scopeId, role) {
  let userInfo = store.getters.userInfo;
  if(userInfo.is_super) return true;

  let roleSets = {
    viewer: ['viewer', 'editor', 'admin'],
    editor: ['editor', 'admin'],
    admin: ['admin']
  }[role]
  console.log(userInfo.permissions)
  for (var perm of userInfo.permissions) {
    if(perm.scope == scope && perm.scope_id == scopeId && roleSets.indexOf(perm.role) >= 0){
      return true
    }
    if(perm.scope == "platform" && perm.scope_id == 0 && roleSets.indexOf(perm.role) >= 0){
      return true
    }
  }
  return false
}

export function projectScopeRole(role) {
  let meta = router.app._route.meta
  if(!meta) {
    return false
  }
  if(meta.group != 'workspace') return false
  let projectId = router.app._route.params.workspaceId
  if(!projectId) return platformScopeRole(role)
  if(isNaN(projectId)) return false
  return hasScopePermission('project', parseInt(projectId), role)
}

export function pipelineScopeRole(role) {
  let meta = router.app._route.meta
  if(!meta) {
    return false
  }
  if(meta.group != 'pipeline') return false
  let workspaceId = router.app._route.params.workspaceId
  if(!workspaceId) return platformScopeRole(role)
  if(isNaN(workspaceId)) return false
  return hasScopePermission('pipeline', parseInt(workspaceId), role)
}

export function clusterScopeRole(role) {
  let meta = router.app._route.meta
  if(!meta) {
    return false
  }
  if(meta.group != 'cluster') return false
  let clusterId = router.app._route.params.clusterId
  if(!clusterId) return platformScopeRole(role)
  if(isNaN(clusterId)) return false
  return hasScopePermission('cluster', parseInt(clusterId), role)
}

export function platformScopeRole(role) {
  return hasScopePermission('platform', 0, role)
}

export function viewerRole() {
  let meta = router.app._route.meta
  if(!meta) {
    return false
  }
  if(meta.group == 'workspace') return projectScopeRole("viewer")
  if(meta.group == 'pipeline') return pipelineScopeRole("viewer")
  if(meta.group == 'cluster') return clusterScopeRole("viewer")
  if(meta.group == 'settings') return platformScopeRole("viewer")
  return false
}

export function editorRole() {
  let meta = router.app._route.meta
  if(!meta) {
    return false
  }
  if(meta.group == 'workspace') return projectScopeRole("editor")
  if(meta.group == 'pipeline') return pipelineScopeRole("editor")
  if(meta.group == 'cluster') return clusterScopeRole("editor")
  if(meta.group == 'settings') return platformScopeRole("editor")
  return false
}

export function adminRole() {
  let meta = router.app._route.meta
  if(!meta) {
    return false
  }
  if(meta.group == 'workspace') return projectScopeRole("admin")
  if(meta.group == 'pipeline') return pipelineScopeRole("admin")
  if(meta.group == 'cluster') return clusterScopeRole("admin")
  if(meta.group == 'settings') return platformScopeRole("admin")
  return false
}