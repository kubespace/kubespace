<template>
  <div>
    <clusterbar :titleName="titleName"/>
    <div class="dashboard-container" v-loading="loading">
      <!-- <div class="dashboard-text"></div> -->
      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" class="pod-item" label-width="120px" style="margin: 15px 10px 20px 10px;">
          <el-form-item label="名称">
            <span>{{ appDetail.name }}</span>
          </el-form-item>
          <el-form-item label="命名空间">
            <span>{{ appDetail.namespace }}</span>
          </el-form-item>
          <el-form-item label="Chart">
            <span>{{ appDetail.chart_name }}</span>
          </el-form-item>
          <el-form-item label="应用版本">
            <span>{{ appDetail.app_version }}</span>
          </el-form-item>
          <el-form-item label="上次部署">
            <span>{{ dateFormat(appDetail.last_deployed) }}</span>
          </el-form-item>
        </el-form>
      </div>

      <div style="padding: 0px 8px 0px;" v-if="appResources['Ingress']">
        <div>Ingress</div>
        <div class="msgClass" style="margin: 15px 10px 20px 10px;">
            <el-table
              ref="table"
              :data="appResources['Ingress']"
              class="table-fix"
              tooltip-effect="dark"
              style="width: 100%"
              :cell-style="cellStyle"
              :default-sort = "{prop: 'metadata.name'}"
              >
              <el-table-column
                prop="metadata.name"
                label="名称"
                min-width="20"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <template v-if="true">
                    <span class="name-class" v-on:click="nameClick(scope.row.kind, scope.row.metadata.name)">
                      {{ scope.row.metadata.name }}
                    </span>
                  </template>
                  <template v-else>
                    <span class="name-class">
                      {{ scope.row.name }}
                    </span>
                  </template>
                </template>
              </el-table-column>
              <el-table-column
                prop="restarts"
                label="域名"
                min-width="45"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span v-for="h of scope.row.spec.rules" :key="h.host" class="back-class">
                    {{ h.host }}
                  </span>
                </template>
              </el-table-column>
            </el-table>
          </div>
      </div>

      <div style="padding: 0px 8px 0px;" v-if="appResources['Service']">
        <div>Service</div>
        <div class="msgClass" style="margin: 15px 10px 15px 10px;">
            <el-table
              ref="table"
              :data="appResources['Service']"
              class="table-fix"
              tooltip-effect="dark"
              style="width: 100%"
              :cell-style="cellStyle"
              :default-sort = "{prop: 'metadata.name'}"
              >
              <el-table-column
                prop="metadata.name"
                label="名称"
                min-width="24"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <template v-if="true">
                    <span class="name-class" v-on:click="nameClick(scope.row.kind, scope.row.metadata.name)">
                      {{ scope.row.metadata.name }}
                    </span>
                  </template>
                  <template v-else>
                    <span class="name-class">
                      {{ scope.row.name }}
                    </span>
                  </template>
                </template>
              </el-table-column>
              <el-table-column
                prop="type"
                label="类型"
                min-width="10"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>
                    {{ scope.row.spec.type || 'ClusterIP' }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column
                prop="ports"
                label="端口"
                min-width="13"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>{{ getPortsDisplay(scope.row.spec.ports) }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="selector"
                label="选择器"
                min-width="55"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <template v-if="scope.row.spec.selector">
                    <span v-for="(val, key) in scope.row.spec.selector" :key="key" class="back-class">
                      {{ key + '=' + val }} 
                    </span>
                  </template>
                  <!-- <span v-else>--</span> -->
                </template>
              </el-table-column>
            </el-table>
          </div>
      </div>

      <template v-if="appResources['Deployment'] || appResources['StatefulSet'] || appResources['DaemonSet'] || appResources['CronJob'] || appResources['Job']">
        <el-tabs :value="defaultWorkloadName" style="padding: 0px 8px 0px;">
          <template v-for="kind of ['Deployment', 'StatefulSet', 'DaemonSet', 'CronJob', 'Job']">
            <el-tab-pane :label="kind" :name="kind" :key="kind" v-if="appResources[kind]">
              <div class="msgClass">
                <el-table
                  ref="table"
                  :data="appResources[kind]"
                  class="table-fix"
                  tooltip-effect="dark"
                  style="width: 100%"
                  :cell-style="cellStyle"
                  :default-sort = "{prop: 'metadata.name'}"
                  >
                  <el-table-column
                    prop="metadata.name"
                    label="名称"
                    min-width="28"
                    show-overflow-tooltip>
                    <template slot-scope="scope">
                      <template v-if="true">
                        <span class="name-class" v-on:click="nameClick(scope.row.kind, scope.row.metadata.name)">
                          {{ scope.row.metadata.name }}
                        </span>
                      </template>
                      <template v-else>
                        <span class="name-class">
                          {{ scope.row.name }}
                        </span>
                      </template>
                    </template>
                  </el-table-column>
                  <el-table-column v-if="kind != 'CronJob' && kind != 'DaemonSet'"
                    prop="type"
                    label="副本"
                    min-width="10"
                    show-overflow-tooltip>
                    <template slot-scope="scope">
                      <span>
                        {{ scope.row.spec.replicas || 1 }}
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column v-if="kind == 'CronJob'"
                    prop="type"
                    label="定时"
                    min-width="10"
                    show-overflow-tooltip>
                    <template slot-scope="scope">
                      <span>
                        {{ scope.row.spec.schedule }}
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column v-if="kind != 'CronJob'"
                    prop="selector"
                    label="POD选择器"
                    min-width="75"
                    show-overflow-tooltip>
                    <template slot-scope="scope">
                      <template v-if="scope.row.spec.selector.matchLabels">
                        <span v-for="(val, key) in scope.row.spec.selector.matchLabels" :key="key" class="back-class">
                          {{ key + '=' + val }} 
                        </span>
                      </template>
                      <!-- <span v-else>--</span> -->
                    </template>
                  </el-table-column>
                  <el-table-column v-else
                    prop="selector"
                    label="POD选择器"
                    min-width="75"
                    show-overflow-tooltip>
                    <template slot-scope="scope">
                      <template v-if="scope.row.spec.jobTemplate.spec.selector.matchLabels">
                        <span v-for="(val, key) in scope.row.spec.jobTemplate.spec.selector.matchLabels" :key="key" class="back-class">
                          {{ key + '=' + val }} 
                        </span>
                      </template>
                      <!-- <span v-else>--</span> -->
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-tab-pane>
          </template>
        </el-tabs>
      </template>

      <template v-if="appResources['ConfigMap'] || appResources['Secret']">
        <el-tabs value="ConfigMap" style="padding: 0px 8px 0px;">
          <template v-for="kind of ['ConfigMap', 'Secret']">
            <el-tab-pane :label="kind" :name="kind" :key="kind" v-if="appResources[kind]">
              <div class="msgClass">
                <el-table
                  ref="table"
                  :data="appResources[kind]"
                  class="table-fix"
                  tooltip-effect="dark"
                  style="width: 100%"
                  :cell-style="cellStyle"
                  :default-sort = "{prop: 'metadata.name'}"
                  >
                  <el-table-column
                    prop="metadata.name"
                    label="名称"
                    min-width="24"
                    show-overflow-tooltip>
                    <template slot-scope="scope">
                      <template v-if="true">
                        <span class="name-class" v-on:click="nameClick(scope.row.kind, scope.row.metadata.name)">
                          {{ scope.row.metadata.name }}
                        </span>
                      </template>
                      <template v-else>
                        <span class="name-class">
                          {{ scope.row.name }}
                        </span>
                      </template>
                    </template>
                  </el-table-column>
                  <el-table-column
                    prop="type"
                    label="Keys"
                    min-width="44"
                    show-overflow-tooltip>
                    <template slot-scope="scope">
                      <span v-for="(k, v) of scope.row.data" :key="v" class="back-class">
                        {{ v }}
                      </span>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-tab-pane>
          </template>
        </el-tabs>
      </template>

      <template v-if="appResources['PersistentVolumeClaim'] || appResources['PersistentVolume']">
        <el-tabs value="PersistentVolumeClaim" style="padding: 0px 8px 0px;">
          <template v-for="kind of ['PersistentVolumeClaim', 'PersistentVolume']">
            <el-tab-pane :label="kind" :name="kind" :key="kind" v-if="appResources[kind]">
              <div class="msgClass">
                <el-table
                  ref="table"
                  :data="appResources[kind]"
                  class="table-fix"
                  tooltip-effect="dark"
                  style="width: 100%"
                  :cell-style="cellStyle"
                  :default-sort = "{prop: 'metadata.name'}"
                  >
                  <el-table-column
                    prop="metadata.name"
                    label="名称"
                    min-width="34"
                    show-overflow-tooltip>
                    <template slot-scope="scope">
                      <template v-if="true">
                        <span class="name-class" v-on:click="nameClick(scope.row.kind, scope.row.metadata.name)">
                          {{ scope.row.metadata.name }}
                        </span>
                      </template>
                      <template v-else>
                        <span class="name-class">
                          {{ scope.row.name }}
                        </span>
                      </template>
                    </template>
                  </el-table-column>
                  <el-table-column
                    prop="type"
                    label="容量"
                    min-width="44"
                    show-overflow-tooltip>
                    <template slot-scope="scope">
                      <span>
                        {{ scope.row.spec.resources.requests }}
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column
                    prop="type"
                    label="访问模式"
                    min-width="44"
                    show-overflow-tooltip>
                    <template slot-scope="scope">
                      <span>
                        {{ scope.row.spec.accessModes }}
                      </span>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-tab-pane>
          </template>
        </el-tabs>
      </template>
    </div>
    <el-dialog title="升级" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="45px"
      @close="yamlDialog=false; updateValues={name: '', namespace: '', config: '', values: ''}; yamlChange=1;">
      <el-button-group style="margin-bottom: 10px;">
        <el-button :type="yamlChange ? 'primary': ''" size="small" @click="yamlChange=1;  ">当前配置</el-button>
        <el-button :type="yamlChange ? '': 'primary'" size="small" @click="yamlChange=0;  yamlValue=updateValues.values">原始values(只读)</el-button>
      </el-button-group>
      <template v-if="yamlDialog">
        <yaml v-show="yamlChange" v-model="updateValues.config" :loading="yamlLoading" :readOnly="false"></yaml>
        <yaml v-show="!yamlChange" v-model="yamlValue" :loading="yamlLoading" :readOnly="true"></yaml>
      </template>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
        <el-button plain @click="updateNode()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listReleases, deleteRelease, getRelease } from '@/api/app'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'
import { dateFormat } from '@/utils/utils'
let yaml = require('js-yaml')

export default {
  name: 'Application',
  components: {
    Clusterbar,
    Yaml
  },
  data() {
      return {
        yamlDialog: false,
        yamlName: "",
        yamlValue: "",
        yamlLoading: true,
        cellStyle: {border: 0},
        maxHeight: window.innerHeight - 150,
        loading: true,
        originReleases: [],
        search_ns: [],
        search_name: '',
        yamlChange: 1,
        appDetail: {},
        appResources: {},
        ingresses: [],
        services: [],
        deployments: [],
        statefulsets: [],
        daemonsets: [],
        jobs: [],
        configmaps: [],
        secrets: []
      }
  },
  created() {
    this.fetchData()
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 150
        that.maxHeight = heightStyle
      })()
    }
  },
  watch: {
  },
  computed: {
    titleName: function() {
      return ['Applications', this.appName]
    },
    appName: function() {
      return this.$route.params ? this.$route.params.appName : ''
    },
    namespace: function() {
      return this.$route.params ? this.$route.params.namespace : ''
    },
    defaultWorkloadName: function() {
      for(let k of ['Deployment', 'StatefulSet', 'DaemonSet', 'CronJob', 'Job']) {
        if(this.appResources[k]) return k
      }
      return 'Deployment'
    }
  },
  methods: {
    dateFormat,
    fetchData: function() {
      this.loading = true
      const cluster = this.$store.state.cluster
      if (cluster) {
        var params = {
          name: this.appName,
          namespace: this.namespace,
        }
        getRelease(cluster, params).then(response => {
          console.log(response.data)
          this.appDetail = response.data
          this.parseResource()
          this.loading = false
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
    parseResource: function() {
      if(!(this.appDetail && this.appDetail.manifest)) return {}
      var res = []
      try{
        var res = yaml.loadAll(this.appDetail.manifest, {
          onWarning: function(e) {
            console.log(e)
          }
        });
      }catch(e) {
        console.log(e)
        Message.error("解析失败：" + e)
      }
      for(let r of res) {
        if(this.appResources[r.kind]) {
          this.appResources[r.kind].push(r)
        } else {
          this.appResources[r.kind] = [r]
        }
      }
      console.log(this.appResources)
      return res
    },
    getUpdateRelease: function(name, namespace) {
      this.yamlLoading = true
      const cluster = this.$store.state.cluster
      this.updateValues = {config: "", values: "", name: name, namespace: namespace}
      this.yamlValue = ""
      if (cluster) {
        var params = {
          name: name,
          namespace: namespace,
          get_option: 'values'
        }
        getRelease(cluster, params).then(response => {
          this.yamlLoading = false
          this.updateValues.config = response.data.config
          this.updateValues.values = response.data.values
          this.yamlValue = this.updateValues.config
        }).catch(() => {
          this.yamlLoading = false
        })
      } else {
        this.yamlLoading = false
        Message.error("获取集群异常，请刷新重试")
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
    buildNodes: function(node) {
      if (!node) return
      let p = {
        uid: node.metadata.uid,
        name: node.metadata.name,
        version: node.status.nodeInfo.kubeletVersion,
        taints: node.spec.taints ? node.spec.taints.length : 0,
        resource_version: node.metadata.resourceVersion,
        created: node.metadata.creationTimestamp
      }
      return p
    },
    nameClick: function(kind, name) {
      if(kind == 'Ingress') {
        this.$router.push({name: 'ingressDetail', params: {namespace: this.appDetail.namespace, ingressName: name}})
        return
      }
      if(kind == 'Service') {
        this.$router.push({name: 'serviceDetail', params: {namespace: this.appDetail.namespace, serviceName: name}})
        return
      }
      if(kind == 'Deployment') {
        this.$router.push({name: 'deploymentDetail', params: {namespace: this.appDetail.namespace, deploymentName: name}})
        return
      }
      if(kind == 'StatefulSet') {
        this.$router.push({name: 'statefulsetDetail', params: {namespace: this.appDetail.namespace, statefulsetName: name}})
        return
      }
      if(kind == 'DaemonSet') {
        this.$router.push({name: 'daemonsetDetail', params: {namespace: this.appDetail.namespace, daemonsetName: name}})
        return
      }
      if(kind == 'CronJob') {
        this.$router.push({name: 'cronjobDetail', params: {namespace: this.appDetail.namespace, cronjobName: name}})
        return
      }
      if(kind == 'Job') {
        this.$router.push({name: 'jobDetail', params: {namespace: this.appDetail.namespace, jobName: name}})
        return
      }
      if(kind == 'ConfigMap') {
        this.$router.push({name: 'configMapDetail', params: {namespace: this.appDetail.namespace, configMapName: name}})
        return
      }
      if(kind == 'Secret') {
        this.$router.push({name: 'secretDetail', params: {namespace: this.appDetail.namespace, secretName: name}})
        return
      }
      if(kind == 'PersistentVolumeClaim') {
        this.$router.push({name: 'pvcDetail', params: {namespace: this.appDetail.namespace, persistentVolumeClaimName: name}})
        return
      }
      Message.error("Not found kind detail")
    },
    getNodeYaml: function(name) {
      this.yamlName = ""
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!name) {
        Message.error("获取Node名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getNode(cluster, name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteRelease: function(release) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      let params = {
        name: release.name,
        namespace: release.namespace,
      }
      this.$confirm('是否确认删除应用' + release.name + '?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        deleteRelease(cluster, params).then(() => {
          Message.success("删除成功")
          this.fetchData();
        }).catch(() => {
          // console.log(e)
        })
      })
    },
    updateNode: function() {
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
        Message.error("获取Node参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      updateNode(cluster, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delNodesFunc: function() {
      if (this.delNodes.length > 0){
        let delNodes = []
        for (var p of this.delNodes) {
          delNodes.push({namespace: p.namespace, name: p.name})
        }
        this.deleteNodes(delNodes)
      }
    },
    handleSelectionChange(val) {
      this.delNodes = val;
      if (val.length > 0){
        this.delFunc = this._delNodesFunc
      } else {
        this.delFunc = undefined
      }
    },
    createFunc() {
      this.$router.push({name: 'appCreate'})
    },
    getPortsDisplay(ports) {
      if (!ports) return ''
      var pd = []
      for (let p of ports) {
        var pds = p.port
        if (p.nodePort) {
          pds += ':' + p.nodePort
        }
        if(p.protocol) pds += '/' + p.protocol
        pd.push(pds)
      }
      return pd.join(',')
    }
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
.msgClass {
  margin: 8px 10px 15px 10px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

</style>

<style lang="scss">
.el-dialog__body {
  padding-top: 5px;
  padding-bottom: 5px;
}
.replicaDialog {
  .el-form-item {
    margin-bottom: 10px;
  }
  .el-dialog--center .el-dialog__body {
    padding: 5px 25px;
  }
}
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

.pod-item {
  margin: 20px 5px 30px 5px;
  padding: 10px 20px;
  font-size: 0;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}
.pod-item label {
  /* width: 120px; */
  color: #99a9bf;
  font-weight: 400;
  /* display: inline-block; */
}
.pod-item .el-form-item {
  margin-right: 0;
  margin-bottom: 0;
  /* width: 50%; */
}
/* .pod-item .el-form-item__content{
  float: left;
} */
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
