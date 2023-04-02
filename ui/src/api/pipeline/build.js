import request from '@/utils/request'

export function listBuilds(pipeline_id, last_build_number, status, limit) {
  return request({
    url: `pipeline/build/list`,
    method: 'get',
    params: {'pipeline_id': pipeline_id, last_build_number: last_build_number, status, limit}
  })
}

export function getBuild(build_id) {
  return request({
    url: `pipeline/build/${build_id}`,
    method: 'get',
  })
}

export function buildPipeline(data) {
  return request({
    url: `pipeline/build`,
    method: 'post',
    data: data
  })
}

export function getJobLog(job_id, with_sse) {
  return request({
    url: `pipeline/build/log/${job_id}${with_sse? "/sse" : ""}`,
    method: 'get',
  })
}

export function manualExec(data) {
  return request({
    url: `pipeline/build/manual_execute`,
    method: 'post',
    data,
  })
}

export function stageRetry(data) {
  return request({
    url: `pipeline/build/retry`,
    method: 'post',
    data,
  })
}

export function stageCancel(data) {
  return request({
    url: `pipeline/build/cancel`,
    method: 'post',
    data,
  })
}

export function stageReexec(data) {
  return request({
    url: `pipeline/build/reexec`,
    method: 'post',
    data,
  })
}