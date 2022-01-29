<template>
  <div>
    <clusterbar :titleName="titleName" :delFunc="deletePods" :editFunc="getPodYaml"/>
    <div class="dashboard-container">
      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" inline class="pod-item" style="margin: 15px 10px 20px 10px;">
          <el-form-item label="状态">
            <span>{{ pod.status }}</span>
          </el-form-item>
          <el-form-item label="创建时间">
            <span>{{ pod.created }}</span>
          </el-form-item>
          <el-form-item label="命名空间">
            <span>{{ pod.namespace }}</span>
          </el-form-item>
          <el-form-item label="节点">
            <span>{{ pod.node_name }}</span>
          </el-form-item>
          <el-form-item label="服务账户">
            <span>{{ pod.service_account }}</span>
          </el-form-item>
          <el-form-item label="Pod IP">
            <span>{{ pod.ip }}</span>
          </el-form-item>
          <el-form-item label="控制器">
            <span>{{ pod.controlled }}/{{ pod.controlled_name}}</span>
          </el-form-item>
          <el-form-item label="QoS Class">
            <span>{{ pod.qos }}</span>
          </el-form-item>
          <el-form-item label="标签">
            <template v-for="(val, key) in pod.labels">
              <span :key="key">{{key}}: {{val}}<br/></span>
            </template>
          </el-form-item>
          <el-form-item label="注解">
            <span v-if="!pod.annonations">—</span>
            
            <template v-else v-for="(val, key) in pod.annonations">
              <span :key="key">{{key}}: {{val}}<br/></span>
            </template>
          </el-form-item>
        </el-form>
      </div>

      <el-tabs value="containers" style="padding: 0px 8px;">
        <el-tab-pane label="容器" name="containers">
          <div class="msgClass">
            <el-table
            ref="table"
            :data="containers"
            class="table-fix"
            tooltip-effect="dark"
            style="width: 100%"
            v-loading="loading"
            :cell-style="cellStyle"
            :default-sort = "{prop: 'name'}"
            >
              <el-table-column type="expand" width="20" style="overflow:hidden">
                <template slot-scope="props">
                  <el-form label-position="left" inline class="table-expand">
                    <el-form-item label="名称">
                      <span>{{ props.row.name }}</span>
                    </el-form-item>
                    <el-form-item label="状态">
                      <span>{{ props.row.status }}</span>
                    </el-form-item>
                    <el-form-item label="镜像">
                      <span>{{ props.row.image }}</span>
                    </el-form-item>
                    <el-form-item label="启动命令" v-if="props.row.command.length">
                      <template v-for="a in props.row.command">
                        <span :key="a">{{a}}<br/></span>
                      </template>
                    </el-form-item>
                    <el-form-item label="启动参数" v-if="props.row.args.length">
                      <template v-for="a in props.row.args">
                        <span :key="a">{{a}}<br/></span>
                      </template>
                    </el-form-item>
                    <el-form-item label="端口" v-if="props.row.ports.length">
                      <template v-for="a in props.row.ports">
                        <span :key="a.name">{{a.name ? `${a.name}:` : ''}} {{a.containerPort}}/{{a.protocol}}<br/></span>
                      </template>
                    </el-form-item>
                    <el-form-item label="环境变量" v-if="props.row.env.length">
                      <!-- <span>{{ props.row.env }}</span> -->
                      <template v-for="(i, a) in props.row.env">
                        <span :key="a">
                          {{ envStr(i) }}<br/>
                        </span>
                      </template>
                    </el-form-item>
                    <el-form-item label="目录挂载" v-if="props.row.volume_mounts.length">
                      <template v-for="a in props.row.volume_mounts">
                        <span :key="a.name">{{a.name}} -> {{a.mountPath}} ({{a.readOnly ? "ro" : "rw"}})<br/></span>
                      </template>
                    </el-form-item>
                    <el-form-item label="资源" v-if="props.row.resources && (props.row.resources.requests || props.row.resources.limits)">
                      <div>
                        <span style="width: 80px; display:inline-block"></span>
                        <span style="width: 80px; display: inline-block;">预留</span>
                        <span style="display: inline-block;">限制</span>
                      </div>
                      <div style="margin-top: -10px;">
                        <span style="width: 80px; display:inline-block">cpu</span>
                        <span style="width: 80px; display: inline-block;">{{ resourceFor(props.row.resources, "requests", "cpu") }}</span>
                        <span style="display: inline-block;">{{ resourceFor(props.row.resources, "limits", "cpu") }}</span>
                      </div>
                      <div style="margin-top: -10px;">
                        <span style="width: 80px; display:inline-block">memory</span>
                        <span style="width: 80px; display: inline-block;">{{ resourceFor(props.row.resources, "requests", "memory") }}</span>
                        <span style="display: inline-block;">{{ resourceFor(props.row.resources, "limits", "memory") }}</span>
                      </div>
                    </el-form-item>
                    <el-form-item label="健康检查" v-if="props.row.readiness_probe || props.row.liveness_probe">
                      <div v-for="p in ['readiness_probe', 'liveness_probe']" :key='p'>
                        <div v-if="props.row[p]">
                          <div>
                            <span style="margin-right: 15px; font-weight: 450;">
                              {{ p == 'readiness_probe' ? 'ReadinessProbe' : 'LivenessProbe' }}
                            </span>
                          </div>
                          <div style="margin-top: -15px">
                            <span v-for="(i, c) in props.row[p]" :key="c">
                              <span class="back-class" v-if="['httpGet', 'exec', 'tcpSocket'].indexOf(c) > -1">
                                {{ c }}: {{ i }}
                              </span>
                            </span>
                          </div>
                          <div style="margin-top: -10px;">
                            <span  v-for="(i, c) in props.row[p]" :key="c">
                              <span class="back-class" v-if="['httpGet', 'exec', 'tcpSocket'].indexOf(c) <= -1">
                                {{ c }}: {{ i }}
                              </span>
                            </span>
                          </div>
                        </div>
                      </div>
                    </el-form-item>
                  </el-form>
                </template>
              </el-table-column>
              <el-table-column
                prop="name"
                label="名称"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span class="name-class" @click="toogleExpand(scope.row)">
                    {{ scope.row.name }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column
                prop="image"
                label="镜像"
                min-width="150"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="restarts"
                label="重启"
                min-width="30"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="start_time"
                label="开始时间"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="status"
                label="状态"
                min-width="50"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                label=""
                show-overflow-tooltip
                min-width="20">
                <template slot-scope="scope">
                  <el-dropdown size="medium" >
                    <el-link :underline="false"><svg-icon style="width: 1.3em; height: 1.3em;" icon-class="operate" /></el-link>
                    <el-dropdown-menu slot="dropdown">
                      <el-dropdown-item @click.native.prevent="selectContainer = scope.row.name; log = true">
                        <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.2em" icon-class="log" /> 
                        <span style="margin-left: 5px;">日志</span>
                      </el-dropdown-item>
                      <el-dropdown-item @click.native.prevent="selectContainer = scope.row.name; terminal = true">
                        <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.2em" icon-class="terminal" /> 
                        <span style="margin-left: 5px;">终端</span>
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </el-dropdown>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
        <el-tab-pane label="存储" name="second">
          <div class="msgClass" style="padding: 10px 0px;">
            <template v-if="pod.volumes && pod.volumes.length > 0">
              <div v-for="v in pod.volumes" :key="v.name" style="margin: 15px 25px; font-size: 14px; color: #606266">
                <div style="margin-bottom: 6px;"><b>{{v.name}}</b></div>
                <template v-for="(val, key) in v">
                    <span v-if="key !== 'name'" :key="key"> 
                      <span class="back-class">{{key}}</span>
                      <span v-for="(ival, ikey) in val" :key="ikey" class="back-class">
                        {{ikey}}: {{ival}}
                      </span>
                    </span>
                  </template>
              </div>
            </template>
            <div v-else style="padding: 25px 15px ;color: #909399; text-align: center">无挂载存储</div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="状态" name="conditions">
          <div class="msgClass">
            <el-table
              :data="pod.conditions"
              class="table-fix"
              tooltip-effect="dark"
              style="width: 100%"
              :cell-style="cellStyle"
              :default-sort = "{prop: 'lastProbeTime'}"
              >
              <el-table-column
                prop="type"
                label="类型"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="status"
                label="状态"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="reason"
                label="原因"
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
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>
                    {{ scope.row.message ? scope.row.message : "—" }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column
                label="触发时间"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>
                    {{ scope.row.lastProbeTime ? scope.row.lastProbeTime : scope.row.lastTransitionTime }}
                  </span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
        <el-tab-pane label="事件" name="events">
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
                prop="object"
                label="对象"
                min-width="70"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>
                    {{ scope.row.object.kind }}/{{ scope.row.object.name }}
                  </span>
                </template>
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
                min-width="120"
                show-overflow-tooltip>
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
              </el-table-column>
            </el-table>
            <div v-else style="padding: 25px 15px ; color: #909399; text-align: center">暂无事件发生</div>
          </div>
        </el-tab-pane>
      </el-tabs>

      <el-dialog title="终端" :visible.sync="terminal" :close-on-click-modal="false" width="80%" top="55px">
        <terminal v-if="terminal" :cluster="cluster" :namespace="namespace" :pod="podName" :container="selectContainer"></terminal>
      </el-dialog>

      <el-dialog title="日志" :visible.sync="log" :close-on-click-modal="false" width="80%" top="55px">
        <log v-if="log" :cluster="cluster" :namespace="namespace" :pod="podName" :container="selectContainer"></log>
      </el-dialog>

      <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
        <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
        <span slot="footer" class="dialog-footer">
          <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
          <el-button plain @click="updatePod()" size="small">确 定</el-button>
        </span>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { getPod, deletePods, updatePod, buildPods, resourceFor } from '@/api/pods'
import { listEvents, buildEvent } from '@/api/event'
import { Message } from 'element-ui'
import { Terminal } from '@/views/components'
import { Log } from '@/views/components'

export default {
  name: 'PodDetail',
  components: {
    Clusterbar,
    Terminal,
    Log,
    Yaml
  },
  data() {
    return {
      yamlDialog: false,
      yamlValue: "",
      yamlLoading: true,
      log: false,
      terminal: false,
      cellStyle: {border: 0},
      loading: true,
      originPod: undefined,
      selectContainer: '',
      podEvents: [],
      eventLoading: true,
      activeName: "first",
      resourceFor: resourceFor,
    }
  },
  created() {
    this.fetchData()
  },
  watch: {
    podWatch: function (newObj) {
      if (newObj && this.pod) {
        let newUid = newObj.resource.metadata.uid
        if (newUid !== this.pod.uid) {
          return
        }
        console.log("watch pod obj", newObj)
        let newRv = newObj.resource.metadata.resourceVersion
        if (this.pod.resource_version < newRv) {
          // this.$set(this.originPod, newPod)
          this.originPod = newObj.resource
        }
      }
    },
    eventWatch: function (newObj) {
      if (newObj && this.originPod) {
        let event = newObj.resource
        if (event.involvedObject.namespace !== this.pod.namespace) return
        if (event.involvedObject.uid !== this.pod.uid) return
        let newUid = newObj.resource.metadata.uid
        if (newObj.event === 'add') {
          this.podEvents.push(buildEvent(event))
        } else if (newObj.event == 'update') {
          let newRv = newObj.resource.metadata.resourceVersion
          for (let i in this.podEvents) {
            let d = this.podEvents[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newEvent = buildEvent(newObj.resource)
                this.$set(this.podEvents, i, newEvent)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.podEvents = this.podEvents.filter(( { uid } ) => uid !== newUid)
        }
      }
    }
  },
  computed: {
    titleName: function() {
      return ['Pods', this.podName]
    },
    podName: function() {
      return this.$route.params ? this.$route.params.podName : ''
    },
    namespace: function() {
      return this.$route.params ? this.$route.params.namespace : ''
    },
    pod: function() {
      let p = buildPods(this.originPod)
      return p
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    containers: function() {
      let c = []
      if (this.pod && this.pod.containers) c = this.pod.containers
      if (this.pod && this.pod.init_containers) c = [...this.pod.init_containers, ...c]
      return c
    },
    podWatch: function() {
      return this.$store.getters["ws/podWatch"]
    },
    eventWatch: function() {
      return this.$store.getters["ws/eventWatch"]
    }
  },
  methods: {
    fetchData: function() {
      this.originPods = []
      this.podEvents = []
      this.loading = true
      this.eventLoading = true
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.namespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.podName) {
        Message.error("获取Pod名称参数异常，请刷新重试")
        this.loading = false
        this.eventLoading = false
        return
      }
      getPod(cluster, this.namespace, this.podName).then(response => {
        this.loading = false
        this.originPod = response.data

        listEvents(cluster, this.originPod.metadata.uid).then(response => {
          this.eventLoading = false
          if (response.data) {
            this.podEvents = response.data.length > 0 ? response.data : []
          }
        }).catch(() => {
          this.eventLoading = false
        })
      }).catch(() => {
        this.loading = false
        this.eventLoading = false
      })
    },
    toogleExpand: function(row) {
      let $table = this.$refs.table;
      $table.toggleRowExpansion(row)
    },
    envStr: function(env) {
      let s = env.name + ': '
      if (env.value) {
        s += env.value
      } else if (env.valueFrom) {
        if (env.valueFrom.configMapKeyRef) {
          s += `configmap(${env.valueFrom.configMapKeyRef.key}:${env.valueFrom.configMapKeyRef.name})`
        } else if (env.valueFrom.fieldRef) {
          s += `fieldRef(${env.valueFrom.fieldRef.apiVersion}:${env.valueFrom.fieldRef.fieldPath})`
        } else if (env.valueFrom.secretKeyRef) {
          s += `secret(${env.valueFrom.secretKeyRef.key}:${env.valueFrom.secretKeyRef.name})`
        } else {
          s += String(env.valueFrom)
        }
      }
      return s
    },
    deletePods: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( !this.pod ) {
        Message.error("获取POD参数异常，请刷新重试")
      }
      let pods = [{
        namespace: this.pod.namespace,
        name: this.pod.name,
      }]
      let params = {
        resources: pods
      }
      deletePods(cluster, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    getPodYaml: function() {
      if (!this.pod) {
        Message.error("获取Pod参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getPod(cluster, this.pod.namespace, this.pod.name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    updatePod: function() {
      if (!this.pod) {
        Message.error("获取Pod参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      updatePod(cluster, this.pod.namespace, this.pod.name, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
  }
}
</script>

<style lang="scss" scoped>
.dashboard {
  &-container {
    margin: 10px 30px;
  }
  &-text {
    font-size: 30px;
    line-height: 46px;
  }

  .table-fix {
    height: calc(100% - 100px);
  }
}
.name-class {
  cursor: pointer;
}
.name-class:hover {
  color: #409EFF;
}
.msgClass {
  margin: 8px 10px 15px 10px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}
</style>

<style>
/* .el-table__expand-icon {
  display: none;
} */
.el-table__expanded-cell[class*=cell] {
  padding-top: 5px;
}
.table-expand {
  font-size: 0;
}
.table-expand label {
  width: 90px;
  color: #99a9bf;
  font-weight: 400;
}
.table-expand .el-form-item {
  margin-right: 0;
  margin-bottom: 0;
  width: 100%;
}

.pod-item {
  margin: 20px 5px 20px 5px;
  padding: 10px 20px;
  font-size: 0;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}
.pod-item label {
  width: 90px;
  color: #99a9bf;
  font-weight: 400;
}
.pod-item .el-form-item {
  margin-right: 0;
  margin-bottom: 0;
  width: 50%;
}
.pod-item span {
  color: #606266;
}
/* .el-collapse {
  border-top: 0px;
} */
.title-class {
  margin-left: 5px;
  color: #606266;
  font-size: 13px;
}
.podCollapse .el-collapse-item__content {
  padding: 0px 10px 15px;
  /* font-size: 13px; */
}
.el-dialog__body {
  padding-top: 5px;
}
.msgClass .el-table::before {
  height: 0px;
}
</style>
