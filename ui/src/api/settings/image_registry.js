import request from '@/utils/request'

export function listImageRegistry() {
  return request({
    url: '/settings/image_registry',
    method: 'get',
  })
}

export function createImageRegistry(data) {
  return request({
    url: '/settings/image_registry',
    method: 'post',
    data,
  })
}

export function updateImageRegistry(id, data) {
  return request({
    url: `/settings/image_registry/${id}`,
    method: 'put',
    data,
  })
}

export function deleteImageRegistry(id) {
  return request({
    url: `/settings/image_registry/${id}`,
    method: 'delete',
  })
}