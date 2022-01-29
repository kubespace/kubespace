import request from '@/utils/request'

export function listPersistentVolume(cluster) {
  return request({
    url: `pv/${cluster}`,
    method: 'get',
  })
}

export function getPersistentVolume(cluster, name, output='') {
  return request({
    url: `pv/${cluster}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function updatePersistentVolume(cluster, name, yaml) {
  return request({
    url: `pv/${cluster}/update/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function deletePersistentVolumes(cluster, data) {
  return request({
    url: `pv/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function buildPv(pv) {
  let size = ''
  if (pv.spec.capacity && pv.spec.capacity.storage) {
    size = pv.spec.capacity.storage
  }
  let claimName = ''
  let claimNamespace = ''
  if (pv.spec.claimRef) {
    claimName = pv.spec.claimRef.name
    claimNamespace = pv.spec.claimRef.namespace
  }
  var bp = {
    uid: pv.metadata.uid,
		name:           pv.metadata.name,
		status:         pv.status.phase,
		storage_class:   pv.spec.storageClassName,
		capacity:       size,
		claim:          claimName,
		claim_namespace: claimNamespace,
		access_modes:    pv.spec.accessModes,
		reclaim_policy:  pv.spec.persistentVolumeReclaimPolicy,
		create_time:     pv.metadata.creationTimestamp,
  }
  return bp
}