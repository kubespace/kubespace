import Vue from 'vue'
import Router from 'vue-router'

const originalPush = Router.prototype.push
Router.prototype.push = function push(location) {
  return originalPush.call(this, location).catch(err => err)
}

Vue.use(Router)

import ClusterRoutes from './ClusterRoutes'
import SettingsRoutes from './SettingsRoutes'
import PipelineRoutes from './PipelineRoutes'
import WorkspaceRoutes from './WorkspaceRoutes'
import AppstoreRoutes from './AppStoreRoutes'
import { Noop } from '../layout/components'

/**
 * Note: sub-menu only appear when route children.length >= 1
 * Detail see: https://panjiachen.github.io/vue-element-admin-site/guide/essentials/router-and-nav.html
 *
 * hidden: true                   if set true, item will not show in the sidebar(default is false)
 * alwaysShow: true               if set true, will always show the root menu
 *                                if not set alwaysShow, when item has more than one children route,
 *                                it will becomes nested mode, otherwise not show the root menu
 * redirect: noRedirect           if set noRedirect will no redirect in the breadcrumb
 * name:'router-name'             the name is used by <keep-alive> (must set!!!)
 * meta : {
    roles: ['admin','editor']    control the page roles (you can set multiple roles)
    title: 'title'               the name show in sidebar and breadcrumb (recommend set)
    icon: 'svg-name'/'el-icon-x' the icon show in the sidebar
    breadcrumb: false            if set false, the item will hidden in breadcrumb(default is true)
    activeMenu: '/example/list'  if set path, the sidebar will highlight the path you set
  }
 */

/**
 * constantRoutes
 * a base page that does not have permission requirements
 * all roles can be accessed
 */

const oRoutes = [
  {
    path: 'login/admin',
    name: 'login_admin',
    hidden: true,
    component: () => import('@/views/login_admin/index'),
  },

  {
    path: 'login',
    name: 'login',
    hidden: true,
    component: () => import('@/views/login/index'),
  },

  {
    path: '404',
    component: () => import('@/views/404'),
    hidden: true
  },

  {
    path: 'test_yaml',
    component: () => import('@/views/yaml'),
    hidden: true
  },
]

const constantRoutes = [...ClusterRoutes, ...SettingsRoutes, ...PipelineRoutes, ...WorkspaceRoutes, ...AppstoreRoutes, ...oRoutes]

export const routes = [
  {
    path: '/',
    component: Noop,
    hidden: true,
    redirect: '/ui/workspace'
  },
  {
  path: '/ui',
  component: Noop,
  hidden: true,
  redirect: '/ui/workspace',
  children: constantRoutes
  },

  // 404 page must be placed at the end !!!
  { path: '*', redirect: '/ui/404', hidden: true }
]


const createRouter = () => new Router({
  mode: 'history', // require service support
  scrollBehavior: () => ({ y: 0 }),
  routes: routes
})
  
const router = createRouter()
  
// Detail see: https://github.com/vuejs/vue-router/issues/1234#issuecomment-357941465
export function resetRouter() {
  const newRouter = createRouter()
  router.matcher = newRouter.matcher // reset router
}
  
export default router
