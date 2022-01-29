import request from '@/utils/request'

export function listServices(cluster) {
  return request({
    url: `service/${cluster}`,
    method: 'get',
  })
}

export function getService(cluster, namespace, name, output='') {
  return request({
    url: `service/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteServices(cluster, data) {
  return request({
    url: `service/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateService(cluster, namespace, name, yaml) {
  return request({
    url: `service/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateServiceObj(cluster, namespace, name, data) {
  return request({
    url: `service/${cluster}/update_obj/${namespace}/${name}`,
    method: 'post',
    data: data
  })
}
