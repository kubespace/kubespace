import router from './router'
import store from './store'
import path from 'path'
import { Message } from 'element-ui'
import { getToken } from '@/utils/auth' // get token from cookie

const loginPath = '/ui/login'
const loginAdminPath = '/ui/login/admin'
const whiteList = [loginPath, loginAdminPath, '/ui/test_yaml'] // no redirect whitelist

router.beforeEach(async(to, from, next) => {

  // determine whether the user has logged in
  const hasToken = getToken()
  const toPath = path.resolve(to.path)
  const fromPath = path.resolve(from.path)
  if (toPath.indexOf('/ui/404') > 0) {
    next()
  } else {
    if (hasToken) {
      if (toPath === loginPath) {
        // if is logged in, redirect to the home page
        next({ path: '/' })
      } else {
        const hasGetUserInfo = store.getters.username
        if (hasGetUserInfo) {
          next()
        } else {
          try {
            // get user info
            await store.dispatch('user/getInfo')
            next()
          } catch (error) {
            // remove token and go to login page to re-login
            await store.dispatch('user/resetToken')
            console.log(error)
            if (fromPath !== loginPath) {
              Message.error(error || 'Has Error')
              parent.location.href = `/ui/login?redirect=${toPath}`
            }
          }
        }
      }
    } else {
      /* has no token*/
  
      if (whiteList.indexOf(toPath) !== -1) {
        // in the free login whitelist, go directly
        next()
      } else {
        // other pages that do not have permission to access are redirected to the login page.
        parent.location.href = `/ui/login?redirect=${toPath}`
      }
    }
  }
})
