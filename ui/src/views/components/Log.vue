<template>
  <div class="log-class" :style="{height: logHeight + 'px'}" id="logDiv">
    <p style="white-space: pre-line">{{ logs }}</p>
  </div>
</template>

<script>
import { getToken } from '@/utils/auth'
import { Message } from 'element-ui'

export default {
  name: 'Logs',
  data() {
    return {
      logs: '',
      socket: null,
      scrollToBottom: true,
    }
  },
  props: {
    cluster: {
      type: String,
      required: true,
      default: ''
    },
    namespace: {
      type: String,
      required: true,
      default: ''
    },
    pod: {
      type: String,
      required: true,
      default: ''
    },
    container: {
      type: String,
      required: false,
      default: ''
    }
  },
  computed: {
    logHeight() {
      return window.innerHeight - 200
    }
  },
  mounted() {
    let logDiv = document.getElementById('logDiv')
    let that = this;
    logDiv.addEventListener('scroll', () => {
      that.scrollToBottom = false
      // console.log(logDiv.scrollHeight, logDiv.scrollTop, logDiv.clientHeight)
      if (logDiv.scrollTop + logDiv.clientHeight === logDiv.scrollHeight) {
        that.scrollToBottom = true
      }
    }, true)
    this.initSocket()
  },
  beforeDestroy() {
    if (this.socket) {
      this.socket.close()
    }
  },
  methods: {
    initSocket() {
      let token = getToken()
      console.log(token)
      if (!this.cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!this.namespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (!this.pod) {
        Message.error("获取POD参数异常，请刷新重试")
        return
      }

      var protocal = window.location.protocol == 'http:' ? 'ws':'wss'
      let wsUrl = `${protocal}://${window.location.host}/ws/log/${this.cluster}/${this.namespace}/${this.pod}`
      this.socket = new WebSocket(wsUrl + `?container=${this.container}&token=${token}`);
      this.socketOnClose();
      this.socketOnOpen();
      this.socketOnError();
      this.socketOnMessage();
    },
    socketOnOpen() {
      this.socket.onopen = () => {
      }
    },
    socketOnClose() {
      this.socket.onclose = () => {
        // console.log('close socket')
      }
    },
    socketOnError() {
      this.socket.onerror = () => {
        // console.log('socket 链接失败')
      }
    },
    socketOnMessage() {
      this.socket.onmessage = (e) => {
        // console.log('socket 链接失败')
        // this.logs.push(e.data)
        // let t = document.getElementById('test').innerHTML
        // document.getElementById('test').innerHTML = t + '<br/>' + e.data
        this.logs += e.data
        let that = this
        this.$nextTick(() => {
          if (that.scrollToBottom) {
            let logDiv = document.getElementById('logDiv')
            logDiv.scrollTop = logDiv.scrollHeight // 滚动高度
          }
        })
      }
    }
  }
}
</script>

<style scoped>
.log-class {
  background: #000;
  color: #fff;
  padding: 0px 10px;
  font-size: 14px;
  height: 500px;
  overflow: auto;
}
</style>
