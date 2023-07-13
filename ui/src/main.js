import Vue from 'vue'
import VueSSE from 'vue-sse'

import 'normalize.css/normalize.css' // A modern alternative to CSS resets

import ElementUI from 'element-ui'
import 'element-ui/lib/theme-chalk/index.css'
import App from './App.vue'

import '@/permissions'
import router from './router'
import store from './store'
// import '@/websocket'

import '@/styles/index.scss' // global css
import '@/icons' // icon
import VueClipboard from 'vue-clipboard2';

import { createPerm, updatePerm, deletePerm, hasPermission, podOpPerm } from "@/api/settings/user_role";
import { viewerRole, editorRole, adminRole, hasScopePermission } from "@/api/settings/user_role";
import { dateFormat } from '@/utils/utils'

Vue.prototype.$createPerm = createPerm
Vue.prototype.$updatePerm = updatePerm
Vue.prototype.$deletePerm = deletePerm
Vue.prototype.$hasPermission = hasPermission
Vue.prototype.$podOpPerm = podOpPerm
Vue.prototype.$dateFormat = dateFormat

Vue.prototype.$viewerRole = viewerRole
Vue.prototype.$editorRole = editorRole
Vue.prototype.$adminRole = adminRole
Vue.prototype.$hasScopePermission = hasScopePermission

Vue.prototype.$contentHeight = 140

Vue.use(VueClipboard)

Vue.use(ElementUI)

Vue.use(VueSSE, {
  polyfill: true
})

Vue.config.productionTip = false

new Vue({
  el: "#app",
  router,
  store,
  render: h => h(App),
}).$mount('#app')
