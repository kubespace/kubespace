<template>
  <div>
    <clusterbar :titleName="titleName" :delFunc="deleteServices" :editFunc="getServiceYaml" :titleLink="[titleLink]"/>
    <div class="dashboard-container detail-dashboard" v-loading="loading" :style="{'max-height': maxHeight + 'px', overflow: 'auto'}">
      <div style="padding: 10px 0px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" class="pod-item pod-item-all" label-width="120px">
          <el-form-item label="名称">
            <span>{{ service.name }}</span>
          </el-form-item>
          <el-form-item label="创建时间">
            <span>{{ $dateFormat(service.created) }}</span>
          </el-form-item>
          <el-form-item label="命名空间">
            <span>{{ service.namespace }}</span>
          </el-form-item>
          <el-form-item label="类型">
            <span>{{ service.type }}</span>
          </el-form-item>
          <el-form-item label="Cluster IP">
            <span>{{ service.cluster_ip }}</span>
          </el-form-item>
          <el-form-item label="端口">
            <template v-if="service.ports && service.ports.length > 0">
              <span>{{ getPortsDisplay(service.ports) }}</span>
            </template>
          </el-form-item>
          <el-form-item label="Endpoints">
            <template v-for="e of endpoints">
              <span :key="e.name">{{ endpointsAddresses(e.subsets) }}</span>
            </template>
          </el-form-item>
          <el-form-item label="会话亲和">
            <span>{{ service.session_affinity }}</span>
          </el-form-item>
          <el-form-item label="选择器">
            <template v-if="service.selector">
              <span v-for="(val, key) in service.selector" :key="key" class="back-class">
                {{ key + ': ' + val }} <br/>
              </span>
            </template>
          </el-form-item>
          <el-form-item label="标签">
            <span v-if="!service.labels">——</span>
            <template v-else v-for="(val, key) in service.labels" >
              <span :key="key" class="back-class">{{key}}: {{val}} </span>
            </template>
          </el-form-item>
          <!-- <el-form-item label="注解">
            <span v-if="!service.annotations">——</span>
            
            <template v-else v-for="(val, key) in service.annotations">
              <span :key="key">{{key}}: {{val}}<br/></span>
            </template>
          </el-form-item> -->
        </el-form>
      </div>

      <div>
        <div>Pods</div>
        <div class="msgClass">
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
              min-width="150"
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
                    <el-dropdown-item @click.native.prevent="namePodClick(scope.row.namespace, scope.row.name)">
                      <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                      <span style="margin-left: 5px;">详情</span>
                    </el-dropdown-item>
                    <div @mouseover="logContainerShow = true;" @mouseout="logContainerShow = false;">
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
                    <div @mouseover="termContainerShow = true;" @mouseout="termContainerShow = false;">
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
                    <el-dropdown-item @click.native.prevent="deletePods([{namespace: scope.row.namespace, name: scope.row.name}])">
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
          <el-button plain @click="updateService()" size="small">确 定</el-button>
        </span>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { ResType, listResource, getResource, delResource, watchResource, updateResource,
         containerClass, podMatch } from '@/api/cluster/resource'
import { Message } from 'element-ui'
import { Terminal } from '@/views/components'
import { Log } from '@/views/components'

export default {
  name: 'ServiceDetail',
  components: {
    Clusterbar,
    Terminal,
    Log,
    Yaml
  },
  data() {
    return {
      maxHeight: window.innerHeight - this.$contentHeight,
      logContainerShow: false,
      termContainerShow: false,
      yamlDialog: false,
      yamlValue: "",
      yamlLoading: true,
      cellStyle: {border: 0},
      loading: true,
      originService: undefined,
      pods: [],
      endpoints: [],
      selectContainer: '',
      selectPodName: '',
      logDialog: false,
      terminalDialog: false,
      containerClass: containerClass,
      selectContainer: '',
      selectPodName: '',
      podSSE: undefined
    }
  },
  created() {
    this.fetchData()
  },
  beforeDestroy() {
    if(this.podSSE) this.podSSE.disconnect()
  },
  computed: {
    titleLink() {
      if(this.projectId) {
        return 'workspaceServices'
      }
      return 'services'
    },
    titleName: function() {
      return ['Services', this.serviceName]
    },
    serviceName: function() {
      return this.$route.params ? this.$route.params.serviceName : ''
    },
    namespace: function() {
      return this.$route.params ? this.$route.params.namespace : ''
    },
    service: function() {
      let p = this.buildService(this.originService)
      return p
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    projectId() {
      return this.$route.params.workspaceId
    },
  },
  methods: {
    fetchData: function() {
      this.originService = null
      // this.serviceEvents = []
      this.loading = true
      const cluster = this.$store.state.cluster
      if (!cluster) {
        if(this.projectId) return
        Message.error("获取集群参数异常，请刷新重试")
        this.loading = false
        return
      }
      if (!this.namespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        this.loading = false
        return
      }
      if (!this.serviceName) {
        Message.error("获取Service名称参数异常，请刷新重试")
        this.loading = false
        return
      }
      getResource(cluster, ResType.Service, this.namespace, this.serviceName, '', {project_id: this.projectId}).then(response => {
        // this.loading = false
        this.originService = response.data
        getResource(cluster, ResType.Endpoint, this.namespace, this.serviceName, '', {project_id: this.projectId}).then(response => {
          // this.loading = false
          this.endpoints = response.data ? [response.data] : []
          let pod_names = []
          for(let p of this.endpoints) {
            for(let e of p.subsets) {
              if(e.addresses){
                for(let a of e.addresses) {
                  if(a.targetRef && a.targetRef.kind === 'Pod') {
                    pod_names.push(a.targetRef.name)
                  }
                }
              }
            }
          }
          if(pod_names.length > 0) {
            let params = {
              namespace: this.namespace,
              label_selector: {
                matchLabels: this.originService.spec.selector
              }
            }
            listResource(cluster, ResType.Pod, params, {project_id: this.projectId}).then(response => {
              this.loading = false
              this.pods = response.data
              this.fetchPodSSE()
            }).catch(() => {
              this.loading = false
            })
          } else {
            this.loading = false
          }
        }).catch(() => {
          this.loading = false
        })

      }).catch(() => {
        this.loading = false
      })
    },
    fetchPodSSE() {
      let params = {
        namespace: this.namespace,
        label_selector: {
          matchLabels: this.originService.spec.selector
        },
        process: true,
        project_id: this.projectId
      }
      this.podSSE = watchResource(this.$sse, this.cluster, ResType.Pod, this.podsWatchFunc, params)
    },
    podsWatchFunc: function (newObj) {
      if (newObj && this.originService) {
        let isPodMatch = podMatch(this.originService.spec.selector, newObj.resource.labels)
        if (isPodMatch) {
          let newUid = newObj.resource.uid
          let newRv = newObj.resource.resource_version
          if (newObj.event === 'add') {
            for (let i in this.pods) {
              let p = this.pods[i]
              if (p.uid === newUid) return
            }
            this.pods.push(newObj.resource)
          } else if (newObj.event === 'update') {
            for (let i in this.pods) {
              let p = this.pods[i]
              if (p.uid === newUid && p.resource_version < newRv) {
                let newPod = newObj.resource
                this.$set(this.pods, i, newPod)
                break
              }
            }
          } else if (newObj.event === 'delete') {
            this.pods = this.pods.filter(( { uid } ) => uid !== newUid)
          }
        }
      }
    },
    buildService: function(service) {
      if (!service) return {}
      let p = {
        uid: service.metadata.uid,
        namespace: service.metadata.namespace,
        name: service.metadata.name,
        type: service.spec.type,
        cluster_ip: service.spec.clusterIP,
        ports: service.spec.ports,
        external_ip: service.spec.externalIPs,
        session_affinity: service.spec.sessionAffinity,
        resource_version: service.metadata.resourceVersion,
        selector: service.spec.selector,
        created: service.metadata.creationTimestamp,
        labels: service.metadata.labels,
        annotations: service.metadata.annotations,
      }
      return p
    },
    toogleExpand: function(row) {
      let $table = this.$refs.table;
      $table.toggleRowExpansion(row)
    },
    deleteServices: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( !this.service ) {
        Message.error("获取Service参数异常，请刷新重试")
      }
      let services = [{
        namespace: this.service.namespace,
        name: this.service.name,
      }]
      let params = {
        resources: services
      }
      delResource(cluster, ResType.Service, params, {project_id: this.projectId}).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    getServiceYaml: function() {
      if (!this.service) {
        Message.error("获取Service参数异常，请刷新重试")
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
      getResource(cluster, ResType.Service, this.service.namespace, this.service.name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    updateService: function() {
      if (!this.service) {
        Message.error("获取Service参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      updateResource(cluster, ResType.Service, this.service.namespace, this.service.name, this.yamlValue, {project_id: this.projectId}).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    getPortsDisplay(ports) {
      if (!ports) return ''
      var pd = []
      for (let p of ports) {
        var pds = p.port
        if (p.nodePort) {
          pds += ':' + p.nodePort
        }
        pds += '/' + p.protocol
        pd.push(pds)
      }
      return pd.join(', ')
    },
    endpointsAddresses(subsets) {
      if (!subsets) return ''
      let as = []
      for(let s of subsets) {
        if(s.addresses) {
          for(let a of s.addresses) {
            if(s.ports) {
              for(let e of s.ports) {
                as.push(a.ip + ':' + e.port)
              }
            } else {
              as.push(a.ip)
            }
          }
        }
      }
      return as.join(', ')
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
      delResource(cluster, ResType.Pod, params, {project_id: this.projectId}).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    namePodClick: function(namespace, name) {
      this.$router.push({name: 'podsDetail', params: {namespace: namespace, podName: name}})
    },
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
/* 
.item-class {
  padding: 20px 20px 20px 5px;
  font-size: 0;
}

.item-class  */

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
</style>
