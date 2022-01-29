import request from '@/utils/request'

export function listEvents(cluster, uid='', kind='', name='', namespace='') {
  var params = {}
  if (uid) params['uid'] = uid
  if (kind) params['kind'] = kind
  if (namespace) params['namespace'] = namespace
  if (name) params['name'] = name
  return request({
    url: `event/${cluster}`,
    method: 'get',
    params
  })
}

export function buildEvent(event) {
  if (!event) return
  let eventTime = event.lastTimestamp
  if (!eventTime) eventTime = event.firstTimestamp
  if (!eventTime) eventTime = event.metadata.creationTimestamp
  return {
    uid: event.metadata.uid,
    namespace: event.metadata.namespace,
    count: event.spec ? event.spec.count : 1,
    reason: event.reason,
    message: event.message,
    type: event.type,
    object: event.involvedObject,
    source: event.source,
    event_time: eventTime,
    resource_version: event.metadata.resourceVersion,
  }
}
