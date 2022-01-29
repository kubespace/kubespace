<template>
  <div>
    <clusterbar :titleName="titleName" :editFunc="getHpaYaml" :delFunc="deleteHpa" />
    <div class="dashboard-container">
      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" class="pod-item" label-width="180px" v-if="hpa.metadata" style="margin: 15px 10px 20px 10px;">
          <el-form-item label="名称">
            <span>{{ hpa.metadata.name }}</span>
          </el-form-item>
          <el-form-item label="创建时间">
            <span>{{ hpa.metadata.creationTimestamp }}</span>
          </el-form-item>
          <el-form-item label="Namespace">
            <span>{{ hpa.metadata.namespace }}</span>
          </el-form-item>
          <el-form-item label="Min Pods">
            <span v-if="hpa.spec">{{ hpa.spec.minReplicas }}</span>
          </el-form-item>
          <el-form-item label="Max Pods">
            <span v-if="hpa.spec">{{ hpa.spec.maxReplicas }}</span>
          </el-form-item>
          <el-form-item label="Replicas">
            <span v-if="hpa.status">{{ hpa.status.currentReplicas }}</span>
          </el-form-item>
          <el-form-item label="Reference">
            <span v-if="hpa.spec">{{ hpa.spec.scaleTargetRef.kind }}/{{hpa.spec.scaleTargetRef.name}}</span>
          </el-form-item>
        </el-form>
      </div>


      <div style="padding: 10px 8px 0px;">
        <div>事件</div>
        <div class="msgClass" style="margin-top: 15px;">
          <el-table
            v-if="hpaEvents && hpaEvents.length > 0"
            :data="hpaEvents"
            class="table-fix"
            tooltip-effect="dark"
            style="width: 100%"
            v-loading="eventLoading"
            :cell-style="cellStyle"
            :default-sort = "{prop: 'event_time', order: 'descending'}"
            >
            <el-table-column
              prop="type"
              label="类型"
              min-width="25"
              show-overflow-tooltip>
            </el-table-column>
            <el-table-column
              prop="object"
              label="对象"
              min-width="55"
              show-overflow-tooltip>
              <template slot-scope="scope">
                <span>
                  {{ scope.row.object.kind }}/{{ scope.row.object.name }}
                </span>
              </template>
            </el-table-column>
            <el-table-column
              prop="reason"
              label="原因"
              min-width="50"
              show-overflow-tooltip>
              <template slot-scope="scope">
                <span>
                  {{ scope.row.reason ? scope.row.reason : "——" }}
                </span>
              </template>
            </el-table-column>
            <el-table-column
              prop="message"
              label="信息"
              min-width="120"
              show-overflow-tooltip>
              <template slot-scope="scope">
                <span>
                  {{ scope.row.message ? scope.row.message : "——" }}
                </span>
              </template>
            </el-table-column>
            <el-table-column
              prop="event_time"
              label="触发时间"
              min-width="50"
              show-overflow-tooltip>
            </el-table-column>
          </el-table>
          <div v-else style="color: #909399; text-align: center">暂无数据</div>
        </div>
      </div>

      <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
        <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
        <span slot="footer" class="dialog-footer">
          <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
          <el-button plain  @click="updateHpa()" size="small">确 定</el-button>
        </span>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { getHpa, updateHpa, deleteHpa } from '@/api/hpa'
import { listEvents } from '@/api/event'
import { Message } from 'element-ui'

export default {
  name: 'HpaDetail',
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
      originHpa: {},
      eventLoading: true,
      activeNames: ["1"],
      hpaEvents: []
    }
  },
  created() {
    this.fetchData()
  },
  watch: {},
  computed: {
    titleName: function() {
      return ['HpaDetail', this.hpaName]
    },
    hpaName: function() {
      return this.$route.params ? this.$route.params.hpaName : ''
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    hpa: function() {
      console.log(this.originHpa)
      return this.originHpa
    },
    namespace: function() {
      return this.$route.params ? this.$route.params.namespace : ""
    }
  },
  methods: {
    handleChange(val) {
        console.log(val);
    },
    fetchData: function() {
      this.originHpa = {}
      this.loading = true
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error('获取集群参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.hpaName) {
        Message.error('获取Hpa名称参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.namespace) {
        Message.error('获取获取Hpa命名空间参数异常，请刷新重试')
      }
      getHpa(cluster, this.namespace, this.hpaName).then(response => {
        this.loading = false
        this.originHpa = response.data
        console.log("*******", this.originHpa)
        listEvents(cluster, this.originHpa.metadata.uid).then(response => {
          this.eventLoading = false
          if (response.data) {
            this.hpaEvents = response.data.length > 0 ? response.data : []
          }
        }).catch(() => {
          this.eventLoading = false
        })
      }).catch(() => {
        this.loading = false
      })
    },
    getHpaYaml: function() {
      if (!this.hpaName) {
        Message.error('获取Hpa参数异常，请刷新重试')
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
      getHpa(cluster, this.namespace, this.hpaName, 'yaml')
        .then((response) => {
          this.yamlLoading = false
          this.yamlValue = response.data
        })
        .catch(() => {
          this.yamlLoading = false
        })
    },
    updateHpa: function() {
      if (!this.hpa) {
        Message.error("获取Hpa参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      console.log(this.hpa)
      updateHpa(cluster, this.namespace, this.hpaName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    deleteHpa: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( !this.hpa ) {
        Message.error("获取Hpa参数异常，请刷新重试")
      }
      let hpas = [{
        namespace: this.namespace,
        name: this.hpaName,
      }]
      let params = {
        resources: hpas
      }
      deleteHpa(cluster, params).then(() => {
        Message.success("删除成功")
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

.msgClass {
  margin: 8px 10px 15px 10px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
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
/* .msgClass {
  margin: 0px 25px;
} */
.msgClass .el-table::before {
  height: 0px;
}
</style>
