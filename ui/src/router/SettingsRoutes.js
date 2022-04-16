/* Layout */
import Layout from '@/layout'
import store from '@/store'

const Routes = [
  // {
  //   path: 'userinfo',
  //   name: 'userInfo',
  //   component: () => import('@/views/settings/secret'),
  //   meta: { title: '个人中心', icon: 'personal', 'group': 'settings', object: 'cluster' }
  // },
  {
    path: 'secret',
    name: 'settinsSecret',
    component: () => import('@/views/settings/secret'),
    meta: { title: '密钥管理', icon: 'settings_secret', 'group': 'settings', object: 'cluster' }
  },
  {
    path: 'image',
    name: 'settinsImage',
    component: () => import('@/views/settings/image_registry'),
    meta: { title: '镜像仓库', icon: 'docker', 'group': 'settings', object: 'cluster' }
  },
  {
    path: 'member',
    name: 'member',
    component: () => import('@/views/settings/member/index'),
    meta: { title: '用户管理', icon: 'member', 'group': 'settings', object: 'user' }
  },
  {
    path: 'platform_role',
    name: 'platform_role',
    component: () => import('@/views/settings/platform_role'),
    meta: { title: '平台权限', icon: 'platform_perm', 'group': 'settings', object: 'role' }
  },
]

const settingsRoutes = [{
  path: 'settings',
  component: Layout,
  hidden: true,
  children: Routes
}]

export default settingsRoutes
