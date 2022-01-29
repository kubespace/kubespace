import request from '@/utils/request'

export function listBuilds(pipeline_id) {
  return request({
    url: `pipeline/build/list`,
    method: 'get',
    params: {'pipeline_id': pipeline_id}
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