<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :delFunc="delFunc" :createFunc="openCreateFormDialog"/>
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
                <!-- <el-dropdown-item @click.native.prevent="nameClick(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item> -->
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="getNamespaceYaml(scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="deleteNamespaces([{name: scope.row.name}])">
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

    <el-dialog title="创建命名空间" :visible.sync="createFormVisible"
      :destroy-on-close="true" width="40%" :close-on-click-modal="false">
      <div v-loading="dialogLoading">
        <div class="dialogContent" style="margin: 0px;">
          <el-form :model="form.metadata" :rules="rules" ref="form" label-position="left" label-width="105px">
            <el-form-item label="名称" prop="name" autofocus required>
              <el-input v-model="form.metadata.name" style="width: 100%;" autocomplete="off" 
                placeholder="只能包含小写字母数字以及-和.,数字或者字母开头或结尾" size="small"></el-input>
            </el-form-item>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter" style="margin-top: 25px;">
          <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="handleCreateNamespace()" >确 定</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { ResType, listResource, getResource, delResource, updateResource } from '@/api/cluster/resource'
import { createYaml } from '@/api/cluster'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'
import yaml from 'js-yaml'
import { createResource } from '../../api/cluster/resource'

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
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      originNamespaces: [],
      search_ns: [],
      search_name: '',
      delFunc: undefined,
      delNamespaces: [],
      createFormVisible: false,
      form: {
        apiVersion: "v1",
        kind: "Namespace",
        metadata: {name: ""}
      },
      rules: {},
      dialogLoading: false
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
    cluster: function() {
      return this.$store.state.cluster
    },
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originNamespaces = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listResource(cluster, ResType.Namespace).then(response => {
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
      getResource(cluster, ResType.Namespace, "", name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteNamespaces(namespaces) {
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
      let cs = ''
      for(let c of namespaces) {
        cs += `${c.name},`
      }
      cs = cs.substr(0, cs.length - 1)
      this.$confirm(`请确认是否删除「${cs}」命名空间?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        delResource(this.cluster, ResType.Namespace, params).then(() => {
          Message.success("删除命名空间成功")
          this.loading = false
          this.fetchData()
        }).catch((err) => {
          this.loading = false
        });
      }).catch(() => {       
      });
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
      updateResource(cluster, ResType.Namespace, "", this.yamlName, this.yamlValue).then(() => {
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
    handleCreateNamespace() {
      if(!this.form.metadata.name) {
        Message.error("请输入命名空间名称")
        return
      }
      let yamlStr = yaml.dump(this.form)
      this.dialogLoading = true
      createResource(this.cluster, yamlStr).then((response) => {
        this.dialogLoading = false
        this.createFormVisible = false
        Message.success("创建命名空间成功")
        this.fetchData()
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    openCreateFormDialog() {
      this.form = {
        apiVersion: "v1",
        kind: "Namespace",
        metadata: {
          name: "",
        }
      }
      this.createFormVisible = true
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
