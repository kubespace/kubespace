import request from '@/utils/request'

export function listPersistentVolumeClaim(cluster) {
  return request({
    url: `pvc/${cluster}`,
    method: 'get',
  })
}

export function getPersistentVolumeClaim(cluster, namespace, name, output='') {
  return request({
    url: `pvc/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function updatePersistentVolumeClaim(cluster, namespace, name, yaml) {
  return request({
    url: `pvc/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function deletePersistentVolumeClaims(cluster, data) {
  return request({
    url: `pvc/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function buildPvc(pvc) {
  let size = ''
  console.log(pvc)
  if(pvc.spec.resources && pvc.spec.resources.requests && pvc.spec.resources.requests.storage) {
    size = pvc.spec.resources.requests.storage
  }
  var bp = {
    uid: pvc.metadata.uid,
		name:         pvc.metadata.name,
		status:       pvc.status.phase,
		access_modes:  pvc.spec.accessModes,
		create_time:   pvc.metadata.creationTimestamp,
		namespace:    pvc.metadata.namespace,
		storage_class: pvc.spec.storageClassName ? pvc.spec.storageClassName : '',
		capacity:     size,
  }
  return bp
}
