<template>
  <div>
    <clusterbar :titleName="titleName" :editFunc="getConfigMapYaml" />
    <div class="dashboard-container">
      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" class="pod-item" v-if="configMap.metadata" style="margin: 15px 10px 20px 10px;">
          <el-form-item label="名称">
            <span>{{ configMap.metadata.name }}</span>
          </el-form-item>
          <el-form-item label="创建时间">
            <span>{{ configMap.metadata.creationTimestamp }}</span>
          </el-form-item>
          <el-form-item label="命名空间">
            <span>{{ configMap.metadata.namespace }}</span>
          </el-form-item>

          <el-form-item label="标签">
            <span v-if="!configMap.metadata.labels">—</span>
            <div v-else v-for="(val, key) in configMap.metadata.labels" :key="key" >
              <span :key="key" class="back-class">{{key}}:{{val}}</span> <br/>
            </div>
          </el-form-item>
        </el-form>
      </div>

      <div style="padding: 10px 8px 0px;">
        <div>数据</div>
        <div class="msgClass" style="margin-top: 15px;">
          <el-table
            v-if="configData"
            :data="configData"
            class="table-fix"
            tooltip-effect="dark"
            style="width: 100%"
            :cell-style="cellStyle"
            :default-sort = "{prop: 'lastProbeTime'}"
            >
            <el-table-column
              prop="key"
              label="键"
              min-width="30"
              show-overflow-tooltip>
              <template slot-scope={row}>
                <el-input placeholder="请输入内容" v-model="row.key"></el-input>
              </template>
            </el-table-column>
            <el-table-column
              prop="value"
              label="值"
              min-width="80"
              show-overflow-tooltip>
              <template slot-scope={row}>
                <el-input placeholder="请输入内容" v-model="row.value"></el-input>
              </template>
            </el-table-column>
            <el-table-column label="操作" min-width="20" fixed="right" align="center">
              <template slot-scope="scope">
                <el-button
                  size="mini"
                  type="danger"
                  @click="handleDelete(scope.$index, scope.row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div v-else style="padding: 25px 15px ; color: #909399; text-align: center">暂无数据</div>
        </div>
      </div>

      <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
        <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
        <span slot="footer" class="dialog-footer">
          <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
          <el-button plain @click="updateConfigMap()" size="small">确 定</el-button>
        </span>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { getConfigMap, updateConfigMap } from '@/api/config_map'
import { listEvents } from '@/api/event'
import { Message } from 'element-ui'

export default {
  name: 'ConfigMapDetail',
  components: {
    Clusterbar,
    Yaml,
  },
  data() {
    return {
      yamlDialog: false,
      yamlValue: '',
      yamlLoading: true,
      cellStyle: { border: 0 },
      loading: true,
      originConfigMap: {},
      selectContainer: '',
      eventLoading: true,
      activeNames: ["1"],
      labels: [],
      configMapEvents: []
    }
  },
  created() {
    this.fetchData()
  },
  watch: {},
  computed: {
    titleName: function() {
      return ['ConfigMap', this.configMapName]
    },
    configMapName: function() {
      return this.$route.params ? this.$route.params.configMapName : ''
    },
    namespace: function() {
      return this.$route.params ? this.$route.params.namespace : ''
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    configMap: function() {
      console.log("####", this.originConfigMap)
      return this.originConfigMap
    },
    configData: function() {
      if (!this.originConfigMap.data) return []
      let d = this.originConfigMap.data
      let dataTable = []
      Object.keys(d).forEach(key => {
      dataTable.push({
          key: key,
          value: d[key]
        })
      })
      return dataTable
    },
    // labels: function() {
    //   if (!this.originConfigMap.metadata.labels) return []
    //   let la = []
    //   Object.keys(this.originConfigMap.metadata.labels).forEach(key => {
    //     la.push({
    //       key: key,
    //       value: this.originConfigMap.metadata.labels[key]
    //     })
    //   })
    //   return la
    // }
  },
  methods: {
    handleChange(val) {
        console.log(val);
    },
    addConfigMap: function() {
        this.configData.push({
          key: "",
          value: ""
        })
    },
    fetchData: function() {
      this.originConfigMap = {}
      this.loading = true
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error('获取集群参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.namespace) {
        Message.error('获取命名空间参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.configMapName) {
        Message.error('获取ConfigMap名称参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      console.log("******************************")
      getConfigMap(cluster, this.namespace, this.configMapName).then(response => {
        this.loading = false
        this.originConfigMap = response.data
        if (this.originConfigMap.metadata) {
        Object.keys(this.originConfigMap.metadata.labels).forEach(key => {
            this.labels.push({
              key: key,
              value: this.originConfigMap.metadata.labels[key]
            })
          })
        }
        listEvents(cluster, this.originConfigMap.metadata.uid).then(response => {
          this.eventLoading = false
          if (response.data) {
            this.configMapEvents = response.data.length > 0 ? response.data : []
          }
        }).catch(() => {
          this.eventLoading = false
        })
      }).catch(() => {
        this.loading = false
      })
    },
    getConfigMapYaml: function() {
      if (!this.configMap) {
        Message.error('获取Pod参数异常，请刷新重试')
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error('获取集群参数异常，请刷新重试')
        return
      }
      this.yamlValue = ''
      this.yamlDialog = true
      this.yamlLoading = true
      getConfigMap(cluster, this.configMap.metadata.namespace, this.configMap.metadata.name, 'yaml')
        .then((response) => {
          this.yamlLoading = false
          this.yamlValue = response.data
        })
        .catch(() => {
          this.yamlLoading = false
        })
    },
    updateConfigMap: function() {
      if (!this.configMap) {
        Message.error("获取ConfigMap参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      updateConfigMap(cluster, this.configMap.metadata.namespace, this.configMap.metadata.name, this.yamlValue).then(() => {
        this.yamlDialog = false
        Message.success("更新成功")
        this.fetchData()
      }).catch(() => {
        // console.log(e) 
      })
    },
  },
}
</script>

<style lang="scss" scoped>
  .my-table >>> .el-table__row>td{
  /* 去除表格线 */
  border: none;
}

.my-table >>> .el-table::before {
	height: 0px;
}

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
  color: #409eff;
}
</style>

<style>
/* .el-table__expand-icon {
  display: none;
} */
.el-table__expanded-cell[class*='cell'] {
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

.pod-item {
  margin: 20px 5px 30px 5px;
  padding: 10px 20px;
  font-size: 0;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}
.pod-item label {
  width: 90px;
  color: #99a9bf;
  font-weight: 400;
}
.pod-item .el-form-item {
  margin-right: 0;
  margin-bottom: 0;
  width: 60%;
}
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
.msgClass {
  margin: 0px 25px;
}
.msgClass .el-table::before {
  height: 0px;
}
.msgClass {
  margin: 8px 10px 15px 10px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}
</style>
