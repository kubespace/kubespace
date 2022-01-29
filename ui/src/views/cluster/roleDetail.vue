<template>
  <div>
    <clusterbar :titleName="titleName" :delFunc="deleteRoles" :editFunc="getRoleYaml"/>
    <div class="dashboard-container" v-loading="loading">

      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" class="pod-item" label-width="120px" style="margin: 15px 10px 20px 10px;">
          <el-form-item label="名称">
            <span>{{ role.name }}</span>
          </el-form-item>
          <el-form-item label="创建时间">
            <span>{{ role.created }}</span>
          </el-form-item>
          <el-form-item label="命名空间">
            <span>{{ role.namespace }}</span>
          </el-form-item>
          <!-- <el-form-item label="规则">
            <span>{{ role.rules }}</span>
          </el-form-item> -->
          <!-- <el-form-item label="Secrets">
            <span>{{ getSecretsName(role.secrets) }}</span>
          </el-form-item> -->
          <el-form-item label="标签">
            <span v-if="!role.labels">—</span>
            <template v-else v-for="(val, key) in role.labels" >
              <span :key="key" class="back-class">{{key}}: {{val}} <br/></span>
            </template>
          </el-form-item>
          <!-- <el-form-item label="注解">
            <span v-if="!role.annotations">—</span>
            
            <template v-else v-for="(val, key) in role.annotations">
              <span :key="key">{{key}}: {{val}}<br/></span>
            </template>
          </el-form-item> -->
        </el-form>
      </div>

      <div style="padding: 0px 8px;">
        <div>Rules</div>
        <div class="msgClass" style="margin: 15px 10px 30px 10px;">
          <el-table
            v-if="role.rules && role.rules.length > 0"
            :data="role.rules"
            class="table-fix"
            tooltip-effect="dark"
            style="width: 100%"
            :cell-style="cellStyle"
            :default-sort = "{prop: 'event_time', order: 'descending'}"
            >
            <el-table-column
              prop="resources"
              label="资源"
              min-width="45"
              show-overflow-tooltip>
              <template slot-scope="scope">
                <span>
                  {{ scope.row.resources.join(',') }}
                </span>
              </template>
            </el-table-column>
            <el-table-column
              prop="verbs"
              label="权限"
              min-width="35"
              show-overflow-tooltip>
              <template slot-scope="scope">
                <span>
                  {{ scope.row.verbs.join(',') }}
                </span>
              </template>
            </el-table-column>
            <el-table-column
              prop="apiGroups"
              label="apiGroups"
              min-width="20"
              show-overflow-tooltip>
              <!-- <template slot-scope="scope">
                <span>
                  {{ scope.row.reason ? scope.row.reason : "—" }}
                </span>
              </template> -->
            </el-table-column>
            <el-table-column
              prop="resourceNames"
              label="资源名称"
              min-width="40"
              show-overflow-tooltip>
              <!-- <template slot-scope="scope">
                <span>
                  {{ scope.row.message ? scope.row.message : "—" }}
                </span>
              </template> -->
            </el-table-column>
          </el-table>
          <div v-else style="color: #909399; text-align: center">暂无数据</div>
        </div>
      </div>

      <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
        <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
        <span slot="footer" class="dialog-footer">
          <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
          <el-button plain @click="updateRole()" size="small">确 定</el-button>
        </span>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { getRole, deleteRoles, updateRole } from '@/api/role'
import { Message } from 'element-ui'

export default {
  name: 'RoleDetail',
  components: {
    Clusterbar,
    Yaml
  },
  data() {
    return {
      yamlDialog: false,
      yamlValue: "",
      yamlLoading: true,
      cellStyle: {border: 0},
      loading: true,
      originRole: undefined,
    }
  },
  created() {
    this.fetchData()
  },
  watch: {
    roleWatch: function (newObj) {
      if (newObj && this.originRole) {
        let newUid = newObj.resource.metadata.uid
        if (newUid !== this.role.uid) {
          return
        }
        let newRv = newObj.resource.metadata.resourceVersion
        if (this.role.resource_version < newRv) {
          this.originRole = newObj.resource
        }
      }
    },
  },
  computed: {
    titleName: function() {
      return ['Roles', this.roleName]
    },
    roleName: function() {
      return this.$route.params ? this.$route.params.roleName : ''
    },
    namespace: function() {
      return this.$route.params ? this.$route.params.namespace : ''
    },
    kind: function() {
      if (this.namespace) return 'Role';
      return 'ClusterRole'
    },
    role: function() {
      let p = this.buildRole(this.originRole)
      return p
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    roleWatch: function() {
      return this.$store.getters["ws/rolesWatch"]
    },
  },
  methods: {
    fetchData: function() {
      this.originRole = null
      this.loading = true
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        this.loading = false
        return
      }
      if (this.kind === 'Role' && !this.namespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        this.loading = false
        return
      }
      var namespace = this.namespace;
      if (this.kind === 'ClusterRole') namespace = 'n';
      if (!this.roleName) {
        Message.error("获取Role名称参数异常，请刷新重试")
        this.loading = false
        return
      }
      getRole(cluster, namespace, this.roleName, this.kind).then(response => {
        this.loading = false
        this.originRole = response.data

      }).catch(() => {
        this.loading = false
      })
    },
    buildRole: function(role) {
      if (!role) return {}
      let p = {
        uid: role.metadata.uid,
        namespace: role.metadata.namespace,
        name: role.metadata.name,
        resource_version: role.metadata.resourceVersion,
        rules: role.rules,
        created: role.metadata.creationTimestamp,
        labels: role.metadata.labels,
        annotations: role.metadata.annotations,
      }
      return p
    },
    toogleExpand: function(row) {
      let $table = this.$refs.table;
      $table.toggleRowExpansion(row)
    },
    deleteRoles: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( !this.role ) {
        Message.error("获取Role参数异常，请刷新重试")
      }
      let roles = [{
        namespace: this.role.namespace,
        name: this.role.name,
      }]
      let params = {
        resources: roles
      }
      deleteRoles(cluster, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    getRoleYaml: function() {
      if (!this.role) {
        Message.error("获取Role参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      var namespace = this.namespace
      if (this.kind === 'ClusterRole') namespace = 'n'
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getRole(cluster, namespace, this.role.name, this.kind, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    updateRole: function() {
      if (!this.role) {
        Message.error("获取Role参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      updateRole(cluster, this.role.namespace, this.role.name, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
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
/* 
.item-class {
  padding: 20px 20px 20px 5px;
  font-size: 0;
}

.item-class  */

.pod-item {
  margin: 20px 5px 30px 5px;
  padding: 10px 20px;
  font-size: 14px;
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
  width: 100%;
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
