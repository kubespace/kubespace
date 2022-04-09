import request from '@/utils/request'

export function listBuilds(pipeline_id, last_build_number) {
  return request({
    url: `pipeline/build/list`,
    method: 'get',
    params: {'pipeline_id': pipeline_id, last_build_number: last_build_number}
  })
}

export function getBuild(build_id) {
  return request({
    url: `pipeline/build/${build_id}`,
    method: 'get',
  })
}

export function buildPipeline(pipeline_id, params) {
  return request({
    url: `pipeline/build`,
    method: 'post',
    data: {'pipeline_id': parseInt(pipeline_id), 'params': params}
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