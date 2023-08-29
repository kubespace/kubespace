import request from '@/utils/request'

export function listPipelines(workspaceId) {
  return request({
    url: `pipeline/pipeline`,
    method: 'get',
    params: {'workspace_id': workspaceId}
  })
}

export function getPipeline(pipelineId) {
  return request({
    url: `pipeline/pipeline/${pipelineId}`,
    method: 'get',
  })
}

export function updatePipeline(id, data) {
  return request({
    url: `pipeline/pipeline/${id}`,
    method: 'put',
    data: data,
  })
}

export function createPipeline(data) {
  return request({
    url: `pipeline/pipeline`,
    method: 'post',
    data: data,
  })
}

export function deletePipeline(id) {
  return request({
    url: `pipeline/pipeline/${id}`,
    method: 'delete',
  })
}

export function listRepoBranches(id) {
  return request({
    url: `pipeline/pipeline/${id}/list_repo_branch`,
    method: 'get',
  })
}