import request from '@/utils/request'

export function listRoleBindings(cluster) {
  return request({
    url: `rolebinding/${cluster}`,
    method: 'get',
  })
}

export function getRoleBinding(cluster, namespace, name, kind, output='') {
  return request({
    url: `rolebinding/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output, kind }
  })
}

export function deleteRoleBindings(cluster, data) {
  return request({
    url: `rolebinding/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateRoleBinding(cluster, namespace, name, kind, yaml) {
  return request({
    url: `rolebinding/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml, kind }
  })
}

export function updateRoleBindingObj(cluster, namespace, name, data) {
  return request({
    url: `rolebinding/${cluster}/update_obj/${namespace}/${name}`,
    method: 'post',
    data: data
  })
}
