const getDefaultState = () => {
  return {
    clusterWatch: null,
  }
}

const state = getDefaultState()

const mutations = {
  UPDATE_CLUSTER_WATCH: (state, obj) => {
    state.clusterWatch = obj
  },
}

const getters = {
  resourceWatch: (state) => (res) => {
    if (state.clusterWatch && state.clusterWatch.obj && state.clusterWatch.obj == res) {
      return state.clusterWatch
    }
  },
  // uidWatch: (state) => (uid) => {
  //   if (state.clusterWatch && state.clusterWatch.resource && state.clusterWatch.resource.metadata.uid == uid) {
  //     return state.clusterWatch
  //   }
  // },
  globalClusterWatch: ( _, getters ) => {
    return getters.resourceWatch("osp:cluster")
  },
  podWatch: ( _, getters ) => {
    return getters.resourceWatch("pods")
  },
  eventWatch: ( _, getters ) => {
    return getters.resourceWatch("event")
  },
  deploymentWatch: ( _, getters ) => {
    return getters.resourceWatch("deployment")
  },
  statefulsetsWatch: ( _, getters ) => {
    return getters.resourceWatch("statefulset")
  },
  daemonsetsWatch: ( _, getters ) => {
    return getters.resourceWatch("daemonset")
  },
  jobsWatch: ( _, getters ) => {
    return getters.resourceWatch("job")
  },
  cronjobsWatch: ( _, getters ) => {
    return getters.resourceWatch("cronjob")
  },
  servicesWatch: ( _, getters ) => {
    return getters.resourceWatch("service")
  },
  ingressesWatch: ( _, getters ) => {
    return getters.resourceWatch("ingress")
  },
  rolebindingsWatch: ( _, getters ) => {
    return getters.resourceWatch("rolebinding")
  },
  rolesWatch: ( _, getters ) => {
    return getters.resourceWatch("role")
  },
  pvcsWatch: ( _, getters ) => {
    return getters.resourceWatch("pvc")
  },
  pvsWatch: ( _, getters ) => {
    return getters.resourceWatch("pv")
  },
  scsWatch: ( _, getters ) => {
    return getters.resourceWatch("sc")
  },
  nodesWatch: ( _, getters ) => {
    return getters.resourceWatch("node")
  }
}

export default {
  namespaced: true,
  state,
  getters,
  mutations,
}
