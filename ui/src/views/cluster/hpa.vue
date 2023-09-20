<template>
  <div>
    <clusterbar
      :titleName="titleName"
      :nsFunc="nsSearch"
      :nameFunc="nameSearch"
    />
    <div class="dashboard-container">
      <el-table
        ref="multipleTable"
        :data="hpas"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort="{ prop: 'name' }"
        row-key="uid"
      >
        <!-- <el-table-column type="selection" width="45"> </el-table-column> -->
        <el-table-column
          prop="name"
          label="名称"
          min-width="45"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.namespace, scope.row.name)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="40"
          show-overflow-tooltip
        >
        </el-table-column>
        <el-table-column
        prop="min_pods"
        label="Min Pods"
        min-width="25"
        show-overflow-tooltip
      >
      </el-table-column>
      <el-table-column
        prop="max_pods"
        label="Max Pods"
        min-width="25"
        show-overflow-tooltip
      >
      </el-table-column>
      <el-table-column
        prop="replicas"
        label="Replicas"
        min-width="25"
        show-overflow-tooltip
      >
      </el-table-column>
      <el-table-column
      prop="target_cpu_per"
      label="Target Cpu Per"
      min-width="35"
      show-overflow-tooltip
    >
    </el-table-column>
        <el-table-column
          prop="create_time"
          label="创建时间"
          min-width="45"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.create_time) }}
          </template>
        </el-table-column>
        <el-table-column label="" show-overflow-tooltip width="45">
          <template slot-scope="scope">
            <el-dropdown size="medium">
              <el-link :underline="false"
                ><svg-icon
                  style="width: 1.3em; height: 1.3em;"
                  icon-class="operate"
              /></el-link>
              <el-dropdown-menu slot="dropdown">
                <!-- <el-dropdown-item
                  @click.native.prevent="
                    nameClick(scope.row.namespace, scope.row.name)
                  "
                >
                  <svg-icon
                    style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em"
                    icon-class="detail"
                  />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item> -->
                <el-dropdown-item v-if="$editorRole()"
                  @click.native.prevent="
                    getYaml(scope.row.namespace, scope.row.name)
                  "
                >
                  <svg-icon
                    style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em"
                    icon-class="edit"
                  />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()"
                  @click.native.prevent="
                    deleteObjs([
                      { namespace: scope.row.namespace, name: scope.row.name },
                    ])
                  "
                >
                  <svg-icon
                    style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em"
                    icon-class="delete"
                  />
                  <span style="margin-left: 5px;">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog
      title="编辑"
      :visible.sync="yamlDialog"
      :close-on-click-modal="false"
      width="60%"
      top="55px"
    >
      <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="yamlDialog = false" size="small"
          >取 消</el-button
        >
        <el-button plain @click="updatePod()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { Message } from 'element-ui'
import { ResType, listResource, getResource, updateResource, delResource } from '@/api/cluster/resource'

export default {
  name: 'HorizontalPodAutoscalers',
  components: {
    Clusterbar,
    Yaml,
  },
  data() {
    return {
      titleName: ["HorizontalPodAutoscalers"],
      originHpas: [],
      search_name: '',
      search_ns: [],
      cellStyle: { border: 0 },
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      yamlDialog: false,
      yamlNamespace: '',
      yamlName: '',
      yamlValue: '',
      yamlLoading: true,
      updateFormVisible: false,
      createFormVisible: false,
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
    hpas: function() {
      let dlist = []
      for (let p of this.originHpas) {
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
    nameClick: function(namespace, name) {
      this.$router.push({
        name: 'hpaDetail',
        params: { namespace: namespace, hpaName: name },
      })
    },
    nsSearch: function(vals) {
      this.search_ns = []
      for (let ns of vals) {
        this.search_ns.push(ns)
      }
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    fetchData: function() {
      this.loading = true
      this.originHpas = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listResource(cluster, ResType.Hpa).then(response => {
          this.loading = false
          this.originHpas = response.data || []
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
    getYaml: function(namespace, name) {
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
      getResource(cluster, ResType.Hpa, namespace, name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteObjs: function(objs) {
      let cs = ''
      for(let c of objs) {
        cs += `${c.namespace}/${c.name}, `
      }
      cs = cs.substr(0, cs.length - 2)
      this.$confirm(`请确认是否删除「${cs}」Hpa?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        delResource(this.cluster, ResType.Hpa, {resources: objs}).then(() => {
          Message.success("删除成功")
          this.fetchData()
        }).catch((err) => {
          console.log(err)
        });
      }).catch(() => {       
      });
    },
    update: function() {
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
        Message.error("参数异常，请刷新重试")
        return
      }
      this.yamlLoading = true
      updateResource(cluster, ResType.Hpa, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
        this.yamlLoading = false
        this.yamlDialog = false
      }).catch(() => {
        // console.log(e) 
      })
    },
    
    _delDeploymentsFunc: function() {
      if (this.delDeployments.length > 0){
        let delDeployments = []
        for (var p of this.delDeployments) {
          delDeployments.push({namespace: p.namespace, name: p.name})
        }
        this.deleteDeployments(delDeployments)
      }
    },
    handleSelectionChange(val) {
      this.delDeployments = val;
      if (val.length > 0){
        this.delFunc = this._delDeploymentsFunc
      } else {
        this.delFunc = undefined
      }
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
  </style>