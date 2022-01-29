import request from '@/utils/request'

export function listNodes(cluster) {
  return request({
    url: `nodes/${cluster}`,
    method: 'get',
  })
}

export function getNode(cluster, name, output='') {
  return request({
    url: `nodes/${cluster}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function updateNode(cluster, name, yaml) {
  return request({
    url: `nodes/${cluster}/update/${name}`,
    method: 'post',
    data: { yaml }
  })
}

export function deleteNodes(cluster, data) {
  return request({
    url: `nodes/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function buildNode(node) {
  let bn = {
		uid:              node.metadata.uid,
		name:             node.metadata.name,
		taints:           len(node.spec.taints),
		version:          node.status.nodeInfo.kubeletVersion,
		os:               node.status.nodeInfo.operatingSystem,
		os_image:          node.status.nodeInfo.osImage,
		kernel_version:    node.status.nodeInfo.kernelVersion,
		container_runtime: node.status.nodeInfo.containerRuntimeVersion,
		labels:           node.metadata.labels,
		allocatable_cpu:   node.status.allocatable.cpu,
		total_cpu:         node.status.capacity.cpu,
		allocatable_mem:   node.status.allocatable.memory,
		total_mem:         node.status.capacity.memory,
    created:          node.metadata.creationTimestamp,
    status: 'Ready',
    internal_ip: ''
  }
  for (c in node.status.conditions) {
		if (c.type === "Ready" && c.status === "True") {
			bn.status = "Ready"
		} else {
			bn.status = "NotReady"
		}
  }
  for (a in node.status.addresses) {
		if (a.type === "InternalIP") {
			bn.internal_ip = i.address
		}
	}
  return bn
}
