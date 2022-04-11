export { default as Release } from './release'
import { existsRelease } from "@/api/pipeline/workspace";

export async function manualCheck(pipeline, stage, stageJobParams) {
  let pluginMap = {
    'release': releaseManaualCheck,
  }
  for(let k in stageJobParams) {
    if(pluginMap[k]) {
      let err = pluginMap[k](pipeline, stage, stageJobParams[k])
      if(err) return err
    }
  }
  return
}

export async function releaseManaualCheck(pipeline, stage, jobParams) {
  let version = jobParams.version
  if(!version) {
    return '发布版本号不能为空'
  }
  let resp = await existsRelease({version: version, workspace_id: pipeline.workspace.id})
  if(resp.code != 'Success') {
    return resp.msg
  }
  if(resp.data.exists) {
    return '发布版本号已存在，请重新输入'
  }
  return
}