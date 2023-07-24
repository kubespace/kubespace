export { default as PipelineStage } from './stage'
export { default as CodeToImage } from './codeToImage'
export { default as ExecuteShell } from './execShell'
export { default as AppDeploy } from './appDeploy'
export { default as Release } from './release'
export { default as DeployK8s } from './deployK8s'

export function checkPluginJob(job) {
  if(!job) {
    return {checked: false, errorMsg: "参数错误"}
  }
  if(!job.plugin_key) {
    return {checked: false, errorMsg: "请选择要执行的任务插件"}
  }
  let res
  if(job.plugin_key == "build_code_to_image") {
    res = checkCodeToImageJob(job)
  }
  if(job.plugin_key == "execute_shell") {
    res = checkExecShell(job)
  }
  if(job.plugin_key == "deploy_k8s") {
    res = checkDeployK8s(job)
  }
  if(job.plugin_key == "upgrade_app") {
    res = checkAppDeploy(job)
  }
  if(job.plugin_key == "release") {
    res = checkRelease(job)
  }

  if(!res) {
    return {checked: false, errorMsg: "任务插件参数错误"}
  }
  return res
}

function checkCodeToImageJob(job) {
  if(!job.params) {
    return {checked: false, errorMsg: "参数错误"}
  }
  let jobParams = job.params
  if(jobParams.code_build && jobParams.code_build_type == 'script') {
    if(!jobParams.code_build_script) {
      return {checked: false, errorMsg: "编译脚本不能为空"}
    }
  }
  if(!jobParams.image_build_registry) {
    return {checked: false, errorMsg: "镜像构建仓库不能为空"}
  }
  if(!jobParams.image_builds || jobParams.image_builds.length == 0) {
    return {checked: false, errorMsg: "镜像构建列表不能为空"}
  }
  for(let b of jobParams.image_builds) {
    if(!b.image) {
      return {checked: false, errorMsg: "镜像构建名称不能为空"}
    }
  }
  return {checked: true}
}

function checkDeployK8s(job) {
  if(!job.params) {
    return {checked: false, errorMsg: "参数错误"}
  }
  if(!job.params.cluster) {
    return {checked: false, errorMsg: "部署集群不能为空"}
  }
  if(!job.params.namespace) {
    return {checked: false, errorMsg: "命名空间不能为空"}
  }
  if(!job.params.yaml) {
    return {checked: false, errorMsg: "Yaml内容不能为空"}
  }
  return {checked: true}
}

function checkAppDeploy(job) {
  if(!job.params) {
    return {checked: false, errorMsg: "参数错误"}
  }
  if(!job.params.project) {
    return {checked: false, errorMsg: "工作空间不能为空"}
  }
  if(!job.params.apps || job.params.apps.length == 0) {
    return {checked: false, errorMsg: "应用不能为空"}
  }
  return {checked: true}
}

function checkExecShell(job) {
  if(!job.params) {
    return {checked: false, errorMsg: "参数错误"}
  }
  if(!job.params.resource) {
    return {checked: false, errorMsg: "执行脚本目标资源不能为空"}
  }
  if(!job.params.script) {
    return {checked: false, errorMsg: "执行脚本不能为空"}
  }
  return {checked: true}
}

function checkRelease(job) {
  return {checked: true}
}
