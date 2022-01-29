<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="namespaces"
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
          min-width="40"
          show-overflow-tooltip>
          <!-- <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.namespace, scope.row.name)">
              {{ scope.row.name }}
            </span>
          </template> -->
        </el-table-column>
        <el-table-column
          prop="lables"
          label="标签"
          min-width="55"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.labels">
              <span v-for="(val, key) in scope.row.labels" :key="key" class="back-class">
                {{ key + '=' + val }} 
              </span>
            </template>
            <!-- <span v-else>--</span> -->
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          label="状态"
          min-width="30"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="created"
          label="创建时间"
          min-width="25"
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
                <el-dropdown-item v-if="$updatePerm()" @click.native.prevent="getNamespaceYaml(scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$deletePerm()" @click.native.prevent="deleteNamespaces([{name: scope.row.name}])">
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
        <el-button plain @click="updateNamespace()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listNamespace, getNamespace } from '@/api/namespace'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'

export default {
  name: 'Namespace',
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
        titleName: ["Namespaces"],
        maxHeight: window.innerHeight - 150,
        loading: true,
        originNamespaces: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delNamespaces: [],
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
    namespacesWatch: function (newObj) {
      if (newObj) {
        let newUid = newObj.resource.metadata.uid
        let newRv = newObj.resource.metadata.resourceVersion
        if (newObj.event === 'add') {
          this.originNamespaces.push(this.buildNamespaces(newObj.resource))
        } else if (newObj.event === 'update') {
          for (let i in this.originNamespaces) {
            let d = this.originNamespaces[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newDp = this.buildNamespaces(newObj.resource)
                this.$set(this.originNamespaces, i, newDp)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originNamespaces = this.originNamespaces.filter(( { uid } ) => uid !== newUid)
        }
      }
    }
  },
  computed: {
    namespaces: function() {
      let dlist = []
      for (let p of this.originNamespaces) {
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
    namespacesWatch: function() {
      return this.$store.getters["ws/namespacesWatch"]
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originNamespaces = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listNamespace(cluster).then(response => {
          this.loading = false
          this.originNamespaces = response.data
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    buildNamespaces: function(namespace) {
      if (!namespace) return
      let p = {
        uid: namespace.metadata.uid,
        status: namespace.status.phase,
        name: namespace.metadata.name,
        labels: namespace.metadata.labels,
        resource_version: namespace.metadata.resourceVersion,
        created: namespace.metadata.creationTimestamp
      }
      return p
    },
    // nameClick: function(namespace, name) {
    //   this.$router.push({name: 'namespaceDetail', params: {namespace: namespace, namespaceName: name}})
    // },
    getNamespaceYaml: function(name) {
      this.yamlNamespace = ""
      this.yamlName = ""
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!name) {
        Message.error("获取Namespace名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getNamespace(cluster, name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteNamespaces: function(namespaces) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( namespaces.length <= 0 ){
        Message.error("请选择要删除的Namespaces")
        return
      }
      let params = {
        resources: namespaces
      }
      deleteNamespaces(cluster, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    updateNamespace: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!this.yamlName) {
        Message.error("获取Namespace参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      updateNamespace(cluster, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delNamespacesFunc: function() {
      if (this.delNamespaces.length > 0){
        let delNamespaces = []
        for (var p of this.delNamespaces) {
          delNamespaces.push({name: p.name})
        }
        this.deleteNamespaces(delNamespaces)
      }
    },
    handleSelectionChange(val) {
      this.delNamespaces = val;
      if (val.length > 0){
        this.delFunc = this._delNamespacesFunc
      } else {
        this.delFunc = undefined
      }
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
