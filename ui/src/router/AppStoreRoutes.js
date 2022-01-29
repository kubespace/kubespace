/* Layout */
import Layout from '@/layout'

const appStoreRoutes = [
  {
    path: 'appstore',
    component: Layout,
    hidden: true,
    children: [
      {
        path: '',
        name: 'appstoreIndex',
        hidden: true,
        component: () => import('@/views/appstore/index'),
        meta: { title: '集群管理', 'group': 'appstore', object: 'cluster', noSidebar: true }
      },
    ]
  }, 
]

export default appStoreRoutes
