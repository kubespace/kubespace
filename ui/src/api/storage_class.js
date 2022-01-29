import request from '@/utils/request'

export function listStorageClass(cluster) {
  return request({
    url: `storageclass/${cluster}`,
    method: 'get',
  })
}

export function getStorageClass(cluster, name, output='') {
  return request({
    url: `storageclass/${cluster}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function updateStorageClass(cluster, name, yaml) {
  return request({
    url: `storageclass/${cluster}/update/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function deleteStorageClasses(cluster, data) {
  return request({
    url: `storageclass/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function buildSc(sc) {
  let bs = {
		uid: sc.metadata.uid,
		name:          sc.metadata.name,
		create_time:    sc.metadata.creationTimestamp,
		provisioner:   sc.provisioner,
		reclaim_policy: sc.reclaimPolicy,
		binding_mode: sc.volumeBindingMode ? sc.volumeBindingMode : '',
  }
  return bs
}