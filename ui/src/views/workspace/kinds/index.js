import { transferSecret, resolveSecret } from '@/api/secret'
export { default as HealthProbe } from './HealthProbe'
export { default as Container } from './container'
export { default as PodVolume } from './pod_volume'
export { default as PodNetwork } from './pod_network'
export { default as PodAffinity } from './pod_affinity'
export { default as PodSecurity } from './pod_security'
export { default as Workload } from './workload'
export { default as Service } from './service'
export { default as ConfigMap } from './configmap'
export { default as Secret } from './secret'
export { default as pvc } from './pvc'

export function kindTemplate(kind) {
  if(kind == 'Workload') return workloadTemplate()
  else if(kind == 'Service') return serviceTemplate()
  else if(kind == 'ConfigMap') return configMapTemplate()
  else if(kind == 'Secret') return secretTemplate()
  else if(kind == 'PersistentVolumeClaim') return pvcTemplate()
}

export function transferTemplate(template, appName) {
  let tpl = JSON.parse(JSON.stringify(template))

  if(!tpl) return {err: "应用资源为空"}
  if(!tpl.kind) return {err: "应用资源kind为空"}
  if(!tpl.metadata) return {err: "应用资源metadata为空"}
  if(!tpl.metadata.name) return {err: "应用资源名称为空"}

  tpl.metadata.labels['kubespace.cn/app'] = appName

  if(['Deployment', 'StatefulSet', 'DaemonSet', 'CronJob', "Job"].indexOf(tpl.kind) > -1) return transferWorkload(tpl)
  else if(tpl.kind == 'Service') return transferService(tpl)
  else if(tpl.kind == 'ConfigMap') return transferConfigMap(tpl)
  else if(tpl.kind == 'Secret') return toTransferSecret(tpl)
  else if(tpl.kind == 'PersistentVolumeClaim') return transferPvc(tpl)

  return {err: `${tpl.kind}/${tpl.metadata.name}未找到对应的资源类型`}
}

function transferWorkload(tpl) {
  tpl.apiVersion = 'apps/v1'
  if(!tpl.spec.template.spec.containers || tpl.spec.template.spec.containers.length <= 0) {
    return {err: `应用资源${tpl.kind}/${tpl.metadata.name}容器为空`}
  }
  if(tpl.kind != 'Job' && tpl.kind != 'CronJob') {
    tpl.spec.selector.matchLabels['kubespace.cn/app'] = tpl.metadata.name
    tpl.spec.template.metadata.labels['kubespace.cn/app'] = tpl.metadata.name
  } else {
    if(tpl.spec.selector != undefined) {
      delete tpl.spec.selector
    }
    if(tpl.spec.template.metadata != undefined) {
      delete tpl.spec.template.metadata
    }
  }
  if(tpl.kind == 'Deployment') {
    let strategy = {}
    if(tpl.spec.strategy.type == 'RollingUpdate') {
      strategy = {
        type: 'RollingUpdate',
        rollingUpdate: {
          maxUnavailable: tpl.spec.strategy.maxUnavailable,
          maxSurge: tpl.spec.strategy.maxSurge,
        }
      }
    } else {
      strategy = {type: 'Recreate'}
    }
    tpl.spec.strategy = strategy
  } else if(tpl.spec.strategy != undefined) {
    delete tpl.spec.strategy
  }
  if(tpl.kind != 'Deployment' && tpl.kind != 'StatefulSet') {
    delete tpl.spec.replicas
  }
  let err = transferContainer(tpl)
  if(err) return err
  err = transferPodVolume(tpl)
  if(err) return err
  err = transferPodNetwork(tpl)
  if(err) return err
  err = transferAffinity(tpl)
  if(err) return err
  if(tpl.kind != 'CronJob' && tpl.kind != "Job") {
    if(tpl.spec.job != undefined) {
      delete tpl.spec.job
    }
    if(tpl.spec.template.spec.restartPolicy) {
      delete tpl.spec.template.spec.restartPolicy
    }
  } else{
    tpl.spec.template.spec.restartPolicy = 'OnFailure'
    if(tpl.kind == 'Job') {
      err = transferJob(tpl)
      if(err) return err
    } else if(tpl.kind == 'CronJob') {
      err = transferCronJob(tpl)
      if(err) return err
    }
  }
  return {tpl}
}

function transferJob(tpl) {
  if(tpl.kind != 'Job') return
  tpl.apiVersion = 'batch/v1beta1'
  if(!tpl.spec.job) return
  if(tpl.spec.job.backoffLimit) {
    tpl.spec.backoffLimit = tpl.spec.job.backoffLimit
  }
  if(tpl.spec.job.completions) {
    tpl.spec.completions = tpl.spec.job.completions
  }
  if(tpl.spec.job.parallelism) {
    tpl.spec.parallelism = tpl.spec.job.parallelism
  }
  delete tpl.spec.job
}

function transferCronJob(tpl) {
  if(tpl.kind != 'CronJob') return
  tpl.apiVersion = 'batch/v1beta1'
  let spec = {
    jobTemplate: {
      spec: {}
    }
  }
  if(tpl.spec.job) {
    if(tpl.spec.job.backoffLimit) {
      tpl.spec.backoffLimit = tpl.spec.job.backoffLimit
    }
    if(tpl.spec.job.completions) {
      tpl.spec.completions = tpl.spec.job.completions
    }
    if(tpl.spec.job.parallelism) {
      tpl.spec.parallelism = tpl.spec.job.parallelism
    }
    if(tpl.spec.job.schedule) {
      spec.schedule = tpl.spec.job.schedule
    }
    if(tpl.spec.job.concurrencyPolicy) {
      spec.concurrencyPolicy = tpl.spec.job.concurrencyPolicy
    }
  }
  if(!spec.schedule) {
    return {err: `应用资源${tpl.kind}/${tpl.metadata.name}定时为空`}
  }
  delete tpl.spec.job
  tpl.spec.template.metadata = {
    'labels': {
      'kubespace.cn/app': tpl.metadata.name
    }
  }
  spec.jobTemplate.spec = tpl.spec
  tpl.spec = spec
  return
}

function transferPodNetwork(tpl) {
  if(tpl.spec.template.spec.hostAliases.length > 0) {
    for(let h of tpl.spec.template.spec.hostAliases) {
      if(!h.hostnames) return {err: `应用资源${tpl.kind}/${tpl.metadata.name}主机别名域名为空`}
      if(!h.ip) return {err: `应用资源${tpl.kind}/${tpl.metadata.name}主机别名ip为空`}
      h.hostnames = [h.hostnames]
    }
  }
}

function transferContainer(tpl) {
  let initContainers = []
  let containers = []
  let err = ''
  for(let c of tpl.spec.template.spec.containers) {
    if(!c.name) {
      return {err: `应用资源${tpl.kind}/${tpl.metadata.name}容器名称为空`}
    }
    if(!c.image) {
      return {err: `应用资源${tpl.kind}/${tpl.metadata.name}中容器镜像为空`}
    }
    let p = transferProbe(c.livenessProbe)
    if(p[1]) {
      return {err: `应用资源${tpl.kind}/${tpl.metadata.name}中容器${p[1]}`}
    }
    if(p[0]) {
      c.livenessProbe = p[0]
    } else {
      delete c.livenessProbe
    }
    p = transferProbe(c.readinessProbe)
    if(p[1]) {
      return {err: `应用资源${tpl.kind}/${tpl.metadata.name}中容器${p[1]}`}
    }
    if(p[0]) {
      c.readinessProbe = p[0]
    } else {
      delete c.readinessProbe
    }
    if(c.command) {
      try{
        c.command = JSON.parse(c.command)
      }catch(e){
        c.command = [c.command]
      }
    } else {
      c.command = []
    }
    if(c.args) {
      try{
        c.args = JSON.parse(c.args)
      }catch(e){
        c.args = [c.args]
      }
    } else {
      c.args = []
    }
    if(c.securityContext.runAsUser) {
      c.securityContext.runAsUser = parseInt(c.securityContext.runAsUser)
    }
    if(c.securityContext.runAsGroup) {
      c.securityContext.runAsGroup = parseInt(c.securityContext.runAsGroup)
    }
    for(let p of c.ports) {
      if(!p.containerPort) {
        return {err: `应用资源${tpl.kind}/${tpl.metadata.name}容器端口为空`}
      }
      try{
        p.containerPort = parseInt(p.containerPort)
      } catch(e) {
        return {err: `应用资源${tpl.kind}/${tpl.metadata.name}容器端口${p.containerPort}错误`}
      }
    }
    err = transferEnv(c)
    if(err) return {err: `应用资源${tpl.kind}/${tpl.metadata.name}容器${err}`}
    transferResource(c.resources.limits)
    transferResource(c.resources.requests)
    if(c.init){
      initContainers.push(c)
    } else {
      containers.push(c)
    }
    delete c.init
  }
  tpl.spec.template.spec.containers = containers
  tpl.spec.template.spec.initContainers = initContainers
}

function transferResource(res) {
  if(!res) return
  if(res.memory) {
    try{
      res.memory = parseFloat(res.memory) + "M"
    } catch(e){
      
    }
  } else if(res.memory != undefined) {
    delete res.memory
  }
  if(!res.cpu && res.cpu != undefined) {
    delete res.cpu
  }
}

function resolveResource(res) {
  if(res.memory) {
    res.memory = res.memory.substr(0, res.memory.length - 1)
  }
}

function transferProbe(probe) {
  if(!probe.probe) return [null, '']
  console.log(probe)
  let obj = {
    successThreshold: probe.successThreshold, 
    failureThreshold: probe.failureThreshold,
    initialDelaySeconds: probe.initialDelaySeconds, 
    timeoutSeconds: probe.timeoutSeconds, 
    periodSeconds: probe.periodSeconds
  } 
  if(probe.type == 'http' || probe.type == 'https') {
    var port = probe.handle.port
    if(!isNaN(port)) port = parseInt(port)
    if(!port) {
      return [null, "健康检查HTTP端口为空"]
    }
    obj['httpGet'] = {
      path: probe.handle.path,  
      port: port,
      scheme: 'HTTP'
    }
    if(probe.type == 'https') obj.httpGet.scheme = 'HTTPS'
  }
  if(probe.type == 'command') {
    let cmd = []
    if(probe.handle.command) {
      try{
        cmd = JSON.parse(probe.handle.command)
      }catch(e){
        cmd = [probe.handle.command]
      }
    } else {
      return [null, "健康检查命令为空"]
    }
    obj['exec'] = {
      command: cmd
    }
  }
  if(probe.type == 'tcp') {
    var port = probe.handle.port
    if(!isNaN(port)) port = parseInt(port)
    if(!port) {
      return [null, "健康检查TCP端口为空"]
    }
    obj['tcpSocket'] = {
      port: port
    }
  }
  return [obj, '']
}

function transferEnv(c) {
  let envs = []
  for(let e of c.env) {
    if(!e.name) {
      return '环境变量名称为空'
    }
    if(e.type == 'value') {
      envs.push({
        name: e.name,
        value: e.value
      })
    } else if(e.type == 'configMap') {
      envs.push({
        name: e.name,
        valueFrom: {
          configMapKeyRef: {
            name: e.value,
            key: e.key,
          }
        }
      })
    } else if(e.type == 'secret') {
      envs.push({
        name: e.name,
        valueFrom: {
          secretKeyRef: {
            name: e.value,
            key: e.key
          }
        }
      })
    } else if(e.type == 'field') {
      envs.push({
        name: e.name,
        valueFrom: {
          fieldRef: {
            fieldPath: e.value
          }
        }
      })
    } else if(e.type == 'resource') {
      envs.push({
        name: e.name,
        valueFrom: {
          resourceFieldRef: {
            resource: e.value
          }
        }
      })
    }
  }
  c.env = envs
}

function resolveContainerEnv(c) {
  let envs = []
  for(let e of c.env) {
    if(e.value) {
      envs.push({type: 'value', name: e.name, value: e.value})
    } else if(e.valueFrom) {
      if('configMapKeyRef' in e.valueFrom) {
        envs.push({type: 'configMap', name: e.name, value: e.valueFrom.configMapKeyRef.name, key: e.valueFrom.configMapKeyRef.key})
      } else if('secretKeyRef' in e.valueFrom) {
        envs.push({type: 'secret', name: e.name, value: e.valueFrom.secretKeyRef.name, key: e.valueFrom.secretKeyRef.key})
      } else if('fieldRef' in e.valueFrom) {
        envs.push({type: 'field', name: e.name, value: e.valueFrom.configMapKeyRef.fieldPath})
      } else if('resourceFieldRef' in e.valueFrom) {
        envs.push({type: 'resource', name: e.name, value: e.valueFrom.configMapKeyRef.resource})
      }
    }
  }
  c.env = envs
}

function transferPodVolume(tpl) {
  let vols = []
  let volTpls = []
  for(let v of tpl.spec.template.spec.volumes) {
    if(!v.name) return {err: `应用资源${tpl.kind}/${tpl.metadata.name}中存储卷名称为空`}
    if(v.type == 'volumeClaimTemplates' && tpl.kind == 'StatefulSet') {
      if(!v[v.type].requests) return {err: `应用资源${tpl.kind}/${tpl.metadata.name}中volumeClaimTemplate存储请求大小为空`}
      let volTpl = {
        metadata: {
          name: v.name
        },
        spec: {
          accessModes: v[v.type].accessModes,
          storageClassName: v[v.type].storageClassName,
          resources: {
            requests: {
              storage: v[v.type].requests + 'Gi'
            }
          }
        }
      }
      volTpls.push(volTpl)
    } else {
      let vol = {
        name: v.name
      }
      if(v.type == 'configMap' || v.type == 'secret') {
        vol[v.type] = {
          items: v[v.type].items,
        }
        if(v[v.type].defaultMode) {
          vol[v.type].defaultMode = parseInt(v[v.type].defaultMode, 8)
        }
        if(v.type == 'configMap') {
          vol[v.type]['name'] = v[v.type].name
        } else {
          vol[v.type]['secretName'] = v[v.type].secretName
        }
      } else {
        vol[v.type] = v[v.type]
      }
      vols.push(vol)
    }
  }
  if(vols.length > 0) tpl.spec.template.spec.volumes = vols
  if(volTpls.length > 0 && tpl.kind == 'StatefulSet') tpl.spec.volumeClaimTemplates = volTpls 
}

function resolvePodVolume(tpl) {
  let vols = []
  let volTypes = ['persistentVolumeClaim', 'hostPath', 'emptyDir', 'nfs', 'glusterfs']
  if(tpl.kind == 'StatefulSet' && tpl.spec.volumeClaimTemplates) {
    for(let v of tpl.spec.volumeClaimTemplates) {
      let vol = newPodVolume()
      let req = v.spec.resources.requests.storage
      vol.type = 'volumeClaimTemplates'
      vol.name = v.metadata.name
      vol.volumeClaimTemplates = {
        accessModes: v.spec.accessModes,
        storageClassName: v.spec.storageClassName,
        requests: req.substr(0, req.length - 2)
      }
      vols.push(vol)
    }
  }
  for(let v of tpl.spec.template.spec.volumes) {
    let vol = newPodVolume()
    vol.name = v.name
    if(v.configMap) {
      vol.type = 'configMap'
      vol.configMap = v.configMap
    } else if(v.secret) {
      vol.type = 'secret'
      vol.secret = {name: v.secret.secretName}
      if(v.secret.items) vol.secret.items = v.secret.items
      if(v.secret.defaultMode) vol.secret.defaultMode = v.secret.defaultMode
    } else {
      for(let t of volTypes) {
        if(t in v) {
          vol.type = t
          vol[t] = v[t]
        }
      }
    }
    vols.push(vol)
  }
  tpl.spec.template.spec.volumes = vols
}

function transferAffinity(tpl) {
  let podSpec = tpl.spec.template.spec
  if(podSpec.nodeSelector.length > 0) {
    let ns = {}
    for(let s of podSpec.nodeSelector) {
      if(!s.key) return {err: `应用资源${tpl.kind}/${tpl.metadata.name}节点选择标签key为空`}
      ns[s.key] = s.value
    }
    podSpec.nodeSelector = ns
  } else {
    podSpec.nodeSelector = {}
  }
  let affinity = tpl.spec.template.spec.affinity

  var err = transferNodeAffinity(affinity)
  if(err) {
    return {err: `应用资源${tpl.kind}/${tpl.metadata.name}${err}`}
  }
  err = transferPodAffinity(affinity, 'podAffinity')
  if(err) {
    return {err: `应用资源${tpl.kind}/${tpl.metadata.name}${err}`}
  }
  err = transferPodAffinity(affinity, 'podAntiAffinity')
  if(err) {
    return {err: `应用资源${tpl.kind}/${tpl.metadata.name}${err}`}
  }
  if(podSpec.tolerations.length > 0) {
    for(let t of podSpec.tolerations) {
      if(!isNaN(t["tolerationSeconds"])) t["tolerationSeconds"] = parseInt(t["tolerationSeconds"])
    }
  }
}

function transferNodeAffinity(affinity) {
  if(affinity.nodeAffinity.length == 0) {
    affinity.nodeAffinity = {}
    return
  }
  var required = []
  var preferred = []
  for(var na of affinity.nodeAffinity) {
    var labelSelectors = []
    var fieldSelectors = []
    for(var s of na.nodeSelectorTerms) {
      var st = {
        key: s.key || '',
        operator: s.operator
      }
      if(['In', 'NotIn'].indexOf(s.operator) >= 0) {
        if(s.values) {
          var values = s.values.split(',')
          st.values = values
        } else {
          st.values = []
        }
      }
      else if(['Gt', 'Lt'].indexOf(s.operator) >= 0) {
        try{
          var values = parseInt(s.vlaues)
        } catch (err) {
          return "节点亲和性操作为Gt/Lt时，值只能为数字"
        }
        st.values = [values]
      }
      if(s.type === 'label') labelSelectors.push(st)
      else if(s.type === 'field') fieldSelectors.push(st)
    }
    var nodeTerm = {}
    if(labelSelectors.length > 0) nodeTerm.matchExpressions = labelSelectors
    if(fieldSelectors.length > 0) nodeTerm.matchFields = fieldSelectors
    if(na.type === 'required') required.push(nodeTerm)
    else if(na.type === 'preferred') {
      if(!na.weight || na.weight < 1 || na.weight > 100) {
        return "节点亲和性权重值范围为1-100"
      }
      preferred.push({
        weight: na.weight,
        preference: nodeTerm
      })
    }
  }
  var nodeAff = {}
  if(required.length > 0) {
    nodeAff.requiredDuringSchedulingIgnoredDuringExecution = {
      nodeSelectorTerms: required
    }
  }
  if(preferred.length > 0) {
    nodeAff.preferredDuringSchedulingIgnoredDuringExecution = preferred
  }
  if(nodeAff.requiredDuringSchedulingIgnoredDuringExecution || nodeAff.preferredDuringSchedulingIgnoredDuringExecution) {
    affinity.nodeAffinity = nodeAff
  }
}

function resolveNodeAffinityTerms(term) {
  var nt = []
  if(term.matchExpressions){
    for(let s of term.matchExpressions) {
      let l = {type: 'label', key: s.key, operator: s.operator}
      if(['In', 'NotIn'].indexOf(s.operator) >= 0) {
        l.values = s.values ? s.values.join(',') : ''
      }
      else if(['Gt', 'Lt'].indexOf(s.operator) >= 0) {
        l.vlaues = s.values ? s.values[0]: ''
      }
      nt.push(l)
    }
  }
  if(term.matchFields){
    for(let s of term.matchFields) {
      let l = {type: 'field', key: s.key, operator: s.operator}
      if(['In', 'NotIn'].indexOf(s.operator) >= 0) {
        l.values = s.values ? s.values.join(',') : ''
      }
      else if(['Gt', 'Lt'].indexOf(s.operator) >= 0) {
        l.vlaues = s.values ? s.values[0]: ''
      }
      nt.push(l)
    }
  }
  
  return nt
}

function resolveNodeAffinity(affinity) {
  let nf = affinity.nodeAffinity
  if(!nf) affinity.nodeAffinity = []
  let affs = []
  if(nf.requiredDuringSchedulingIgnoredDuringExecution) {
    if(nf.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms) {
      for(let r of nf.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms) {
        affs.push({
          type: 'required',
          nodeSelectorTerms: resolveNodeAffinityTerms(r)
        })
      }
    }
  }
  if(nf.preferredDuringSchedulingIgnoredDuringExecution) {
    for(let p of nf.preferredDuringSchedulingIgnoredDuringExecution) {
      affs.push({
        type: 'preferred',
        weight: p.weight,
        nodeSelectorTerms: resolveNodeAffinityTerms(p.preference)
      })
    }
  }
  affinity.nodeAffinity = affs
}

function transferPodAffinity(affinity, key) {
  var required = []
  var preferred = []
  for(var pa of affinity[key]) {
    if(!pa.podAffinityTerm.topologyKey) {
      return " Pod亲和性中拓扑键不能为空"
    }
    var expressions = []
    var labels = {}
    for(var s of pa.podAffinityTerm.labelSelector) {
      if(s.operator === 'Equal') labels[s.key] = s.values
      else if(['In', 'NotIn'].indexOf(s.operator) >= 0) {
        if(s.values) {
          var values = s.values.split(',')
        } else {
          var values = []
        }
        expressions.push({
          key: s.key,
          operator: s.operator,
          values: values
        })
      } else if (['Exists', 'DoseNotExist'].indexOf(s.operator) >= 0) {
        expressions.push({
          key: s.key,
          operator: s.operator
        })
      }
    }
    var selector = {}
    if(expressions.length > 0) {
      selector.matchExpressions = expressions
    }
    if(JSON.stringify(labels) !== '{}') {
      selector.matchLabels = labels
    }
    var term = {
      namespaces: pa.podAffinityTerm.namespaces,
      topologyKey: pa.podAffinityTerm.topologyKey,
    }
    if(selector.matchExpressions || selector.matchLabels) term.labelSelector = selector
    if(pa.type === 'required') {
      required.push(term)
    } else if(pa.type === 'preferred') {
      if(!pa.weight || pa.weight < 1 || pa.weight > 100) {
        return " Pod亲和性权重值范围为1-100"
      }
      preferred.push({
        weight: parseInt(pa.weight),
        podAffinityTerm: term
      })
    }
  }
  var aff = {}
  if(required.length > 0) {
    aff.requiredDuringSchedulingIgnoredDuringExecution = required
  }
  if(preferred.length > 0) {
    aff.preferredDuringSchedulingIgnoredDuringExecution = preferred
  }
  affinity[key] = aff
}

function resolvePodAffinityTerm(terms) {
  if(!terms) return {}
  let pt = {
    namespaces: terms.namespaces,
    topologyKey: terms.topologyKey,
  }
  let selectors = []
  if(terms.labelSelector && terms.labelSelector.matchLabels) {
    for(let k in terms.labelSelector.matchLabels) {
      selectors.push({
        key: k,
        values: terms.labelSelector.matchLabels[k],
        operator: 'Equal'
      })
    }
  }
  if(terms.labelSelector && terms.labelSelector.matchExpressions) {
    for(let e of terms.labelSelector.matchExpressions) {
      let values = e.values
      if(['In', 'NotIn'].indexOf(e.operator) >= 0) values = e.values.join(',')
      selectors.push({
        key: e.key,
        operator: e.operator,
        values: values,
      })
    }
  }
  pt['labelSelector'] = selectors
  return pt
}

function resolvePodAffinity(affinity, key) {
  var affi = affinity[key]
  if(!affi) affinity[key] = []
  let affs = []
  if(affi.requiredDuringSchedulingIgnoredDuringExecution) {
    for(let r of affi.requiredDuringSchedulingIgnoredDuringExecution) {
      affs.push({
        type: 'required',
        podAffinityTerm: resolvePodAffinityTerm(r)
      })
    }
  }
  if(affi.preferredDuringSchedulingIgnoredDuringExecution) {
    for(let r of affi.preferredDuringSchedulingIgnoredDuringExecution) {
      affs.push({
        type: 'preferred',
        weight: r.weight,
        podAffinityTerm: resolvePodAffinityTerm(r.podAffinityTerm)
      })
    }
  }
  affinity[key] = affs
}

function resolveAffinity(tpl) {
  let podSpec = tpl.spec.template.spec
  // podSpec.affinity = {nodeAffinity: [], podAffinity: [], podAntiAffinity: []}
  if(podSpec.nodeSelector) {
    let ns = []
    for(let k in podSpec.nodeSelector) {
      ns.push([{key: k, values: podSpec.nodeSelector[k]}])
    }
    podSpec.nodeSelector = ns
  } else {
    podSpec.nodeSelector = []
  }
  if(!podSpec.tolerations) podSpec.tolerations = []
  resolveNodeAffinity(podSpec.affinity)
  resolvePodAffinity(podSpec.affinity, 'podAffinity')
  resolvePodAffinity(podSpec.affinity, 'podAntiAffinity')
}

export function newPodVolume() {
  return {
    name: '',
    type: 'persistentVolumeClaim',
    persistentVolumeClaim: {},
    glusterfs: {},
    nfs: {},
    secret: {items: [], name: ''},
    configMap: {items: [], secretName: ''},
    emptyDir: {},
    hostPath: {},
    volumeClaimTemplates: {accessModes: ['ReadWriteOnce']}
  }
}

export function newContainer() {
  return {
    init: false,
    name: '',
    image: '',
    command: '',
    args: '',
    workingDir: '',
    ports: [],
    env: [],
    resources: {limits: {}, requests: {}},
    livenessProbe: {probe: false, type: 'http', handle: {}, successThreshold: 1, failureThreshold: 3,
                    initialDelaySeconds: 0, timeoutSeconds: 1, periodSeconds: 10},
    readinessProbe: {probe: false, type: 'http', handle: {}, successThreshold: 1, failureThreshold: 3,
                    initialDelaySeconds: 0, timeoutSeconds: 1, periodSeconds: 10},
    imagePullPolicy: '',
    volumeMounts: [],
    stdin: false,
    tty: false,
    securityContext: {seLinuxOptions: {}, capabilities: {add: [], drop: []}},
  }
}

function newJobConfig() {
  return {
    backoffLimit: 6,
    concurrencyPolicy: "Allow",
    schedule: "* * * * *"
  }
}

function workloadTemplate() {
  return {
    kind: "Deployment",
    apiVersion: "apps/v1",
    metadata: {
      name: "",
      labels: {},
      namespace: "{{ .Release.Namespace }}"
    },
    spec: {
      job: newJobConfig(),
      replicas: 1,
      strategy: {
        type: 'RollingUpdate',
        maxSurge: "25%",
        maxUnavailable: "25%"
      },
      selector: {
        matchLabels: {}
      },
      template: {
        metadata: {
          labels: {}
        },
        spec: {
          nodeSelector: [],
          tolerations: [],
          affinity: {nodeAffinity: [], podAffinity: [], podAntiAffinity: []},
          securityContext: {sysctls: [], seLinuxOptions: {}},
          hostAliases: [],
          containers: [newContainer()],
          volumes: [],
        }
      }
    }
  }
}

function serviceTemplate() {
  return {
    kind: "Service",
    apiVersion: "v1",
    metadata: {
      name: "",
      labels: {},
      namespace: "{{ .Release.Namespace }}"
    },
    spec: {
      ports: [],
      selector: {},
      type: 'ClusterIP',
    }
  }
}

function transferService(tpl) {
  for(let p of tpl.spec.ports) {
    if(!p.port) {
      return {err: `应用资源${tpl.kind}/${tpl.metadata.name}服务端口为空`}
    }
    try{
      p.port = parseInt(p.port)
    } catch(e) {
      return {err: `应用资源${tpl.kind}/${tpl.metadata.name}服务端口${p.port}错误`}
    }
    if(!p.targetPort) {
      return {err: `应用资源${tpl.kind}/${tpl.metadata.name}容器端口为空`}
    }
    try{
      p.targetPort = parseInt(p.targetPort)
    } catch(e) {
      return {err: `应用资源${tpl.kind}/${tpl.metadata.name}容器端口${p.targetPort}错误`}
    }
    if(tpl.spec.type == 'NodePort' && p.nodePort) {
      try{
        p.nodePort = parseInt(p.nodePort)
      } catch(e) {
        return {err: `应用资源${tpl.kind}/${tpl.metadata.name} nodePort ${p.targetPort}错误`}
      }
    } else if(p.nodePort != undefined) {
      delete p.nodePort
    }
  }
  return {tpl}
}

function configMapTemplate() {
  return {
    kind: "ConfigMap",
    apiVersion: "v1",
    metadata: {
      name: "",
      labels: {},
      namespace: "{{ .Release.Namespace }}"
    },
    data: []
  }
}

function transferConfigMap(tpl) {
  let data = {}
  for(let d of tpl.data) {
    if(!d.key){
      return {err: `应用资源${tpl.kind}/${tpl.metadata.name}配置项key为空`}
    }
    data[d.key] = d.value
  }
  tpl.data = data
  return {tpl}
}

function resolveConfigMap(tpl) {
  let data = []
  for(let k in tpl.data) {
    data.push({key: k, value: tpl.data[k]})
  }
  tpl.data = data
}

function secretTemplate() {
  return {
    kind: "Secret",
    apiVersion: "v1",
    metadata: {
      name: "",
      labels: {},
      namespace: "{{ .Release.Namespace }}"
    },
    data: [],
    tls: {},
    userPass: {},
    imagePass: {},
    type: 'Opaque'
  }
}

function toTransferSecret(tpl) {
  let err = transferSecret(tpl)
  if(err) {
    return {err: `应用资源${tpl.kind}/${tpl.metadata.name}${err}`}
  }
  return {tpl}
}

function toResolveSecret(tpl) {
  resolveSecret(tpl)
}

export function resolveToTemplate(template) {
  if(['Deployment', 'StatefulSet', 'DaemonSet', 'CronJob', 'Job'].indexOf(template.kind) >= 0){
    resolveWorkload(template)
  }
  else if(template.kind == 'ConfigMap') resolveConfigMap(template)
  else if(template.kind == 'Secret') toResolveSecret(template)
  else if(template.kind == 'PersistentVolumeClaim') resolvePvc(template)
}

function resolveWorkload(tpl) {
  if(tpl.kind == 'Job') {
    resolveJob(tpl)
  } else if(tpl.kind == 'CronJob') {
    resolveCronJob(tpl)
  } else {
    tpl.spec.job = {
      backoffLimit: 6,
      concurrencyPolicy: "Allow",
    }
  }
  let strategy = {
    type: 'RollingUpdate',
    maxUnavailable: '25%',
    maxSurge: '25%'
  }
  if(tpl.kind == 'Deployment') {
    if(!tpl.spec.strategy || tpl.spec.strategy.type == 'RollingUpdate' || !tpl.spec.strategy.type) {
      let roll = tpl.spec.strategy ? tpl.spec.strategy.rollingUpdate : null
      strategy = {
        type: 'RollingUpdate',
        maxUnavailable: roll ? roll.maxUnavailable : '25%',
        maxSurge: roll ? roll.maxSurge : '25%'
      }
    } else {
      strategy = {'type': 'Recreate'}
    }
  }
  tpl.spec.strategy = strategy
  if(!tpl.spec.replicas) tpl.spec.replicas = 1
  
  resolveContainers(tpl)
  resolveAffinity(tpl)
  resolvePodVolume(tpl)
  let podSpec = tpl.spec.template.spec
  if(!podSpec.hostAliases) {
    podSpec.hostAliases = []
  }
  if(!podSpec.securityContext) {
    podSpec.securityContext = {sysctls: [], seLinuxOptions: {}}
  }
}

function resolveJob(tpl) {
  tpl.spec.job = newJobConfig()
  if(tpl.spec.backoffLimit) {
    tpl.spec.job.backoffLimit = tpl.spec.backoffLimit
    delete tpl.spec.backoffLimit
  }
  if(tpl.spec.completions) {
    tpl.spec.job.completions = tpl.spec.completions
    delete tpl.spec.completions
  }
  if(tpl.spec.parallelism) {
    tpl.spec.job.parallelism = tpl.spec.parallelism
    delete tpl.spec.parallelism
  }
}

function resolveCronJob(tpl) {
  let jobSpec = tpl.spec.jobTemplate.spec
  jobSpec.job = newJobConfig()
  if(jobSpec.backoffLimit) {
    jobSpec.job.backoffLimit = jobSpec.backoffLimit
    delete jobSpec.backoffLimit
  }
  if(jobSpec.completions) {
    jobSpec.job.completions = jobSpec.completions
    delete jobSpec.completions
  }
  if(jobSpec.parallelism) {
    jobSpec.job.parallelism = jobSpec.parallelism
    delete jobSpec.parallelism
  }
  if(tpl.spec.concurrencyPolicy) {
    jobSpec.job.concurrencyPolicy = tpl.spec.concurrencyPolicy
    delete tpl.spec.concurrencyPolicy
  }
  if(tpl.spec.schedule) {
    jobSpec.job.schedule = tpl.spec.schedule
    delete tpl.spec.schedule
  }
  tpl.spec = jobSpec
}

function resolveContainers(tpl) {
  let podSpec = tpl.spec.template.spec
  if(podSpec.initContainers) {
    for(let c of podSpec.initContainers) {
      c.init = true
      resolveContainer(c)
      podSpec.containers.push(c)
    }
  }
  for(let c of podSpec.containers) {
    resolveContainer(c)
  }
  
}

function resolveContainer(c) {
  c.livenessProbe = resolveProbe(c.livenessProbe)
  c.readinessProbe = resolveProbe(c.readinessProbe)
  if(c.command && c.command.length > 0) c.command = JSON.stringify(c.command)
  else c.command = ''
  if(c.args && c.args.length > 0) c.args = JSON.stringify(c.args)
  else c.args = ''
  resolveContainerEnv(c)
  if(c.resources) {
    resolveResource(c.resources.requests)
    resolveResource(c.resources.limits)
  } else {
    c.resources = {}
  }
}

function resolveProbe(probe) {
  if(!probe) return {probe: false, type: 'http', handle: {}, successThreshold: 1, failureThreshold: 3,
  initialDelaySeconds: 0, timeoutSeconds: 1, periodSeconds: 10}
  probe.probe = true
  if('httpGet' in probe) {
    probe.type = 'http'
    if(probe.httpGet.scheme == 'HTTPS') probe.type = 'https'
    probe.handle = probe.httpGet
    delete probe.httpGet
  } else if('tcpSocket' in probe) {
    probe.type = 'tcp'
    probe.handle = probe.tcpSocket
    delete probe.tcpSocket
  } else if('exec' in probe) {
    probe.type = 'command'
    probe.handle = probe.exec
    delete probe.exec
  }
  return probe
}

function pvcTemplate() {
  return {
    kind: "PersistentVolumeClaim",
    apiVersion: "v1",
    metadata: {
      name: "",
      labels: {},
      namespace: "{{ .Release.Namespace }}"
    },
    spec: {
      accessModes: [],
      storageClassName: '',
      resources: {requests: {}},
    }
  }
}

function transferPvc(tpl) {
  if(tpl.spec.accessModes.length == 0) {
    return {err: `应用资源${tpl.kind}/${tpl.metadata.name} accessModes为空`}
  }
  if(tpl.spec.resources.requests.storage) {
    tpl.spec.resources.requests.storage += 'Gi'
  }
  return {tpl}
}

function resolvePvc(tpl) {
  let req = tpl.spec.resources.requests.storage
  if(req) {
    tpl.spec.resources.requests.storage = req.substr(0, req.length - 2)
  }
  return {tpl}
}