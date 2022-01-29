import request from '@/utils/request'

export function listServiceAccounts(cluster) {
  return request({
    url: `serviceaccount/${cluster}`,
    method: 'get',
  })
}

export function getServiceAccount(cluster, namespace, name, output='') {
  return request({
    url: `serviceaccount/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteServiceAccounts(cluster, data) {
  return request({
    url: `serviceaccount/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateServiceAccount(cluster, namespace, name, yaml) {
  return request({
    url: `serviceaccount/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateServiceAccountObj(cluster, namespace, name, data) {
  return request({
    url: `serviceaccount/${cluster}/update_obj/${namespace}/${name}`,
    method: 'post',
    data: data
  })
}
