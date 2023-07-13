<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="serviceaccounts"
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
          min-width="25"
          show-overflow-tooltip>
          <template slot-scope="scope">
              {{ scope.row.name }}
          </template>
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="20"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="type"
          label="Secrets"
          min-width="35"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.secrets && scope.row.secrets.length > 0">
              <span>{{ getSecretsName(scope.row.secrets) }}</span>
            </template>
          </template>
        </el-table-column>
        <el-table-column
          prop="created"
          label="创建时间"
          min-width="15"
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
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="getServiceAccountYaml(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="deleteServiceAccounts([{namespace: scope.row.namespace, name: scope.row.name}])">
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
        <el-button plain @click="updateServiceAccount()" size="small">确 定</el-button>
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
  name: 'ServiceAccount',
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
        titleName: ["ServiceAccounts"],
        maxHeight: window.innerHeight - this.$contentHeight,
        loading: true,
        originServiceAccounts: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delServiceAccounts: [],
      }
  },
  created() {
    this.fetchData()
  },
  watch: {
    cluster: function() {
      this.fetchData()
    }
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
  computed: {
    serviceaccounts: function() {
      let dlist = []
      for (let p of this.originServiceAccounts) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        
        dlist.push(p)
      }
      return dlist
    },
    cluster() {
      return this.$store.state.cluster
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originServiceAccounts = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listResource(cluster, ResType.ServiceAccount).then(response => {
          this.loading = false
          this.originServiceAccounts = response.data
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
    buildServiceAccounts: function(serviceaccount) {
      if (!serviceaccount) return
      let p = {
        uid: serviceaccount.metadata.uid,
        namespace: serviceaccount.metadata.namespace,
        name: serviceaccount.metadata.name,
        secrets: serviceaccount.secrets,
        resource_version: serviceaccount.metadata.resourceVersion,
        created: serviceaccount.metadata.creationTimestamp
      }
      return p
    },
    nameClick: function(namespace, name) {
      this.$router.push({name: 'serviceaccountDetail', params: {namespace: namespace, serviceaccountName: name}})
    },
    getServiceAccountYaml: function(namespace, name) {
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
        Message.error("获取ServiceAccount名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getResource(cluster, ResType.ServiceAccount, namespace, name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteServiceAccounts: function(serviceaccounts) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( serviceaccounts.length <= 0 ){
        Message.error("请选择要删除的ServiceAccounts")
        return
      }
      let params = {
        resources: serviceaccounts
      }
      delResource(cluster, ResType.ServiceAccount, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    updateServiceAccount: function() {
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
        Message.error("获取ServiceAccount参数异常，请刷新重试")
        return
      }
      updateResource(cluster, ResType.ServiceAccount, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delServiceAccountsFunc: function() {
      if (this.delServiceAccounts.length > 0){
        let delServiceAccounts = []
        for (var p of this.delServiceAccounts) {
          delServiceAccounts.push({namespace: p.namespace, name: p.name})
        }
        this.deleteServiceAccounts(delServiceAccounts)
      }
    },
    handleSelectionChange(val) {
      this.delServiceAccounts = val;
      if (val.length > 0){
        this.delFunc = this._delServiceAccountsFunc
      } else {
        this.delFunc = undefined
      }
    },
    getSecretsName(secrets) {
      if (!secrets) return ''
      var n = []
      for(let s of secrets) n.push(s.name)
      return n.join(",")
    }
  }
}
</script>
