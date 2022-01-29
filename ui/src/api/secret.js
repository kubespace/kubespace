import request from '@/utils/request'

export function listSecrets(cluster) {
  return request({
    url: `secret/${cluster}`,
    method: 'get',
  })
}

export function getSecret(cluster, namespace, name, output='') {
  return request({
    url: `secret/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function updateSecret(cluster, namespace, name, yaml) {
  return request({
    url: `secret/${cluster}/update_secret`,
    method: 'post',
    data: { yaml, name, namespace }
  })
}