<template>
  <div>
    <clusterbar :titleName="titleName" :delFunc="deleteJobs" :editFunc="getJobYaml"/>
    <div class="dashboard-container">

      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" inline class="pod-item" style="margin: 15px 10px 30px 10px;">
          <el-form-item label="名称">
            <span>{{ job.name }}</span>
          </el-form-item>
          <el-form-item label="创建时间">
            <span>{{ job.created }}</span>
          </el-form-item>
          <el-form-item label="命名空间">
            <span>{{ job.namespace }}</span>
          </el-form-item>
          <el-form-item label="Completions">
            <span>{{ job.completions }}</span>
          </el-form-item>
          <el-form-item label="Pods">
            <!-- <span>{{ job.number_ready + "/" + job.desired_number_scheduled }}</span> -->
            <span v-if="job.active > 0" class="back-class">
              {{ job.active }} Running
            </span>
            <span v-if="job.succeeded > 0" class="back-class">
              {{ job.succeeded }} Succeeded
            </span>
            <span v-if="job.failed > 0" class="back-class">
              {{ job.failed }} Failed
            </span>
          </el-form-item>
          <el-form-item label="选择器">
            <span v-if="!job.label_selector">—</span>
            <template v-else v-for="(val, key) in job.label_selector.matchLabels">
              <span :key="key">{{key}}: {{val}}<br/></span>
            </template>
          </el-form-item>
          <el-form-item label="标签">
            <span v-if="!job.labels">—</span>
            <template v-else v-for="(val, key) in job.labels">
              <span :key="key">{{key}}: {{val}}<br/></span>
            </template>
          </el-form-item>
          <!-- <el-form-item label="注解">
            <span v-if="!job.annotations">—</span>
            
            <template v-else v-for="(val, key) in job.annotations">
              <span :key="key">{{key}}: {{val}}<br/></span>
            </template>
          </el-form-item> -->
        </el-form>
      </div>

      <div style="padding: 0px 8px;">
        <div>Pods</div>
        <div class="msgClass" style="margin: 15px 10px 20px 10px;">
          <el-table
            ref="table"
            :data="pods"
            class="table-fix"
            tooltip-effect="dark"
            style="width: 100%"
            v-loading="loading"
            :cell-style="cellStyle"
            :default-sort = "{prop: 'name'}"
            >
            <el-table-column
              prop="name"
              label="名称"
              min-width="100"
              show-overflow-tooltip>
              <template slot-scope="scope">
                <span class="name-class" v-on:click="namePodClick(scope.row.namespace, scope.row.name)">
                  {{ scope.row.name }}
                </span>
              </template>
            </el-table-column>
            <el-table-column
              prop="containerNum"
              label="容器"
              min-width="45"
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
              prop="created"
              label="创建时间"
              min-width="100"
              show-overflow-tooltip>
            </el-table-column>
            <el-table-column
              prop="status"
              label="状态"
              min-width="60"
              show-overflow-tooltip>
            </el-table-column>
            <el-table-column
              label=""
              show-overflow-tooltip
              width="45">
              <template slot-scope="scope">
                <el-dropdown size="medium" >
                  <el-link :underline="false"><svg-icon style="width: 1.3em; height: 1.3em;" icon-class="operate" /></el-link>
                  <el-dropdown-menu slot="dropdown">
                    <el-dropdown-item v-if="$podOpPerm('get')" @click.native.prevent="namePodClick(scope.row.namespace, scope.row.name)">
                      <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                      <span style="margin-left: 5px;">详情</span>
                    </el-dropdown-item>
                    <div v-if="$podOpPerm('get')" @mouseover="logContainerShow = true;" @mouseout="logContainerShow = false;">
                      <el-dropdown-item @click.native.prevent="selectContainer = scope.row.containers[0].name; selectPodName = scope.row.name; logDialog = true;">
                        <div class="download">
                          <div>
                            <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="log" />
                            <span style="margin-left: 5px;">日志</span>
                          </div>
                          <div class="download-right" v-show="scope.row.containerNum > 1 && logContainerShow">
                            <div class="download-item" v-for="c in scope.row.init_containers" :key="c.name"
                                @click="selectContainer = c.name; selectPodName = scope.row.name; logDialog = true;">
                                {{ c.name }}
                            </div>
                            <div class="download-item" v-for="c in scope.row.containers" :key="c.name"
                                @click="selectContainer = c.name; selectPodName = scope.row.name; logDialog = true;">
                                {{ c.name }}
                            </div>
                          </div>
                        </div>
                      </el-dropdown-item>
                    </div>
                    <div v-if="$podOpPerm('update')" @mouseover="termContainerShow = true;" @mouseout="termContainerShow = false;">
                      <el-dropdown-item @click.native.prevent="selectContainer = scope.row.containers[0].name; selectPodName = scope.row.name; terminalDialog = true;">
                        <div class="download">
                          <div>
                            <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="terminal" />
                            <span style="margin-left: 5px;">终端</span>
                          </div>
                          <div class="download-right" v-show="scope.row.containers.length > 1 && termContainerShow">
                            <div class="download-item" v-for="c in scope.row.containers" :key="c.name"
                                @click="selectContainer = c.name; selectPodName = scope.row.name; terminalDialog = true;">
                                {{ c.name }}
                            </div>
                          </div>
                        </div>
                      </el-dropdown-item>
                    </div>
                    <el-dropdown-item v-if="$podOpPerm('delete')" @click.native.prevent="deletePods([{namespace: scope.row.namespace, name: scope.row.name}])">
                      <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                      <span style="margin-left: 5px;">删除</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </el-dropdown>
              </template>
            </el-table-column>
          </el-table>
        </div>
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
            :cell-style="cellStyle"
            :default-sort = "{prop: 'name'}"
            >
              <el-table-column type="expand" width="20" style="overflow:hidden">
                <template slot-scope="props">
                  <el-form label-position="left" inline class="table-expand">
                    <el-form-item label="容器名称">
                      <span>{{ props.row.name }}</span>
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
                min-width=""
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="image_pull_policy"
                label="镜像拉取策略"
                min-width=""
                show-overflow-tooltip>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
        <el-tab-pane label="存储" name="volumes">
          <div class="msgClass" style="padding: 10px 0px;">
            <template v-if="job.volumes && job.volumes.length > 0">
              <div v-for="v in job.volumes" :key="v.name" style="margin: 15px 25px; font-size: 14px; color: #606266">
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
            <div v-else style="padding: 25px 15px ; color: #909399; text-align: center">无挂载存储</div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="状态" name="conditions">
          <div class="msgClass">
            <el-table
              v-if="job && job.conditions && job.conditions.length > 0"
              :data="job.conditions"
              class="table-fix"
              tooltip-effect="dark"
              style="width: 100%"
              :cell-style="cellStyle"
              :default-sort = "{prop: 'lastProbeTime'}"
              >
              <el-table-column
                prop="type"
                label="类型"
                min-width="30"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="status"
                label="状态"
                min-width="20"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="reason"
                label="原因"
                min-width="50"
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
                min-width="40"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>
                    {{ scope.row.lastProbeTime ? scope.row.lastProbeTime : scope.row.lastTransitionTime }}
                  </span>
                </template>
              </el-table-column>
            </el-table>
            <div v-else style="padding: 25px 15px ; color: #909399; text-align: center">暂无数据</div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="事件" name="events">
          <div class="msgClass">
            <el-table
              v-if="jobEvents && jobEvents.length > 0"
              :data="jobEvents"
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
                min-width="25"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="object"
                label="对象"
                min-width="55"
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
                min-width="50"
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

      <el-dialog title="终端" :visible.sync="terminalDialog" :close-on-click-modal="false" width="80%" top="55px">
        <terminal v-if="terminalDialog" :cluster="cluster" :namespace="namespace" :pod="selectPodName" :container="selectContainer"></terminal>
      </el-dialog>

      <el-dialog title="日志" :visible.sync="logDialog" :close-on-click-modal="false" width="80%" top="55px">
        <log v-if="logDialog" :cluster="cluster" :namespace="namespace" :pod="selectPodName" :container="selectContainer"></log>
      </el-dialog>

      <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
        <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
        <span slot="footer" class="dialog-footer">
          <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
          <el-button plain @click="updateJob()" size="small">确 定</el-button>
        </span>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { getJob, deleteJobs, updateJob } from '@/api/job'
import { listEvents, buildEvent } from '@/api/event'
import { listPods, containerClass, buildPods, podMatch, deletePods, buildContainer, envStr, resourceFor } from '@/api/pods'
import { Message } from 'element-ui'
import { Terminal } from '@/views/components'
import { Log } from '@/views/components'

export default {
  name: 'JobDetail',
  components: {
    Clusterbar,
    Terminal,
    Log,
    Yaml
  },
  data() {
    return {
      logContainerShow: false,
      termContainerShow: false,
      yamlDialog: false,
      yamlValue: "",
      yamlLoading: true,
      logDialog: false,
      terminalDialog: false,
      cellStyle: {border: 0},
      loading: true,
      originJob: undefined,
      pods: [],
      selectContainer: '',
      selectPodName: '',
      jobEvents: [],
      eventLoading: true,
      envStr: envStr,
      resourceFor: resourceFor,
    }
  },
  created() {
    this.fetchData()
  },
  watch: {
    jobWatch: function (newObj) {
      if (newObj && this.originJob) {
        let newUid = newObj.resource.metadata.uid
        if (newUid !== this.job.uid) {
          return
        }
        let newRv = newObj.resource.metadata.resourceVersion
        if (this.job.resource_version < newRv) {
          this.originJob = newObj.resource
        }
      }
    },
    eventWatch: function (newObj) {
      if (newObj && this.originJob) {
        let event = newObj.resource
        if (event.involvedObject.namespace !== this.job.namespace) return
        if (event.involvedObject.uid !== this.job.uid) return
        let newUid = newObj.resource.metadata.uid
        if (newObj.event === 'add') {
          this.jobEvents.push(buildEvent(event))
        } else if (newObj.event == 'update') {
          let newRv = newObj.resource.metadata.resourceVersion
          for (let i in this.jobEvents) {
            let d = this.jobEvents[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newEvent = buildEvent(newObj.resource)
                this.$set(this.jobEvents, i, newEvent)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.jobEvents = this.jobEvents.filter(( { uid } ) => uid !== newUid)
        }
      }
    },
    podsWatch: function (newObj) {
      if (newObj && this.originJob) {
        let isPodMatch = podMatch(this.originJob.spec.selector, newObj.resource.metadata.labels)
        if (isPodMatch) {
          let newUid = newObj.resource.metadata.uid
          let newRv = newObj.resource.metadata.resourceVersion
          if (newObj.event === 'add') {
            this.pods.push(buildPods(newObj.resource))
          } else if (newObj.event === 'update') {
            for (let i in this.pods) {
              let p = this.pods[i]
              if (p.uid === newUid && p.resource_version < newRv) {
                let newPod = buildPods(newObj.resource)
                this.$set(this.pods, i, newPod)
                break
              }
            }
          } else if (newObj.event === 'delete') {
            this.pods = this.pods.filter(( { uid } ) => uid !== newUid)
          }
        }
      }
    }
  },
  computed: {
    titleName: function() {
      return ['Jobs', this.jobName]
    },
    jobName: function() {
      return this.$route.params ? this.$route.params.jobName : ''
    },
    namespace: function() {
      return this.$route.params ? this.$route.params.namespace : ''
    },
    job: function() {
      let p = this.buildJob(this.originJob)
      return p
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    jobWatch: function() {
      return this.$store.getters["ws/jobsWatch"]
    },
    eventWatch: function() {
      return this.$store.getters["ws/eventWatch"]
    },
    podsWatch: function() {
      return this.$store.getters["ws/podWatch"]
    },
    containers: function() {
      if (!this.originJob) return []
      let containers = []
      let tmpl = this.originJob.spec.template;
      for (let c of tmpl.spec.containers) {
        let bc = buildContainer(c)
        containers.push(bc)
      }
      let init_containers = []
      if (tmpl.spec.initContainers) {
        for (let c of tmpl.spec.initContainers) {
          init_containers.push(buildContainer(c))
        }
      }
      return [...init_containers, ...containers]
    },
  },
  methods: {
    fetchData: function() {
      this.originJob = null
      this.jobEvents = []
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
      if (!this.jobName) {
        Message.error("获取Job名称参数异常，请刷新重试")
        this.loading = false
        this.eventLoading = false
        return
      }
      getJob(cluster, this.namespace, this.jobName).then(response => {
        // this.loading = false
        this.originJob = response.data
        listPods(cluster, this.originJob.spec.selector).then(response => {
          this.loading = false
          this.pods = response.data
        }).catch(() => {
          this.loading = false
        })

        listEvents(cluster, this.originJob.metadata.uid).then(response => {
          this.eventLoading = false
          if (response.data) {
            this.jobEvents = response.data.length > 0 ? response.data : []
          }
        }).catch(() => {
          this.eventLoading = false
        })
      }).catch(() => {
        this.loading = false
        this.eventLoading = false
      })
    },
    buildJob: function(job) {
      if (!job) return {}
      let p = {
        uid: job.metadata.uid,
        namespace: job.metadata.namespace,
        name: job.metadata.name,
        completions: job.spec.completions || 0,
        active: job.status.active || 0,
        succeeded: job.status.succeeded || 0,
        failed: job.status.failed || 0,
        resource_version: job.metadata.resourceVersion,
        conditions: job.status.conditions,
        created: job.metadata.creationTimestamp,
        label_selector: job.spec.selector,
        labels: job.metadata.labels,
        annotations: job.metadata.annotations,
        volumes: job.spec.template.spec.volumes,
      }
      return p
    },
    toogleExpand: function(row) {
      let $table = this.$refs.table;
      $table.toggleRowExpansion(row)
    },
    deleteJobs: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( !this.job ) {
        Message.error("获取Job参数异常，请刷新重试")
      }
      let jobs = [{
        namespace: this.job.namespace,
        name: this.job.name,
      }]
      let params = {
        resources: jobs
      }
      deleteJobs(cluster, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    getJobYaml: function() {
      if (!this.job) {
        Message.error("获取Job参数异常，请刷新重试")
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
      getJob(cluster, this.job.namespace, this.job.name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    updateJob: function() {
      if (!this.job) {
        Message.error("获取Job参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      updateJob(cluster, this.job.namespace, this.job.name, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    containerClass: function(status) {
      return containerClass(status)
    },
    namePodClick: function(namespace, name) {
      this.$router.push({name: 'podsDetail', params: {namespace: namespace, podName: name}})
    },
    deletePods: function(pods) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( pods.length <= 0 ){
        Message.error("请选择要删除的Pod")
        return
      }
      let params = {
        resources: pods
      }
      deletePods(cluster, params).then(() => {
        Message.success("删除成功")
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
.msgClass {
  margin: 6px 6px;
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
  margin: 20px 5px 30px 5px;
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
/* .msgClass {
  margin: 0px 25px;
} */
.msgClass .el-table::before {
  height: 0px;
}
</style>
