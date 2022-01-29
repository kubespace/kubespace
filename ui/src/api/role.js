import request from '@/utils/request'

export function listRoles(cluster) {
  return request({
    url: `role/${cluster}`,
    method: 'get',
  })
}

export function getRole(cluster, namespace, name, kind, output='') {
  return request({
    url: `role/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output, kind }
  })
}

export function deleteRoles(cluster, data) {
  return request({
    url: `role/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateRole(cluster, namespace, name, yaml) {
  return request({
    url: `role/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateRoleObj(cluster, namespace, name, data) {
  return request({
    url: `role/${cluster}/update_obj/${namespace}/${name}`,
    method: 'post',
    data: data
  })
}
