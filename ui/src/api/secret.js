import request from '@/utils/request'

export function listSecrets(cluster, params) {
  return request({
    url: `secret/${cluster}`,
    method: 'get',
    params
  })
}

export function getSecret(cluster, namespace, name, output='') {
  return request({
    url: `secret/${cluster}/${namespace}/${name}`,
    method: 'get',
    params: { output }
  })
}

export function updateSecret(cluster, namespace, name, yaml) {
  return request({
    url: `secret/${cluster}/update/${namespace}/${name}`,
    method: 'post',
    data: { yaml, name, namespace }
  })
}

export function deleteSecrets(cluster, data) {
  return request({
    url: `secret/${cluster}/delete`,
    method: 'post',
    data: data
  })
}

export function transferSecret(secret) {
  let data = {}
  if(["kubernetes.io/tls", "kubernetes.io/basic-auth", "kubernetes.io/dockerconfigjson"].indexOf(secret.type) == -1) {
    for(let d of secret.data) {
      if(!d.key){
        return `配置项key不能为空`
      }
      data[d.key] = btoa(encodeURIComponent(d.value))
    }
    secret.data = data
  } else if(secret.type == 'kubernetes.io/tls') {
    secret.data = {
      'tls.crt': btoa(encodeURIComponent(secret.tls['crt'])),
      'tls.key': btoa(encodeURIComponent(secret.tls['key']))
    }
  } else if(secret.type == 'kubernetes.io/basic-auth') {
    secret.data = {
      'username': btoa(encodeURIComponent(secret.userPass['username'])),
      'password': btoa(encodeURIComponent(secret.userPass['password']))
    }
  } else if(secret.type == 'kubernetes.io/dockerconfigjson') {
    if(!secret.imagePass.url) {
      return `镜像仓库地址不能为空`
    }
    let auth = {auths: {}}
    auth.auths[secret.imagePass.url] = {
      'username': secret.imagePass.username,
      'password': secret.imagePass.password,
      'email': secret.imagePass.email,
      'auth': btoa(encodeURIComponent(`${secret.imagePass.username}:${secret.imagePass.password}`))
    }
    secret.data = {
      '.dockerconfigjson': btoa(JSON.stringify(auth))
    }
    console.log(secret.data)
  }
  delete secret.tls
  delete secret.userPass
  delete secret.imagePass
  return
}

export function resolveSecret(secret) {
  secret.tls = {}
  secret.userPass = {}
  secret.imagePass = {}
  let data = []
  if(["kubernetes.io/tls", "kubernetes.io/basic-auth", "kubernetes.io/dockerconfigjson"].indexOf(secret.type) == -1) {
    for(let k in secret.data) {
      data.push({key: k, value: decodeURIComponent(atob(secret.data[k]))})
    }
  } else if(secret.type == 'kubernetes.io/tls') {
    secret.tls['crt'] = decodeURIComponent(atob(secret.data['tls.crt']))
    secret.tls['key'] = decodeURIComponent(atob(secret.data['tls.key']))
  } else if(secret.type == 'kubernetes.io/basic-auth') {
    secret.userPass['username'] = decodeURIComponent(atob(secret.data['username']))
    secret.userPass['password'] = decodeURIComponent(atob(secret.data['password']))
  } else if(secret.type == 'kubernetes.io/dockerconfigjson') {
    let auths = JSON.parse(decodeURIComponent(atob(secret.data['.dockerconfigjson'])))
    for(let k in auths.auths) {
      secret.imagePass = {
        url: k,
        username: auths.auths[k].username,
        password: auths.auths[k].password,
        email: auths.auths[k].email
      }
    }
  }
  secret.data = data
}