
import store from './store'
import { getToken } from '@/utils/auth' // get token from cookie

var wsOnOpen = function() {
  // ws.send(JSON.stringify({action: "watchCluster", params: {cluster: "test"}}))
}

var wsOnError = function(e) {
  console.log("ws connect error", e)
}

var wsOnMessage = function(e) {
  let data = JSON.parse(e.data)
  store.commit('ws/UPDATE_CLUSTER_WATCH', data)
}

var wsOnClose = function(e) {
  console.log("ws closed", e)
}

var ws = null

function connect() {
  ws = new WebSocket(`ws://${window.location.host}/osp/api/connect`)
  ws.onopen = wsOnOpen
  ws.onerror = wsOnError
  ws.onmessage = wsOnMessage
  ws.onclose = wsOnClose
}

// function reConnect(num) {
//   if (num <= 0) return
//   setTimeout(() => {
//     connect()
//   }, 3000)
// }

const hasToken = getToken()

if (hasToken) {
  // connect()
}
// connect()

export default ws
