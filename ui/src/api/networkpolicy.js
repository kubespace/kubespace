import request from '@/utils/request'

export function listNetworkPolicies(cluster) {
  return request({
    url: `networkpolicy/${cluster}`,
    method: 'get',
  })
}

export function getNetworkPolicy(cluster, namespace, name, output='') {
  return request({
    url: `networkpolicy/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteNetworkPolicies(cluster, data) {
  return request({
    url: `networkpolicy/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateNetworkPolicy(cluster, namespace, name, yaml) {
  return request({
    url: `networkpolicy/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateNetworkPolicyObj(cluster, namespace, name, data) {
  return request({
    url: `networkpolicy/${cluster}/update_obj/${namespace}/${name}`,
    method: 'post',
    data: data
  })
}
