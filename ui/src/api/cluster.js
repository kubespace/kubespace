import request from '@/utils/request'
import qs from 'qs'

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

export function updateCluster(id, data) {
  return request({
    url: '/cluster/' + id,
    method: 'put',
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

export function updateGvr(cluster, yaml) {
  return request({
    url: `/cluster/updateGvr/${cluster}`,
    method: 'post',
    data: {yaml: yaml},
  })
}

export function sse(vueSse, watchFunc, cluster, params) {
  let p = qs.stringify(params)
  let url = `/api/v1/cluster/${cluster}/sse?${p}`
  let clusterSSE = vueSse.create({
    url: url,
    includeCredentials: false,
    format: 'plain'
  });
  clusterSSE.on("message", (res) => {
    // console.log(res)
    if(res && res != "\n") {
      let data = JSON.parse(res)
      console.log(data)
      watchFunc(data)
    }
  })
  clusterSSE.connect().then(() => {
    console.log('[info] connected', 'system')
  }).catch(() => {
    console.log('[error] failed to connect', 'system')
  })
  clusterSSE.on('error', () => { // eslint-disable-line
    console.log('[error] disconnected, automatically re-attempting connection', 'system')
  })
  return clusterSSE
}
