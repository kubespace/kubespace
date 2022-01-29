import request from '@/utils/request'

export function listIngresses(cluster) {
  return request({
    url: `ingress/${cluster}`,
    method: 'get',
  })
}

export function getIngress(cluster, namespace, name, output='') {
  return request({
    url: `ingress/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteIngresses(cluster, data) {
  return request({
    url: `ingress/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateIngress(cluster, namespace, name, yaml) {
  return request({
    url: `ingress/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateIngressObj(cluster, namespace, name, data) {
  return request({
    url: `ingress/${cluster}/update_obj/${namespace}/${name}`,
    method: 'post',
    data: data
  })
}
