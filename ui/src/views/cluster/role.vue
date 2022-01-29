<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="roles"
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
            <span class="name-class" v-on:click="nameClick(scope.row.namespace, scope.row.name)">
              {{ scope.row.name }}
            </span>
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
        <!-- <el-table-column
          prop="subjects"
          label="绑定主体"
          min-width="45"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.subjects && scope.row.subjects.length > 0">
              <span v-for="s of scope.row.subjects" :key="s.name" class="back-class">
                {{ s.kind + '/' + s.name }} 
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
        </el-table-column> -->
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
                <el-dropdown-item @click.native.prevent="nameClick(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$updatePerm()" @click.native.prevent="getRoleYaml(scope.row.namespace, scope.row.name, scope.row.kind)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$deletePerm()" @click.native.prevent="deleteRoles([{namespace: scope.row.namespace, name: scope.row.name}])">
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
        <el-button plain @click="updateRole()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listRoles, getRole, deleteRoles, updateRole } from '@/api/role'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'

export default {
  name: 'Role',
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
        titleName: ["Roles"],
        maxHeight: window.innerHeight - 150,
        loading: true,
        originRoles: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delRoles: [],
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
    rolesWatch: function (newObj) {
      if (newObj) {
        let newUid = newObj.resource.metadata.uid
        let newRv = newObj.resource.metadata.resourceVersion
        if (newObj.event === 'add') {
          this.originRoles.push(this.buildRoles(newObj.resource))
        } else if (newObj.event === 'update') {
          for (let i in this.originRoles) {
            let d = this.originRoles[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newDp = this.buildRoles(newObj.resource)
                this.$set(this.originRoles, i, newDp)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originRoles = this.originRoles.filter(( { uid } ) => uid !== newUid)
        }
      }
    }
  },
  computed: {
    roles: function() {
      var dlist = []
      for (let p of this.originRoles) {
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
    rolesWatch: function() {
      return this.$store.getters["ws/rolesWatch"]
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originRoles = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listRoles(cluster).then(response => {
          this.loading = false
          this.originRoles = response.data
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
    buildRoles: function(role) {
      if (!role) return
      let p = {
        uid: role.metadata.uid,
        namespace: role.metadata.namespace,
        name: role.metadata.name,
        resource_version: role.metadata.resourceVersion,
        created: role.metadata.creationTimestamp
      }
      return p
    },
    nameClick: function(namespace, name) {
      if(namespace) {
        this.$router.push({name: 'roleDetail', params: {namespace: namespace, roleName: name}})
      } else {
        this.$router.push({name: 'clusterroleDetail', params: {roleName: name}})
      }
    },
    getRoleYaml: function(namespace, name, kind) {
      this.yamlNamespace = ""
      this.yamlName = ""
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (kind === 'Role' && !namespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (kind === 'ClusterRole') namespace = 'n'
      
      if (!name) {
        Message.error("获取Role名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getRole(cluster, namespace, name, kind, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlKind = kind
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteRoles: function(roles) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( roles.length <= 0 ){
        Message.error("请选择要删除的Roles")
        return
      }
      let params = {
        resources: roles
      }
      deleteRoles(cluster, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    updateRole: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!this.yamlKind) {
        Message.error("获取Role参数异常，请刷新重试")
        return
      }
      if (this.yamlKind === 'Role' && !this.yamlNamespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (this.yamlKind === 'ClusterRole') this.yamlNamespace = 'n'
      if (!this.yamlName) {
        Message.error("获取Role参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      updateRole(cluster, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delRolesFunc: function() {
      if (this.delRoles.length > 0){
        let delRoles = []
        for (var p of this.delRoles) {
          delRoles.push({namespace: p.namespace, name: p.name})
        }
        this.deleteRoles(delRoles)
      }
    },
    handleSelectionChange(val) {
      this.delRoles = val;
      if (val.length > 0){
        this.delFunc = this._delRolesFunc
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
