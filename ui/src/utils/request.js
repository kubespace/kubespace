import axios from 'axios'
import { Message } from 'element-ui'
import { getToken } from '@/utils/auth'
import store from '@/store'

// create an axios instance
const service = axios.create({
  baseURL: '/api/v1', // url = base url + request url
  // withCredentials: true, // send cookies when cross-domain requests
  timeout: 300000 // request timeout
})

// request interceptor
service.interceptors.request.use(
  config => {
    // do something before request is sent

    const token = getToken()
    if (token) {
      // let each request carry token
      // ['X-Token'] is a custom headers key
      // please modify it according to the actual situation
      config.headers['Authorization'] = 'Berear ' + token
    }
    return config
  },
  error => {
    // do something with request error
    console.log(error) // for debug
    return Promise.reject(error)
  }
)

// response interceptor
service.interceptors.response.use(
  /**
   * If you want to get http information such as headers or status
   * Please return  response => response
  */

  /**
   * Determine the request status by custom code
   * Here is just an example
   * You can also judge the status by HTTP Status Code
   */
  response => {
    if(response.status == 401) {
      let loginUrl = `/ui/login?redirect=${href}`
      store.dispatch('user/resetToken')
      parent.location.href = loginUrl
    }
    const res = response.data
    // if the custom code is not Success, it is judged as an error.
    if (res.code !== "Success") {
      if(response.config.url != "project/apps/status") {
        Message({
          message: res.msg || 'Error',
          type: 'error',
          duration: 5 * 1000
        })
      }
      return Promise.reject(new Error(res.msg || 'Error'))
    } else {
      return res
    }
  },
  error => {
    console.log('err' + error) // for debug
    let href = null
    try {
      href = parent.location.href
    } catch {
      href = document.referrer
    }
    let loginIndex = href && href.indexOf("/ui/login")
    if (loginIndex < 0 && error && error.response && error.response.status === 401) {
      let loginUrl = `/ui/login?redirect=${href}`
      store.dispatch('user/resetToken')
      parent.location.href = loginUrl
    } else {
      let errMsg = error.message
      if (error && error.response && error.response.data && error.response.data.msg) {
        errMsg = error.response.data.msg
      }
      // Message({
      //   message: errMsg,
      //   type: 'error',
      //   duration: 5 * 1000,
      //   offset: 12
      // })
      if(error.response.config.url != 'project/apps/status') Message.error(errMsg)
    }
    return Promise.reject(error)
  }
)

export default service
