/* Layout */
import Layout from '@/layout'

const appStoreRoutes = [
  {
    path: 'storeapp',
    component: Layout,
    hidden: true,
    children: [
      {
        path: '',
        name: 'appstoreIndex',
        hidden: true,
        component: () => import('@/views/appstore/store'),
        meta: { title: '集群管理', 'group': 'appstore', object: 'cluster', noSidebar: true }
      },
      {
        path: 'version/:appId',
        name: 'appstoreVersions',
        hidden: true,
        component: () => import('@/views/appstore/appVersion'),
        meta: { title: '集群管理', 'group': 'appstore', object: 'cluster', noSidebar: true }
      },
    ]
  }, 
]

export default appStoreRoutes
