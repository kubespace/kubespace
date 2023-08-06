import { globalSettings } from '@/api/settings/settings'

const getDefaultState = () => {
  return {
    releaseVersion: "",
    globalSettings: {},
  }
}

const state = getDefaultState()

const mutations = {
  SET_GLOBALSETTINGS: (state, globalSettings) => {
    state.globalSettings = globalSettings
  },
  SET_RELEASE_VERSION: (state, releaseVersion) => {
    state.releaseVersion = releaseVersion
  },
}

const actions = {

  // get global settings
  getGlobalSettings({ commit, state }) {
    return new Promise((resolve, reject) => {
      globalSettings().then(response => {
        const { data } = response

        if (!data) {
          return reject('Verification failed, please Login again.')
        }
        
        commit('SET_RELEASE_VERSION', data.release_version)
        commit('SET_GLOBALSETTINGS', data)
        // console.log(data)
        resolve(data)
      }).catch(error => {
        reject(error)
      })
    })
  },
}

export default {
  namespaced: true,
  state,
  mutations,
  actions
}
