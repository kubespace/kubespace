<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc"
      :createFunc="createFunc" createDisplay="创建"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="daemonsets"
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
          min-width="70"
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
          min-width="45"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="ready_replicas"
          label="Pods"
          min-width="30"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ scope.row.number_ready }}/{{ scope.row.desired_number_scheduled }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="strategy"
          label="更新策略"
          min-width="45"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="node_selector"
          label="节点选择"
          min-width="70"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.node_selector">
              <span v-for="(val, key) in scope.row.node_selector" :key="key" class="back-class">
                {{ key + '=' + val }} 
              </span>
            </template>
            <!-- <span v-else>--</span> -->
          </template>
        </el-table-column>
        <el-table-column
          prop="conditions"
          label="状态"
          min-width="40"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.conditions && scope.row.conditions.length > 0">
              <span v-for="c in scope.row.conditions" :key="c">
                {{ c }}
              </span>
            </template>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="created"
          label="创建时间"
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
                <el-dropdown-item @click.native.prevent="nameClick(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$updatePerm()" @click.native.prevent="getDaemonSetYaml(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$deletePerm()" @click.native.prevent="deleteDaemonSets([{namespace: scope.row.namespace, name: scope.row.name}])">
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
        <el-button plain @click="updateDaemonSet()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listDaemonSets, getDaemonSet, deleteDaemonSets, updateDaemonSet } from '@/api/daemonset'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'

export default {
  name: 'DaemonSet',
  components: {
    Clusterbar,
    Yaml
  },
  data() {
      return {
        yamlDialog: false,
        yamlNamespace: "",
        yamlName: "",
        yamlValue: "",
        yamlLoading: true,
        cellStyle: {border: 0},
        titleName: ["DaemonSets"],
        maxHeight: window.innerHeight - 150,
        loading: true,
        originDaemonSets: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delDaemonSets: [],
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
        // console.log(heightStyle)
        that.maxHeight = heightStyle
      })()
    }
  },
  watch: {
    daemonsetsWatch: function (newObj) {
      if (newObj) {
        let newUid = newObj.resource.metadata.uid
        let newRv = newObj.resource.metadata.resourceVersion
        if (newObj.event === 'add') {
          this.originDaemonSets.push(this.buildDaemonSets(newObj.resource))
        } else if (newObj.event === 'update') {
          for (let i in this.originDaemonSets) {
            let d = this.originDaemonSets[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newDp = this.buildDaemonSets(newObj.resource)
                this.$set(this.originDaemonSets, i, newDp)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originDaemonSets = this.originDaemonSets.filter(( { uid } ) => uid !== newUid)
        }
      }
    }
  },
  computed: {
    daemonsets: function() {
      let dlist = []
      for (let p of this.originDaemonSets) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        if (p.conditions && p.conditions.length > 0) {
          p.conditions.sort()
        } else {
          p.conditions = []
        }
        dlist.push(p)
      }
      return dlist
    },
    daemonsetsWatch: function() {
      return this.$store.getters["ws/daemonsetsWatch"]
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originDaemonSets = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listDaemonSets(cluster).then(response => {
          this.loading = false
          this.originDaemonSets = response.data
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.loading = false
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
    buildDaemonSets: function(daemonset) {
      if (!daemonset) return
      var conditions = []
      if(daemonset.status.conditions) {
        for (let c of daemonset.status.conditions) {
          if (c.status === "True") {
            conditions.push(c.type)
          }
        }
      }
      let p = {
        uid: daemonset.metadata.uid,
        namespace: daemonset.metadata.namespace,
        name: daemonset.metadata.name,
        desired_number_scheduled: daemonset.status.desiredNumberScheduled || 0,
        number_ready: daemonset.status.numberReady || 0,
        resource_version: daemonset.metadata.resourceVersion,
        strategy: daemonset.spec.updateStrategy.type,
        conditions: conditions,
        node_selector: daemonset.spec.template.spec.nodeSelector,
        created: daemonset.metadata.creationTimestamp
      }
      return p
    },
    nameClick: function(namespace, name) {
      this.$router.push({name: 'daemonsetDetail', params: {namespace: namespace, daemonsetName: name}})
    },
    getDaemonSetYaml: function(namespace, name) {
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
      if (!name) {
        Message.error("获取Deployment名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getDaemonSet(cluster, namespace, name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteDaemonSets: function(daemonsets) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( daemonsets.length <= 0 ){
        Message.error("请选择要删除的DaemonSets")
        return
      }
      let params = {
        resources: daemonsets
      }
      deleteDaemonSets(cluster, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    updateDaemonSet: function() {
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
        Message.error("获取DaemonSet参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      updateDaemonSet(cluster, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delDaemonSetsFunc: function() {
      if (this.delDaemonSets.length > 0){
        let delDaemonSets = []
        for (var p of this.delDaemonSets) {
          delDaemonSets.push({namespace: p.namespace, name: p.name})
        }
        this.deleteDaemonSets(delDaemonSets)
      }
    },
    handleSelectionChange(val) {
      this.delDaemonSets = val;
      if (val.length > 0){
        this.delFunc = this._delDaemonSetsFunc
      } else {
        this.delFunc = undefined
      }
    },
    createFunc() {
      this.$router.push({name: 'daemonsetCreate'})
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
</style>
