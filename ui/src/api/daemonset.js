import request from '@/utils/request'

export function listDaemonSets(cluster) {
  return request({
    url: `daemonset/${cluster}`,
    method: 'get',
  })
}

export function getDaemonSet(cluster, namespace, name, output='') {
  return request({
    url: `daemonset/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteDaemonSets(cluster, data) {
  return request({
    url: `daemonset/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateDaemonSet(cluster, namespace, name, yaml) {
  return request({
    url: `daemonset/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateDaemonSetObj(cluster, namespace, name, data) {
  return request({
    url: `daemonset/${cluster}/${namespace}/${name}/update_obj`,
    method: 'post',
    data: data
  })
}
