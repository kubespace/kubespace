<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="networkpolicies"
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
          min-width="30"
          show-overflow-tooltip>
          <!-- <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.namespace, scope.row.name)">
              {{ scope.row.name }}
            </span>
          </template> -->
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="30"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="policy_types"
          label="策略类型"
          min-width="35"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="created"
          label="创建时间"
          min-width="30"
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
                <!-- <el-dropdown-item @click.native.prevent="nameClick(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item> -->
                <el-dropdown-item v-if="$updatePerm()" @click.native.prevent="getNetworkPolicyYaml(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$deletePerm()" @click.native.prevent="deleteNetworkPolicies([{namespace: scope.row.namespace, name: scope.row.name}])">
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
        <el-button plain @click="updateNetworkPolicy()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listNetworkPolicies, getNetworkPolicy, deleteNetworkPolicies, updateNetworkPolicy } from '@/api/networkpolicy'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'

export default {
  name: 'NetworkPolicy',
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
        titleName: ["Network Policies"],
        maxHeight: window.innerHeight - 150,
        loading: true,
        originNetworkPolicies: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delNetworkPolicies: [],
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
    networkpoliciesWatch: function (newObj) {
      if (newObj) {
        let newUid = newObj.resource.metadata.uid
        let newRv = newObj.resource.metadata.resourceVersion
        if (newObj.event === 'add') {
          this.originNetworkPolicies.push(this.buildNetworkPolicies(newObj.resource))
        } else if (newObj.event === 'update') {
          for (let i in this.originNetworkPolicies) {
            let d = this.originNetworkPolicies[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newDp = this.buildNetworkPolicies(newObj.resource)
                this.$set(this.originNetworkPolicies, i, newDp)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originNetworkPolicies = this.originNetworkPolicies.filter(( { uid } ) => uid !== newUid)
        }
      }
    }
  },
  computed: {
    networkpolicies: function() {
      let dlist = []
      for (let p of this.originNetworkPolicies) {
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
    networkpoliciesWatch: function() {
      return this.$store.getters["ws/networkpoliciesWatch"]
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originNetworkPolicies = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listNetworkPolicies(cluster).then(response => {
          this.loading = false
          this.originNetworkPolicies = response.data ? response.data : []
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
    buildNetworkPolicies: function(networkpolicy) {
      if (!networkpolicy) return
      let p = {
        uid: networkpolicy.metadata.uid,
        namespace: networkpolicy.metadata.namespace,
        name: networkpolicy.metadata.name,
        policy_types: networkpolicy.spec.policyTypes,
        resource_version: networkpolicy.metadata.resourceVersion,
        created: networkpolicy.metadata.creationTimestamp
      }
      return p
    },
    nameClick: function(namespace, name) {
      this.$router.push({name: 'networkpolicyDetail', params: {namespace: namespace, networkpolicyName: name}})
    },
    getNetworkPolicyYaml: function(namespace, name) {
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
        Message.error("获取NetworkPolicy名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getNetworkPolicy(cluster, namespace, name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteNetworkPolicies: function(networkpolicies) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( networkpolicies.length <= 0 ){
        Message.error("请选择要删除的NetworkPolicies")
        return
      }
      let params = {
        resources: networkpolicies
      }
      deleteNetworkPolicies(cluster, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    updateNetworkPolicy: function() {
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
        Message.error("获取NetworkPolicy参数异常，请刷新重试")
        return
      }
      // console.log(this.yamlValue)
      updateNetworkPolicy(cluster, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delNetworkPoliciesFunc: function() {
      if (this.delNetworkPolicies.length > 0){
        let delNetworkPolicies = []
        for (var p of this.delNetworkPolicies) {
          delNetworkPolicies.push({namespace: p.namespace, name: p.name})
        }
        this.deleteNetworkPolicies(delNetworkPolicies)
      }
    },
    handleSelectionChange(val) {
      this.delNetworkPolicies = val;
      if (val.length > 0){
        this.delFunc = this._delNetworkPoliciesFunc
      } else {
        this.delFunc = undefined
      }
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
