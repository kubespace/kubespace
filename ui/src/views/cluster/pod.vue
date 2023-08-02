<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="pods"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort = "{prop: 'name'}"
        @selection-change="handleSelectionChange"
        row-key="uid"
        >
        <el-table-column
          type="selection"
          width="45">
        </el-table-column>
        <el-table-column
          prop="name"
          label="名称"
          min-width="170"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.namespace, scope.row.name)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="90"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="containerNum"
          label="容器"
          min-width="65"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.init_containers">
            <el-tooltip :content="`${c.name} (${c.status})`" placement="top" v-for="c in scope.row.init_containers" :key="c.name">
              <svg-icon style="margin-top: 7px;" :class="containerClass(c.status)" icon-class="square" />
            </el-tooltip>
            </template>
            <el-tooltip :content="`${c.name} (${c.status})`" placement="top" v-for="c in scope.row.containers" :key="c.name">
              <svg-icon style="margin-top: 7px;" :class="containerClass(c.status)" icon-class="square" />
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column
          prop="restarts"
          label="重启"
          min-width="45"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="node_name"
          label="节点"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="ip"
          label="IP"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="controlled"
          label="控制器"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="created"
          label="创建时间"
          min-width="135"
          show-overflow-tooltip>
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.created) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          label="状态"
          min-width="60"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span :class="podStatusClass(scope.row.status)" style="font-weight: 450">{{ scope.row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label=""
          show-overflow-tooltip
          width="45">
          <template slot-scope="scope">
            <el-dropdown size="medium" >
              <el-link :underline="false"><svg-icon style="width: 1.3em; height: 1.3em;" icon-class="operate" /></el-link>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item @click.native.prevent="nameClick(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="getPodYaml(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>

                <div @mouseover="logContainerShow = true;" @mouseout="logContainerShow = false;">
                  <el-dropdown-item @click.native.prevent="namespace=scope.row.namespace; selectContainer = scope.row.containers[0].name; selectPodName = scope.row.name; logDialog = true;">
                    <div class="download">
                      <div>
                        <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="log" />
                        <span style="margin-left: 5px;">日志</span>
                      </div>
                      <div class="download-right" v-show="scope.row.containerNum > 1 && logContainerShow">
                        <div class="download-item" v-for="c in scope.row.init_containers" :key="c.name"
                            @click="namespace=scope.row.namespace; selectContainer = c.name; selectPodName = scope.row.name; logDialog = true;">
                            {{ c.name }}
                        </div>
                        <div class="download-item" v-for="c in scope.row.containers" :key="c.name"
                            @click="namespace=scope.row.namespace; selectContainer = c.name; selectPodName = scope.row.name; logDialog = true;">
                            {{ c.name }}
                        </div>
                      </div>
                    </div>
                  </el-dropdown-item>
                </div>
                <div @mouseover="termContainerShow = true;" @mouseout="termContainerShow = false;">
                  <el-dropdown-item :disabled="!$editorRole()" @click.native.prevent="namespace=scope.row.namespace; selectContainer = scope.row.containers[0].name; selectPodName = scope.row.name; terminalDialog = true;">
                    <div class="download">
                      <div>
                        <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="terminal" />
                        <span style="margin-left: 5px;">终端</span>
                      </div>
                      <div class="download-right" v-show="scope.row.containers.length > 1 && termContainerShow">
                        <div class="download-item" v-for="c in scope.row.containers" :key="c.name"
                            @click="namespace=scope.row.namespace; selectContainer = c.name; selectPodName = scope.row.name; terminalDialog = true;">
                            {{ c.name }}
                        </div>
                      </div>
                    </div>
                  </el-dropdown-item>
                </div>

                <el-dropdown-item @click.native.prevent="openEventDialog(scope.row)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="event_dialog" />
                  <span style="margin-left: 5px;">事件</span>
                </el-dropdown-item>

                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="deletePods([{namespace: scope.row.namespace, name: scope.row.name}])">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                  <span style="margin-left: 5px;">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
      <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
        <el-button plain @click="updatePod()" size="small">确 定</el-button>
      </span>
    </el-dialog>

    <el-dialog title="终端" :visible.sync="terminalDialog" :close-on-click-modal="false" width="80%" top="55px">
      <terminal v-if="terminalDialog" :cluster="cluster" :namespace="namespace" :pod="selectPodName" :container="selectContainer"></terminal>
    </el-dialog>

    <el-dialog title="日志" :visible.sync="logDialog" :close-on-click-modal="false" width="80%" top="55px">
      <log v-if="logDialog" :cluster="cluster" :namespace="namespace" :pod="selectPodName" :container="selectContainer"></log>
    </el-dialog>

    <el-dialog :title="'事件:'+eventPodName" :visible.sync="eventDialog" width="60%" top="55px">
      <div class="msgClass">
        <el-table
          v-if="podEvents && podEvents.length > 0"
          :data="podEvents"
          class="table-fix"
          tooltip-effect="dark"
          style="width: 100%"
          v-loading="eventLoading"
          :cell-style="cellStyle"
          :default-sort = "{prop: 'event_time', order: 'descending'}"
          >
          <el-table-column
            prop="type"
            label="类型"
            min-width="30"
            show-overflow-tooltip>
          </el-table-column>
          <el-table-column
            prop="reason"
            label="原因"
            min-width="30"
            show-overflow-tooltip>
            <template slot-scope="scope">
              <span>
                {{ scope.row.reason ? scope.row.reason : "—" }}
              </span>
            </template>
          </el-table-column>
          <el-table-column
            prop="message"
            label="信息"
            min-width="120">
            <template slot-scope="scope">
              <span>
                {{ scope.row.message ? scope.row.message : "—" }}
              </span>
            </template>
          </el-table-column>
          <el-table-column
            prop="event_time"
            label="触发时间"
            min-width="50"
            show-overflow-tooltip>
            <template slot-scope="scope">
              <span>
                {{ $dateFormat(scope.row.event_time) }}
              </span>
            </template>
          </el-table-column>
        </el-table>
        <div v-else style="padding: 25px 15px ; color: #909399; text-align: center">暂无事件发生</div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { ResType, listResource, getResource, delResource, updateResource, watchResource } from '@/api/cluster/resource'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'
import { Terminal } from '@/views/components'
import { Log } from '@/views/components'

export default {
  name: 'Pod',
  components: {
    Clusterbar,
    Yaml,
    Terminal,
    Log,
  },
  data() {
    return {
      yamlDialog: false,
      yamlNamespace: "",
      yamlName: "",
      yamlValue: "",
      yamlLoading: true,
      cellStyle: {border: 0},
      titleName: ["Pods"],
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      originPods: [],
      search_ns: [],
      search_name: '',
      delFunc: undefined,
      delPods: [],
      clusterSSE: undefined,

      logContainerShow: false,
      termContainerShow: false,
      selectContainer: '',
      selectPodName: '',
      namespace: '',
      logDialog: false,
      terminalDialog: false,

      podEvents: [],
      eventPodName: '',
      eventDialog: false,
      eventLoading: false,
    }
  },
  created() {
    this.fetchData()
  },
  beforeDestroy() {
    if(this.clusterSSE) this.clusterSSE.disconnect()
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - this.$contentHeight
        // console.log(heightStyle)
        that.maxHeight = heightStyle
      })()
    }
  },
  watch: {
    cluster: function() {
      this.fetchData()
    }
  },
  computed: {
    pods: function() {
      let plist = []
      for (let p of this.originPods) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        plist.push(p)
      }
      return plist
    },
    cluster() {
      return this.$store.state.cluster
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originPods = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listResource(cluster, "pod").then(response => {
          this.loading = false
          this.originPods = response.data
          this.podsWatch()
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
    podsWatch() {
      this.clusterSSE = watchResource(this.$sse, this.cluster, ResType.Pod, this.podsWatchFunc, {process: true})
    },
    podsWatchFunc: function (newObj) {
      if (newObj) {
        let newUid = newObj.resource.uid
        let newRv = newObj.resource.resource_version
        if (newObj.event === 'add') {
          for(let i in this.originPods) {
            let p = this.originPods[i]
            if (p.uid === newUid) {
              return
            }
          }
          this.originPods.push(newObj.resource)
        } else if (newObj.event === 'update') {
          for (let i in this.originPods) {
            let p = this.originPods[i]
            if (p.uid === newUid && p.resource_version < newRv) {
              let newPod = newObj.resource
              this.$set(this.originPods, i, newPod)
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originPods = this.originPods.filter(( { uid } ) => uid !== newUid)
        }
      }
    },
    nsSearch: function(vals) {
      this.search_ns = []
      for(let ns of vals) {
        this.search_ns.push(ns)
      }
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    nameClick: function(namespace, name) {
      this.$router.push({name: 'podsDetail', params: {namespace: namespace, podName: name}})
    },
    containerClass: function(status) {
      if (status === 'running') return 'running-class'
      if (status === 'terminated') return 'terminate-class'
      if (status === 'waiting') return 'waiting-class'
    },
    getPodYaml: function(namespace, podName) {
      this.yamlNamespace = ""
      this.yamlName = ""
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!namespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (!podName) {
        Message.error("获取Pod名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getResource(cluster, ResType.Pod, namespace, podName, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlName = podName
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deletePods: function(pods) {
      let cs = ''
      for(let c of pods) {
        cs += `${c.namespace}/${c.name}, `
      }
      cs = cs.substr(0, cs.length - 2)
      this.$confirm(`请确认是否删除「${cs}」Pods?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        delResource(this.cluster, ResType.Pod, {resources: pods}).then(() => {
          Message.success("删除成功")
        }).catch((err) => {
          console.log(err)
        });
      }).catch(() => {       
      });
    },
    updatePod: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!this.yamlNamespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (!this.yamlName) {
        Message.error("获取POD参数异常，请刷新重试")
        return
      }
      updateResource(cluster, ResType.Pod, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delPodsFunc: function() {
      if (this.delPods.length > 0){
        let delPods = []
        for (var p of this.delPods) {
          delPods.push({namespace: p.namespace, name: p.name})
        }
        this.deletePods(delPods)
      }
    },
    handleSelectionChange(val) {
      this.delPods = val;
      if (val.length > 0){
        this.delFunc = this._delPodsFunc
      } else {
        this.delFunc = undefined
      }
      // this.multipleSelection = val;
    },
    podStatusClass(status) {
      if(status=='Running') return 'running-class'
      if(status=='Pending') return 'waiting-class'
      if(status=='Succeeded') return 'succeeded-class'
      if(status=='Failed') return 'error-class'
      if(status=='Unknown') return 'terminate-class'
    },
    openEventDialog(pod) {
      this.podEvents = []
      this.eventPodName = pod.name
      this.eventLoading = true
      listResource(this.cluster, ResType.Event, {kind: "Pod", namespace: pod.namespace, name: pod.name}).then(response => {
        this.eventLoading = false
        if (response.data) {
          this.podEvents = response.data.length > 0 ? response.data : []
        }
      }).catch(() => {
        this.eventLoading = false
      })
      this.eventDialog = true
    }
  }
}
</script>

<style lang="scss" scoped>

.name-class {
  cursor: pointer;
}
.name-class:hover {
  color: #409EFF;
}

.scrollbar-wrapper {
  overflow-x: hidden !important;
}
.el-scrollbar__bar.is-vertical {
  right: 0px;
}

.el-scrollbar {
  height: 100%;
}

.running-class {
  color: #67C23A;
}

.terminate-class {
  color: #909399;
}

.waiting-class {
  color: #E6A23C;
}

.error-class {
  color: #F56C6C;
}

.succeeded-class {
  color: #409EFF;
}

.download {
  // width: 70px;
  // height: 40px;
  position: relative;

  .download-right {
    position: absolute;
    right: 70px;
    top: 0px;
    background: #FFF;
    box-shadow: 0 2px 12px 0 rgba(0,0,0,.1);
    border: 1px solid #EBEEF5;
    .download-item {
      display: inline-block;
      margin-right: -8px;
      white-space: nowrap;
      width: auto;
      padding: 0px 12px;
      cursor: pointer;
      color: #606266;
      .item-txt {
        flex: 1;
        display: flex;
        // flex-wrap: nowrap;
        align-items:center;
        font-size: 14px;
      }
    }
    .download-item:hover {
      // background: #1f2326;
      color: #66b1ff;
      // border-radius: 6px;
    }
  }
}
</style>

<style>
.el-dialog__body {
  padding-top: 5px;
  padding-bottom: 5px;
}
</style>
