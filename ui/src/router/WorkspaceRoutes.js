/* Layout */
import Layout from '@/layout'
import { Noop } from '@/layout/components'

const Routes = [
  {
    path: '',
    name: 'workspaceOverview',
    component: () => import('@/views/workspace/overview'),
    meta: { title: '项目概览', icon: 'overview', 'group': 'workspace', object: 'pipeline' }
  },
  {
    path: 'app',
    name: 'workspaceApp',
    component: () => import('@/views/workspace/overview'),
    meta: { title: '项目应用', icon: 'workspace_app', 'group': 'workspace', object: 'pipeline' }
  },
  {
    path: '',
    component: Noop,
    name: 'workspaceConfiguration',
    meta: { title: '应用配置', icon: 'configuration', group: 'workspace' },
    children: [
      {
        path: 'configmaps',
        name: 'workspaceConfigmaps',
        component: () => import('@/views/workspace/overview'),
        meta: { title: '配置项', group: 'workspace', object: 'configmap' },
      },
      {
        path: 'secrets',
        name: 'workspaceSecrets',
        component: () => import('@/views/workspace/overview'),
        meta: { title: '保密字典', group: 'workspace', object: 'secret' },
      },
    ],
  },
  {
    path: '',
    component: Noop,
    name: 'workspaceNetwork',
    meta: { title: '网络', icon: 'network', group: 'workspace' },
    children: [
      { 
        path: 'services',
        name: 'workspaceServices',
        component: () => import('@/views/workspace/overview'),
        meta: { title: '服务', group: 'workspace', object: 'service' },
      },
      {
        path: 'ingress',
        name: 'workspaceIngress',
        component: () => import('@/views/workspace/overview'),
        meta: { title: '路由', group: 'workspace', object: 'ingress' },
      },
    ],
  },
  {
    path: 'pvc',
    name: 'workspacePvc',
    component: () => import('@/views/workspace/overview'),
    meta: { title: '存储声明', icon: 'storage', group: 'workspace', object: 'pvc' },
  },
  {
    path: 'permission',
    name: 'workspacePermission',
    component: () => import('@/views/workspace/overview'),
    meta: { title: '成员管理', icon: 'permission', 'group': 'workspace', object: 'pipeline' }
  },
]

const workspaceRoutes = [
  {
    path: 'workspace',
    component: Layout,
    hidden: true,
    children: [
      {
        path: '',
        name: 'workspaceIndex',
        hidden: true,
        component: () => import('@/views/workspace/index'),
        meta: { title: '集群管理', icon: 'settings_cluster', 'group': 'workspace', object: 'cluster', noSidebar: true }
      },
    ]
  }, 
  {
    path: 'workspace/:workspaceId',
    component: Layout,
    hidden: true,
    children: Routes
  }
]

export default workspaceRoutes
