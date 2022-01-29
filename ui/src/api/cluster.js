import request from '@/utils/request'

export function listCluster() {
  return request({
    url: '/cluster',
    method: 'get',
  })
}

export function createCluster(data) {
  return request({
    url: '/cluster',
    method: 'post',
    data,
  })
}

export function clusterMembers(data) {
  return request({
    url: '/cluster/members',
    method: 'post',
    data,
  })
}

export function clusterDetail(cluster) {
  return request({
    url: `/cluster/${cluster}/detail`,
    method: 'get',
  })
}

export function deleteCluster(clusters) {
  return request({
    url: `/cluster/delete`,
    method: 'post',
    data: clusters,
  })
}

export function applyYaml(cluster, yaml) {
  return request({
    url: `/cluster/apply/${cluster}`,
    method: 'post',
    data: {yaml: yaml},
  })
}

export function createYaml(cluster, yaml) {
  return request({
    url: `/cluster/createYaml/${cluster}`,
    method: 'post',
    data: {yaml: yaml},
  })
}
