/* Layout */
import Layout from '@/layout'
import { Noop } from '@/layout/components'

const Routes = [
  {
    path: '',
    name: 'pipeline',
    component: () => import('@/views/pipeline/pipeline'),
    meta: { title: '流水线', icon: 'pipeline', 'group': 'pipeline', object: 'pipeline' }
  },
  {
    path: 'create',
    name: 'pipelineCreate',
    hidden: true,
    component: () => import('@/views/pipeline/pipelineEdit'),
    meta: { title: '流水线', icon: 'pipeline', sideName: 'pipeline', 'group': 'pipeline', object: 'cluster'}
  },
  {
    path: 'resource',
    name: 'pipelineResource',
    component: () => import('@/views/pipeline/resource'),
    meta: { title: '资源管理', icon: 'resource', 'group': 'pipeline', object: 'pipeline' }
  },
  {
    path: 'permission',
    name: 'pipelinePermission',
    component: () => import('@/views/pipeline/pipeline'),
    meta: { title: '权限配置', icon: 'permission', 'group': 'pipeline', object: 'pipeline' }
  },
  {
    path: 'pipeline/:pipelineId',
    component: Noop,
    hidden: true,
    children: [
      {
        path: 'builds',
        name: 'pipelineBuilds',
        hidden: true,
        component: () => import('@/views/pipeline/build'),
        meta: { title: '流水线', icon: 'pipeline', sideName: 'pipeline', 'group': 'pipeline', object: 'cluster'}
      },
      {
        path: 'edit',
        name: 'pipelineEdit',
        hidden: true,
        component: () => import('@/views/pipeline/pipelineEdit'),
        meta: { title: '流水线', icon: 'pipeline', sideName: 'pipeline', 'group': 'pipeline', object: 'cluster'}
      },
    ]
  }, 
  {
    path: 'pipeline/:pipelineId/build/:buildId',
    name: 'pipelineBuildDetail',
    hidden: true,
    component: () => import('@/views/pipeline/buildDetail'),
    meta: { title: '流水线', icon: 'pipeline', sideName: 'pipeline', 'group': 'pipeline', object: 'cluster'}
  }, 
]

const pipelineRoutes = [
  {
    path: 'pipespace',
    component: Layout,
    hidden: true,
    children: [
      {
        path: '',
        name: 'pipelineWorkspace',
        hidden: true,
        component: () => import('@/views/pipeline/workspace'),
        meta: { title: '集群管理', icon: 'settings_cluster', 'group': 'pipeline', object: 'cluster', noSidebar: true }
      },
    ]
  }, 
  {
    path: 'pipespace/:workspaceId',
    component: Layout,
    hidden: true,
    children: Routes
  }
]

export default pipelineRoutes
