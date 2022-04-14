/* Layout */
import Layout from '@/layout'
import { Noop } from '@/layout/components'

const Routes = [
  {
    path: '',
    name: 'cluster',
    component: () => import('@/views/cluster/cluster'),
    meta: { title: '集群', icon: 'cluster', group: 'cluster', "perm": true },
  },

  {
    path: 'node',
    name: 'node',
    component: () => import('@/views/cluster/node'),
    meta: { title: '节点', icon: 'node', group: 'cluster', object: "node" },
  },

  {
    path: 'node/:nodeName',
    name: 'nodeDetail',
    hidden: true,
    component: () => import('@/views/cluster/nodeDetail'),
    meta: { title: '节点', icon: 'node', group: 'cluster', sideName: 'node', object: 'node' },
  },

  {
    path: 'component',
    name: 'app',
    component: () => import('@/views/cluster/app'),
    meta: { title: '组件', icon: 'app', group: 'cluster', object: "app" },
  },

  {
    path: 'appCreate',
    name: 'appCreate',
    hidden: true,
    component: () => import('@/views/cluster/appCreate'),
    meta: { title: '应用', icon: 'app', group: 'cluster', sideName: 'app', object: "app" },
  },
  {
    path: 'app/:namespace/:appName',
    name: 'appDetail',
    hidden: true,
    component: () => import('@/views/cluster/appDetail'),
    meta: { title: '容器组', icon: 'app', group: 'cluster', sideName: 'app', object: 'app' },
  },

  {
    path: 'workloads',
    component: Noop,
    name: 'workloads',
    meta: { title: '工作负载', icon: 'workloads', group: 'cluster' },
    children: [
      // {
      //   path: 'detail',
      //   name: 'detail',
      //   component: () => import('@/views/dashboard/index'),
      //   meta: { title: '概览', group: 'cluster' },
      // },
      {
        path: 'pods',
        name: 'pods',
        component: () => import('@/views/cluster/pod'),
        meta: { title: '容器组', group: 'cluster', object: 'pod' },
      },
      {
        path: 'pods/:namespace/:podName',
        name: 'podsDetail',
        hidden: true,
        component: () => import('@/views/cluster/podDetail'),
        meta: { title: '容器组', group: 'cluster', sideName: 'pods', object: 'pod' },
      },
      {
        path: 'deployments',
        name: 'deployments',
        component: () => import('@/views/cluster/deployment'),
        meta: { title: '无状态', group: 'cluster', object: 'deployment' },
      },
      {
        path: 'deployments/:namespace/:deploymentName',
        name: 'deploymentDetail',
        hidden: true,
        component: () => import('@/views/cluster/deploymentDetail'),
        meta: { title: '无状态', group: 'cluster', sideName: 'deployments', object: 'deployment' },
      },
      {
        path: 'deployments/create',
        name: 'deploymentCreate',
        hidden: true,
        component: () => import('@/views/cluster/deploymentCreate1'),
        meta: { title: '无状态', group: 'cluster', sideName: 'deployments', object: 'deployment' },
      },
      {
        path: 'statefulsets',
        name: 'statefulsets',
        component: () => import('@/views/cluster/statefulset'),
        meta: { title: '有状态', group: 'cluster', object: 'statefulset' }
      },
      {
        path: 'statefulsets/:namespace/:statefulsetName',
        name: 'statefulsetDetail',
        hidden: true,
        component: () => import('@/views/cluster/statefulsetDetail'),
        meta: { title: '有状态', group: 'cluster', sideName: "statefulsets", object: 'statefulset' }
      },
      {
        path: 'statefulsets/create',
        name: 'statefulsetCreate',
        hidden: true,
        component: () => import('@/views/cluster/statefulsetCreate'),
        meta: { title: '有状态', group: 'cluster', sideName: 'statefulsets', object: 'statefulset' },
      },
      {
        path: 'daemonsets',
        name: 'daemonsets',
        component: () => import('@/views/cluster/daemonset'),
        meta: { title: '守护进程集', group: 'cluster', object: 'daemonset' }
      },
      {
        path: 'daemonsets/:namespace/:daemonsetName',
        name: 'daemonsetDetail',
        hidden: true,
        component: () => import('@/views/cluster/daemonsetDetail'),
        meta: { title: '守护进程集', group: 'cluster', sideName: "daemonsets", object: 'daemonset' }
      },
      {
        path: 'daemonsets/create',
        name: 'daemonsetCreate',
        hidden: true,
        component: () => import('@/views/cluster/daemonsetCreate'),
        meta: { title: '守护进程集', group: 'cluster', sideName: 'daemonsets', object: 'daemonset' },
      },
      {
        path: 'job',
        name: 'job',
        component: () => import('@/views/cluster/job'),
        meta: { title: '任务', group: 'cluster', object: 'job' },
      },
      {
        path: 'job/:namespace/:jobName',
        name: 'jobDetail',
        hidden: true,
        component: () => import('@/views/cluster/jobDetail'),
        meta: { title: '任务', group: 'cluster', sideName: "job", object: 'job' }
      },
      {
        path: 'cronjob',
        name: 'cronjob',
        component: () => import('@/views/cluster/cronjob'),
        meta: { title: '定时任务', group: 'cluster', object: 'cronjob' }
      },
      {
        path: 'cronjob/:namespace/:cronjobName',
        name: 'cronjobDetail',
        hidden: true,
        component: () => import('@/views/cluster/cronjobDetail'),
        meta: { title: '定时任务', group: 'cluster', sideName: "cronjob", object: 'cronjob'}
      },
    ]
  },

  {
    path: 'configuration',
    component: Noop,
    name: 'configuration',
    meta: { title: '应用配置', icon: 'configuration', group: 'cluster' },
    children: [
      {
        path: 'configmaps',
        name: 'configmaps',
        component: () => import('@/views/cluster/configMap'),
        meta: { title: '配置项', group: 'cluster', object: 'configmap' },
      },
      {
        path: 'configmaps/:namespace/:configMapName',
        name: 'configMapDetail',
        hidden: true,
        component: () => import('@/views/cluster/configMapDetail'),
        meta: { title: '配置项', group: 'cluster', sideName: 'configmaps', object: 'configmap' },
      },
      {
        path: 'secrets',
        name: 'secrets',
        component: () => import('@/views/cluster/secret'),
        meta: { title: '保密字典', group: 'cluster', object: 'secret' },
      },
      {
        path: 'secrets/:namespace/:secretName',
        name: 'secretDetail',
        hidden: true,
        component: () => import('@/views/cluster/secretDetail'),
        meta: { title: '配置项', group: 'cluster', sideName: 'secrets', object: 'secret' },
      },
      {
        path: 'hpa',
        name: 'hpa',
        component: () => import('@/views/cluster/hpa'),
        meta: { title: '水平扩缩容', group: 'cluster', object: 'hpa' },
      },
      {
        path: 'hpa/:namespace/:hpaName',
        name: 'hpaDetail',
        hidden: true,
        component: () => import('@/views/cluster/hpaDetail'),
        meta: { title: '配置项', group: 'cluster', sideName: 'hpa', object: 'hpa' },
      },
    ],
  },

  {
    path: 'network',
    component: Noop,
    name: 'network',
    meta: { title: '网络', icon: 'network', group: 'cluster' },
    children: [
      {
        path: 'services',
        name: 'services',
        component: () => import('@/views/cluster/service'),
        meta: { title: '服务', group: 'cluster', object: 'service' },
      },
      {
        path: 'services/:namespace/:serviceName',
        name: 'serviceDetail',
        hidden: true,
        component: () => import('@/views/cluster/serviceDetail'),
        meta: { title: '服务', group: 'cluster', sideName: 'services', object: 'service' },
      },
      {
        path: 'ingress',
        name: 'ingress',
        component: () => import('@/views/cluster/ingress'),
        meta: { title: '路由', group: 'cluster', object: 'ingress' },
      },
      {
        path: 'ingress/:namespace/:ingressName',
        name: 'ingressDetail',
        hidden: true,
        component: () => import('@/views/cluster/ingressDetail'),
        meta: { title: '路由', group: 'cluster', sideName: 'ingress', object: 'ingress' },
      },
      {
        path: 'networkpolicies',
        name: 'networkpolicies',
        component: () => import('@/views/cluster/networkpolicy'),
        meta: { title: '网络策略', group: 'cluster', object: 'networkPolicy' },
      },
    ],
  },

  {
    path: 'storage',
    component: Noop,
    name: 'storage',
    meta: { title: '存储', icon: 'storage', group: 'cluster' },
    children: [
      {
        path: 'pvc',
        name: 'pvc',
        component: () => import('@/views/cluster/persistentVolumeClaim'),
        meta: { title: '存储声明', group: 'cluster', object: 'pvc' },
      },
      {
        path: 'pvc/:namespace/:persistentVolumeClaimName',
        name: 'pvcDetail',
        hidden: true,
        component: () => import('@/views/cluster/persistentVolumeClaimDetail'),
        meta: { title: '配置项', group: 'cluster', sideName: 'pvc', object: 'pvc' },
      },
      {
        path: 'pv',
        name: 'pv',
        component: () => import('@/views/cluster/persistentVolume'),
        meta: { title: '存储卷', group: 'cluster', object: 'pv' }
      },
      {
        path: 'pv/:persistentVolumeName',
        name: 'pvDetail',
        hidden: true,
        component: () => import('@/views/cluster/persistentVolumeDetail'),
        meta: { title: '配置项', group: 'cluster', sideName: 'pv', object: 'pv' },
      },
      {
        path: 'storageclass',
        name: 'storageclass',
        component: () => import('@/views/cluster/storageClass'),
        meta: { title: '存储类', group: 'cluster', object: 'sc' },
      },
    ],
  },

  {
    path: 'namespace',
    name: 'namespace',
    component: () => import('@/views/cluster/namespace'),
    meta: { title: '命名空间', icon: 'namespace', group: 'cluster', object: 'namespace' },
  },
  {
    path: 'event',
    name: 'event',
    component: () => import('@/views/cluster/event'),
    meta: { title: '事件', icon: 'event', group: 'cluster', object: 'event' },
  },
  {
    path: 'rbac',
    component: Noop,
    name: 'rbac',
    meta: { title: '访问控制', icon: 'security', group: 'cluster' },
    children: [
      {
        path: 'serviceaccount',
        name: 'serviceaccount',
        component: () => import('@/views/cluster/serviceaccount'),
        meta: { title: '服务账户', group: 'cluster', object: 'serviceaccount' },
      },
      {
        path: 'serviceaccount/:namespace/:serviceaccountName',
        name: 'serviceaccountDetail',
        hidden: true,
        component: () => import('@/views/cluster/serviceaccountDetail'),
        meta: { title: '服务账户', group: 'cluster', sideName: 'serviceaccount', object: 'serviceaccount' },
      },
      {
        path: 'rolebinding',
        name: 'rolebinding',
        component: () => import('@/views/cluster/rolebinding'),
        meta: { title: '角色绑定', group: 'cluster', object: 'rolebinding' },
      },
      {
        path: 'rolebinding/:rolebindingName',
        name: 'clusterrolebindingDetail',
        hidden: true,
        component: () => import('@/views/cluster/rolebindingDetail'),
        meta: { title: '角色绑定', group: 'cluster', sideName: 'rolebinding', object: 'rolebinding' },
      },
      {
        path: 'rolebinding/:namespace/:rolebindingName',
        name: 'rolebindingDetail',
        hidden: true,
        component: () => import('@/views/cluster/rolebindingDetail'),
        meta: { title: '角色绑定', group: 'cluster', sideName: 'rolebinding', object: 'rolebinding' },
      },
      {
        path: 'role',
        name: 'role',
        component: () => import('@/views/cluster/role'),
        meta: { title: '角色', group: 'cluster', object: 'role' },
      },
      {
        path: 'role/:roleName',
        name: 'clusterroleDetail',
        hidden: true,
        component: () => import('@/views/cluster/roleDetail'),
        meta: { title: '角色绑定', group: 'cluster', sideName: 'role', object: 'role' },
      },
      {
        path: 'role/:namespace/:roleName',
        name: 'roleDetail',
        hidden: true,
        component: () => import('@/views/cluster/roleDetail'),
        meta: { title: '角色绑定', group: 'cluster', sideName: 'role', object: 'role' },
      },
    ]
  },
  {
    path: 'crd',
    name: 'crd',
    component: () => import('@/views/cluster/crd'),
    meta: { title: 'CRD', icon: 'crd', group: 'cluster', object: 'crd' },
  },
  {
    path: 'permission',
    name: 'clusterPermission',
    component: () => import('@/views/cluster/clusterRole'),
    meta: { title: '成员管理', icon: 'cluster_perm', 'group': 'cluster', object: 'permission' }
  },
]

const clusterRoutes = [
  {
    path: 'cluster',
    component: Layout,
    hidden: true,
    children: [{
      path: '',
      name: 'clusterIndex',
      hidden: true,
      component: () => import('@/views/cluster/index'),
      meta: { title: 'cluster', icon: '', group: 'cluster', object: '', noSidebar: true },
    }],
    meta: { group: 'cluster' },
  },
  {
    path: 'cluster/:clusterId',
    component: Layout,
    hidden: true,
    children: Routes,
    meta: { group: 'cluster' },
  },
]

export default clusterRoutes
