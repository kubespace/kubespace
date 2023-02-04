import Vue from 'vue'
import Vuex from 'vuex'
import user from './modules/user'

Vue.use(Vuex)

const getters = {
  token: state => state.user.token,
  username: state => state.user.name,
  permissions: state => state.user.permissions,
  userInfo: state => state.user.userInfo,
}

const state = {
  cluster: "",
  namespace: "",
}

const mutations = {
  SET_CLUSTER: (state, cluster) => {
    if (state.cluster != cluster) {
      // console.log("set cluster ", state.cluster, cluster)
      state.cluster = cluster
    }
  },
  SET_NAMESPACE: (state, namespace) => {
    state.namespace = namespace
  },
}

const actions = {
  watchCluster({ commit }, cluster) {
    commit('SET_CLUSTER', cluster)
  },
  watchNamespace({ commit }, namespace) {
    commit('SET_NAMESPACE', namespace)
  },
}

const store = new Vuex.Store({
  modules: {
    user,
  },
  getters,
  state,
  mutations,
  actions
})

export default store
