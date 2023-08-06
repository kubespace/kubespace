/* Layout */
import Layout from '@/layout'
import { Noop } from '@/layout/components'

const Routes = [
  {
    path: '',
    name: 'workspaceOverview',
    component: () => import('@/views/workspace/overview'),
    meta: { title: '空间概览', icon: 'overview', 'group': 'workspace', object: 'pipeline' }
  },
  {
    path: 'work_apps',
    name: 'workspaceApp',
    component: () => import('@/views/workspace/apps'),
    meta: { title: '应用管理', icon: 'workspace_app', 'group': 'workspace', object: 'pipeline' }
  },
  {
    path: 'create_app',
    name: 'workspaceCreateApp',
    hidden: true,
    component: () => import('@/views/workspace/appCreate'),
    meta: { title: '应用管理', icon: 'workspace_app', 'group': 'workspace', sideName: 'workspaceApp',  object: 'pipeline' }
  },
  {
    path: 'edit_app/:appVersionId',
    name: 'workspaceEditApp',
    hidden: true,
    component: () => import('@/views/workspace/appCreate'),
    meta: { title: '应用管理', icon: 'workspace_app', 'group': 'workspace', sideName: 'workspaceApp',  object: 'pipeline' }
  },
  {
    path: 'edit_import_app/:appVersionId',
    name: 'workspaceEditImportApp',
    hidden: true,
    component: () => import('@/views/workspace/appEditImport'),
    meta: { title: '应用管理', icon: 'workspace_app', 'group': 'workspace', sideName: 'workspaceApp',  object: 'pipeline' }
  },
  {
    path: 'detail_app/:appId',
    name: 'workspaceAppDetail',
    hidden: true,
    component: () => import('@/views/workspace/appDetail'),
    meta: { title: '应用管理', icon: 'workspace_app', 'group': 'workspace', sideName: 'workspaceApp',  object: 'pipeline' }
  },
  {
    path: 'version_app/:appId',
    name: 'workspaceAppVersion',
    hidden: true,
    component: () => import('@/views/workspace/appVersion'),
    meta: { title: '应用管理', icon: 'workspace_app', 'group': 'workspace', sideName: 'workspaceApp',  object: 'pipeline' }
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
        component: () => import('@/views/cluster/configMap'),
        meta: { title: '配置项', group: 'workspace', object: 'configmap' },
      },
      {
        path: 'secrets',
        name: 'workspaceSecrets',
        component: () => import('@/views/cluster/secret'),
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
        component: () => import('@/views/cluster/service'),
        meta: { title: '服务', group: 'workspace', object: 'service' },
      },
      {
        path: 'services/:namespace/:serviceName',
        name: 'workspaceServiceDetail',
        hidden: true,
        component: () => import('@/views/cluster/serviceDetail'),
        meta: { title: '服务', group: 'workspace', sideName: 'workspaceServices', object: 'service' },
      },
      {
        path: 'ingress',
        name: 'workspaceIngress',
        component: () => import('@/views/cluster/ingress'),
        meta: { title: '路由', group: 'workspace', object: 'ingress' },
      },
    ],
  },
  {
    path: 'pvc',
    name: 'workspacePvc',
    component: () => import('@/views/cluster/persistentVolumeClaim'),
    meta: { title: '存储声明', icon: 'storage', group: 'workspace', object: 'pvc' },
  },
  {
    path: 'permission',
    name: 'workspacePermission',
    component: () => import('@/views/workspace/workspaceRole'),
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
