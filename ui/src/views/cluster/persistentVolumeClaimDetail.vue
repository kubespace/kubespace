<template>
  <div>
    <clusterbar :titleName="titleName" :editFunc="getPersistentVolumeClaimYaml" />
    <div class="dashboard-container" v-loading="loading">
      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" class="pod-item" label-width="120px" v-if="PersistentVolumeClaim"
        style="margin: 15px 10px 20px 10px;">
          <el-form-item label="名称">
            <span>{{ PersistentVolumeClaim.metadata ? PersistentVolumeClaim.metadata.name : '' }}</span>
          </el-form-item>
          <el-form-item label="创建时间">
            <span>{{ PersistentVolumeClaim.metadata ? PersistentVolumeClaim.metadata.creationTimestamp : '' }}</span>
          </el-form-item>
          <el-form-item label="命名空间">
            <span>{{ PersistentVolumeClaim.metadata ? PersistentVolumeClaim.metadata.namespace : '' }}</span>
          </el-form-item>
          <el-form-item label="状态">
            <span>{{ PersistentVolumeClaim.metadata ? PersistentVolumeClaim.status.phase : '' }}</span>
          </el-form-item>
          <el-form-item label="容量">
            <span>{{ PersistentVolumeClaim.metadata ? PersistentVolumeClaim.spec.resources.requests.storage : '' }}</span>
          </el-form-item>
          <el-form-item label="存储卷">
            <span class="name-class" v-on:click="nameClick()">
              {{ PersistentVolumeClaim.metadata ? PersistentVolumeClaim.spec.volumeName : '' }}
            </span>
          </el-form-item>
          <el-form-item label="存储类">
            <span v-if="PersistentVolumeClaim.metadata && !PersistentVolumeClaim.spec.storageClassName">—</span>
            <span v-else>{{ PersistentVolumeClaim.metadata ? PersistentVolumeClaim.spec.storageClassName : '' }}</span>
          </el-form-item>
          <el-form-item label="访问模式" v-if="PersistentVolumeClaim.metadata" >
            <template v-for="key in PersistentVolumeClaim.spec.accessModes" >
              <span :key="key" class="back-class">{{key}} <br/></span>
            </template>
          </el-form-item>
          <el-form-item label="存储类型" v-if="PersistentVolumeClaim.metadata" >
            <span>{{ PersistentVolumeClaim.spec.volumeMode }}</span>
          </el-form-item>
        </el-form>
      </div>

      <el-tabs value="resource" style="padding: 0px 8px;" v-if="PersistentVolumeClaim.metadata">
        <el-tab-pane label="资源要求" name="resource">
          <div class="msgClass">
            <div v-if="PersistentVolumeClaim.spec.resources">
              <el-form label-position="left" class="pod-item" style="box-shadow: 0 0 0 0;" label-width="150px">
                <el-form-item label="Requests">
                  <span v-if="!PersistentVolumeClaim.spec.resources.requests">—</span>
                  <template v-else v-for="(val, key) in PersistentVolumeClaim.spec.resources.requests">
                    <span :key="key" class="back-class">{{key}}: {{val}}</span>
                  </template>
                </el-form-item>
                <el-form-item label="Limits">
                  <span v-if="!PersistentVolumeClaim.spec.resources.limits">—</span>
                  <template v-else v-for="(val, key) in PersistentVolumeClaim.spec.resources.limits">
                    <span :key="key" class="back-class">{{key}}: {{val}}</span>
                  </template>
                </el-form-item>
              </el-form>
            </div>
            <div v-else style="padding: 25px 15px; color: #909399; text-align: center">无存储资源要求</div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="选择器" name="selector">
          <div class="msgClass">
            <el-form label-position="left" class="pod-item" style="box-shadow: 0 0 0 0;" label-width="150px">
              <el-form-item label="Match Labels">
                <span v-if="!PersistentVolumeClaim.spec.selector || !PersistentVolumeClaim.spec.selector.matchLabels">—</span>
                <template v-else v-for="(val, key) in PersistentVolumeClaim.spec.selector.matchLabels">
                  <span :key="key" class="back-class">{{key}}: {{val}}</span>
                </template>
              </el-form-item>
              <el-form-item label="Match Expressions">
                <span v-if="!PersistentVolumeClaim.spec.selector || !PersistentVolumeClaim.spec.selector.matchExpressions">—</span>
                <template v-else v-for="(key, val) in PersistentVolumeClaim.spec.selector.matchExpressions">
                  <span :key="key" class="back-class">{{key}}:{{val}}</span>
                </template>
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>
        <el-tab-pane label="事件" name="events">
          <div class="msgClass">
            <el-table
              v-if="persistentVolumeClaimEvents && persistentVolumeClaimEvents.length > 0"
              :data="persistentVolumeClaimEvents"
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
            <div v-else style="padding: 25px 15px; color: #909399; text-align: center">暂无事件发生</div>
          </div>
        </el-tab-pane>
      </el-tabs>

      <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
        <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
        <span slot="footer" class="dialog-footer">
          <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
          <el-button plain @click="updatePersistentVolumeClaim()" size="small">确 定</el-button>
        </span>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { getPersistentVolumeClaim, updatePersistentVolumeClaim } from '@/api/persistent_volume_claim'
import { listEvents } from '@/api/event'
import { Message } from 'element-ui'

export default {
  name: 'PersistentVolumeClaimDetail',
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
      originPersistentVolumeClaim: {},
      selectContainer: '',
      eventLoading: true,
      activeNames: ["1"],
      persistentVolumeClaimEvents: []
    }
  },
  created() {
    this.fetchData()
  },
  watch: {},
  computed: {
    titleName: function() {
      return ['PersistentVolumeClaim', this.PersistentVolumeClaimName]
    },
    PersistentVolumeClaimName: function() {
      return this.$route.params ? this.$route.params.persistentVolumeClaimName : ''
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    PersistentVolumeClaim: function() {
      console.log(this.originPersistentVolumeClaim)
      return this.originPersistentVolumeClaim
    },
    namespace: function() {
      return this.$route.params ? this.$route.params.namespace : ""
    }
  },
  methods: {
    nameClick() {
      if (this.PersistentVolumeClaim.spec && this.PersistentVolumeClaim.spec.volumeName) {
        this.$router.push({name: "pvDetail", params: { persistentVolumeName: this.PersistentVolumeClaim.spec.volumeName },});
      }
    },
    handleChange(val) {
        console.log(val);
    },
    fetchData: function() {
      this.originPersistentVolumeClaim = {}
      this.loading = true
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error('获取集群参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.PersistentVolumeClaimName) {
        Message.error('获取PersistentVolumeClaim名称参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.namespace) {
        Message.error('获取获取PersistentVolumeClaim命名空间参数异常，请刷新重试')
      }
      getPersistentVolumeClaim(cluster, this.namespace, this.PersistentVolumeClaimName).then(response => {
        this.loading = false
        this.originPersistentVolumeClaim = response.data
        listEvents(cluster, this.originPersistentVolumeClaim.metadata.uid).then(response => {
          this.eventLoading = false
          if (response.data) {
            this.persistentVolumeClaimEvents = response.data.length > 0 ? response.data : []
          }
        }).catch(() => {
          this.eventLoading = false
        })
      }).catch(() => {
        this.loading = false
      })
    },
    getPersistentVolumeClaimYaml: function() {
      if (!this.PersistentVolumeClaimName) {
        Message.error('获取PersistentVolumeClaim参数异常，请刷新重试')
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
      getPersistentVolumeClaim(cluster, this.namespace, this.PersistentVolumeClaimName, 'yaml')
        .then((response) => {
          this.yamlLoading = false
          this.yamlValue = response.data
        })
        .catch(() => {
          this.yamlLoading = false
        })
    },
    updatePersistentVolumeClaim: function() {
      if (!this.PersistentVolumeClaim) {
        Message.error("获取PersistentVolumeClaim参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      updatePersistentVolumeClaim(cluster, this.namespace, this.PersistentVolumeClaimName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
  },
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
  margin: 0px 5px 30px 5px;
  padding: 10px 20px;
  font-size: 0;
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
  /* width: 50%; */
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
