<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="rolebindings"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
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
          min-width="40"
          show-overflow-tooltip>
          <template slot-scope="scope">
              {{ scope.row.name }}
          </template>
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="25"
          show-overflow-tooltip>
        </el-table-column>
        <!-- <el-table-column
          prop="kind"
          label="Kind"
          min-width="30"
          show-overflow-tooltip>
        </el-table-column> -->
        <el-table-column
          prop="subjects"
          label="绑定主体"
          min-width="45"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.subjects && scope.row.subjects.length > 0">
              <span v-for="(i,j) in scope.row.subjects" :key="scope.row.uid+i.name+j" class="back-class">
                {{ i.kind + '/' + i.name }} 
              </span>
            </template>
          </template>
        </el-table-column>
        <el-table-column
          prop="role"
          label="角色"
          min-width="35"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ scope.row.role.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="created"
          label="创建时间"
          min-width="30"
          show-overflow-tooltip>
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.created) }}
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
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="getRoleBindingYaml(scope.row.namespace, scope.row.name, scope.row.kind)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="deleteRoleBindings([{namespace: scope.row.namespace, name: scope.row.name}])">
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
        <el-button plain @click="updateRoleBinding()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { ResType, listResource, getResource, delResource, updateResource } from '@/api/cluster/resource'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'

export default {
  name: 'RoleBinding',
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
        yamlKind: "",
        yamlLoading: true,
        cellStyle: {border: 0},
        titleName: ["Role Bindings"],
        maxHeight: window.innerHeight - this.$contentHeight,
        loading: true,
        originRoleBindings: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delRoleBindings: [],
      }
  },
  created() {
    this.fetchData()
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
    rolebindings: function() {
      var dlist = []
      for (let p of this.originRoleBindings) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        
        dlist.push(p)
      }
      dlist.sort((a, b) => {
        if (a.kind < b.kind) return 1;
        return a.name > b.name ? 1 : -1
      })
      return dlist
    },
    cluster() {
      return this.$store.state.cluster
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originRoleBindings = []
      let originRoleBindings = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listResource(cluster, ResType.RoleBinding).then(response => {
          originRoleBindings = response.data || []
          listResource(cluster, ResType.ClusterRoleBinding).then(response => {
            this.loading = false
            let clusterRoleBindings = response.data || []
            for(let r of clusterRoleBindings) {
              originRoleBindings.push(r)
            }
            this.originRoleBindings = originRoleBindings
          }).catch(() => {
            this.loading = false
          })
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
    buildRoleBindings: function(rolebinding) {
      if (!rolebinding) return
      let p = {
        uid: rolebinding.metadata.uid,
        namespace: rolebinding.metadata.namespace,
        name: rolebinding.metadata.name,
        subjects: rolebinding.subjects,
        role: rolebinding.roleRef,
        resource_version: rolebinding.metadata.resourceVersion,
        created: rolebinding.metadata.creationTimestamp
      }
      return p
    },
    nameClick: function(namespace, name) {
      if (namespace) {
        this.$router.push({name: 'rolebindingDetail', params: {namespace: namespace, rolebindingName: name}})
      } else {
        this.$router.push({name: 'clusterrolebindingDetail', params: {rolebindingName: name}})
      }
    },
    getRoleBindingYaml: function(namespace, name, kind) {
      this.yamlNamespace = ""
      this.yamlName = ""
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!kind) {
        Message.error("获取RoleBinding参数异常，请刷新重试")
        return
      }
      let resType = ResType.RoleBinding
      if (kind === 'RoleBinding' && !namespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (kind === 'ClusterRoleBinding') {
        namespace = ''
        resType = ResType.ClusterRoleBinding
      }
      if (!name) {
        Message.error("获取RoleBinding名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getResource(cluster, resType, namespace, name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlKind = kind
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteRoleBindings: function(rolebindings) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( rolebindings.length <= 0 ){
        Message.error("请选择要删除的RoleBindings")
        return
      }
      let params = {
        resources: rolebindings
      }
      delResource(cluster, ResType.RoleBinding, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    updateRoleBinding: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!this.yamlKind) {
        Message.error("获取RoleBinding参数异常，请刷新重试")
        return
      }
      let resType = ResType.RoleBinding
      if (this.yamlKind === 'RoleBinding' && !this.yamlNamespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (this.yamlKind === 'ClusterRoleBinding') {
        this.yamlNamespace = ''
        resType = ResType.ClusterRoleBinding
      }
      if (!this.yamlName) {
        Message.error("获取RoleBinding参数异常，请刷新重试")
        return
      }
      this.yamlLoading = true
      updateResource(cluster, resType, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
        this.yamlLoading = false
        this.yamlDialog = false
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delRoleBindingsFunc: function() {
      if (this.delRoleBindings.length > 0){
        let delRoleBindings = []
        for (var p of this.delRoleBindings) {
          delRoleBindings.push({namespace: p.namespace, name: p.name})
        }
        this.deleteRoleBindings(delRoleBindings)
      }
    },
    handleSelectionChange(val) {
      this.delRoleBindings = val;
      if (val.length > 0){
        this.delFunc = this._delRoleBindingsFunc
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
