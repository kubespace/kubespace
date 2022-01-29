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