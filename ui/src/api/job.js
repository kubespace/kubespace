import request from '@/utils/request'

export function listJobs(cluster, cronjob_uid) {
  return request({
    url: `job/${cluster}`,
    method: 'get',
    params: {cronjob_uid}
  })
}

export function getJob(cluster, namespace, name, output='') {
  return request({
    url: `job/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function deleteJobs(cluster, data) {
  return request({
    url: `job/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function updateJob(cluster, namespace, name, yaml) {
  return request({
    url: `job/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function updateJobObj(cluster, namespace, name, data) {
  return request({
    url: `job/${cluster}/${namespace}/${name}/update_obj`,
    method: 'post',
    data: data
  })
}

export function buildJobs(job) {
  if (!job) return
  var conditions = []
  if(job.status.conditions) {
    for (let c of job.status.conditions) {
      if (c.status === "True") {
        conditions.push(c.type)
      }
    }
  }
  let p = {
    uid: job.metadata.uid,
    namespace: job.metadata.namespace,
    name: job.metadata.name,
    completions: job.spec.completions || 0,
    active: job.status.active || 0,
    succeeded: job.status.succeeded || 0,
    failed: job.status.failed || 0,
    resource_version: job.metadata.resourceVersion,
    conditions: conditions,
    node_selector: job.spec.template.spec.nodeSelector,
    created: job.metadata.creationTimestamp
  }
  return p
}
