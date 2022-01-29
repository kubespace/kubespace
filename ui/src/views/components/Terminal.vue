<template>
    <div id="xterm" class="xterm" />
</template>

<script>
import 'xterm/css/xterm.css'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { AttachAddon } from 'xterm-addon-attach'
import { Message } from 'element-ui'

export default {
  name: 'Xterm',
  data() {
    return {
      socket: null,
      term: null
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
  mounted() {
    this.initTerm()
  },
  beforeDestroy() {
    if (this.socket) {
      this.socket.send("\r\nexit\r")
      this.socket.close()
    }
    if (this.term) {
      this.term.dispose()
    }
  },
  methods: {
    initTerm() {
      let rows = Math.floor((window.innerHeight - 100) / 20)
      console.log(rows)
      const term = new Terminal({
        fontSize: 14,
        cursorBlink: true,
        rows: rows,
      });
      const fitAddon = new FitAddon();
      term.loadAddon(fitAddon);
      term.open(document.getElementById('xterm'));
      fitAddon.fit();
      term.focus();
      this.term = term
      this.initSocket()
    },
    initSocket() {
      let width = this.term.cols
      let height = this.term.rows
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
      let wsUrl = `${protocal}://${window.location.host}/ws/exec/${this.cluster}/${this.namespace}/${this.pod}`
      this.socket = new WebSocket(wsUrl + `?container=${this.container}&cols=${width}&rows=${height}`);
      this.socketOnClose();
      this.socketOnOpen();
      this.socketOnError();
    },
    socketOnOpen() {
      this.socket.onopen = () => {
        // 链接成功后
        const attachAddon = new AttachAddon(this.socket);
        this.term.loadAddon(attachAddon)
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
    }
  }
}
</script>
