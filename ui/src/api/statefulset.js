import request from '@/utils/request'

export function listStatefulSets(cluster) {
  return request({
    url: `statefulset/${cluster}`,
    method: 'get',
  })
}

export function getStatefulSet(cluster, namespace, name, output='') {
  return request({
    url: `statefulset/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteStatefulSets(cluster, data) {
  return request({
    url: `statefulset/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateStatefulSet(cluster, namespace, name, yaml) {
  return request({
    url: `statefulset/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateStatefulSetObj(cluster, namespace, name, data) {
  return request({
    url: `statefulset/${cluster}/update_obj/${namespace}/${name}`,
    method: 'post',
    data: data
  })
}
