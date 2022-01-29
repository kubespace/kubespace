import request from '@/utils/request'

export function listCronJobs(cluster) {
  return request({
    url: `cronjob/${cluster}`,
    method: 'get',
  })
}

export function getCronJob(cluster, namespace, name, output='') {
  return request({
    url: `cronjob/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteCronJobs(cluster, data) {
  return request({
    url: `cronjob/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateCronJob(cluster, namespace, name, yaml) {
  return request({
    url: `cronjob/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateCronJobObj(cluster, namespace, name, data) {
  return request({
    url: `cronjob/${cluster}/${namespace}/${name}/update_obj`,
    method: 'post',
    data: data
  })
}
