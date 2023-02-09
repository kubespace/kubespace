import request from '@/utils/request'

export function listWorkspaces(params) {
  return request({
    url: `pipeline/workspace`,
    method: 'get',
    params
  })
}

export function getWorkspace(id) {
  return request({
    url: `pipeline/workspace/${id}`,
    method: 'get',
  })
}

export function createWorkspace(data) {
  return request({
    url: '/pipeline/workspace',
    method: 'post',
    data,
  })
}

export function updateWorkspace(id, data) {
  return request({
    url: `/pipeline/workspace/${id}`,
    method: 'put',
    data,
  })
}

export function deleteWorkspace(id) {
  return request({
    url: `/pipeline/workspace/${id}`,
    method: 'delete',
  })
}

export function getLatestRelease(params) {
  return request({
    url: `pipeline/workspace/latest_release`,
    method: 'get',
    params
  })
}

export function listGitRepos(params) {
  return request({
    url: `pipeline/workspace/list_git_repos`,
    method: 'get',
    params
  })
}

export async function existsRelease(params) {
  let data = {}
  await request({
    url: `pipeline/workspace/exists_release`,
    method: 'get',
    params
  }).then(function(resp) {
    data = resp
  }).catch(function(e) {
    console.log(e)
    data = {Code: 'error', Msg: e}
  })
  return data
}