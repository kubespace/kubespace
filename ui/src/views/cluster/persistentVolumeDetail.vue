<template>
  <div>
    <clusterbar :titleName="titleName" :editFunc="getPersistentVolumeYaml" />
    <div class="dashboard-container">
      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" class="pod-item" label-width="120px" v-if="persistentVolume.metadata"
        style="margin: 15px 10px 20px 10px;">
          <el-form-item label="名称">
            <span>{{ persistentVolume.metadata.name }}</span>
          </el-form-item>
          <el-form-item label="创建时间">
            <span>{{ persistentVolume.metadata.creationTimestamp }}</span>
          </el-form-item>
          <el-form-item label="状态">
            <span>{{ persistentVolume.status.phase }}</span>
          </el-form-item>
          <el-form-item label="容量">
            <span>{{ persistentVolume.spec.capacity.storage }}</span>
          </el-form-item>
          <el-form-item label="访问模式">
            <template v-for="key in persistentVolume.spec.accessModes" >
              <span :key="key" class="back-class">{{key}} <br/></span>
            </template>
          </el-form-item>
          <el-form-item label="存储声明">
            <span v-if="persistentVolume.spec.claimRef">
              {{ persistentVolume.spec.claimRef.namespace + '/' + persistentVolume.spec.claimRef.name }}
            </span>
          </el-form-item>
          <el-form-item label="存储类">
            <span>{{ persistentVolume.spec.storageClassName }}</span>
          </el-form-item>
          <el-form-item label="存储类型">
            <span>{{ persistentVolume.spec.volumeMode }}</span>
          </el-form-item>
          <el-form-item label="重声明策略">
            <span>{{ persistentVolume.spec.persistentVolumeReclaimPolicy }}</span>
          </el-form-item>
          <el-form-item label="标签">
            <span v-if="!persistentVolume.metadata.labels">——</span>
            <template v-else v-for="(val, key) in persistentVolume.metadata.labels" >
              <span :key="key" class="back-class">{{key}}: {{val}} <br/></span>
            </template>
          </el-form-item>
        </el-form>
      </div>

      <el-tabs value="back" style="padding: 0px 8px;">
        <el-tab-pane label="后端存储" name="back">
          <div v-for="(val, key) in persistentVolume.spec" :key="key">
            <template v-if="pvSpec.indexOf(key) < 0">
              <div style="margin: 0px 0px 5px 20px;">{{ key }}</div>
              <div class="msgClass">
                <el-table
                  :data="dictToList(val)"
                  class="table-fix"
                  tooltip-effect="dark"
                  style="width: 100%"
                  v-loading="eventLoading"
                  :cell-style="cellStyle"
                  :default-sort = "{prop: 'event_time', order: 'descending'}"
                  >
                  <el-table-column
                    prop="key"
                    label="键"
                    min-width="20"
                    show-overflow-tooltip>
                  </el-table-column>
                  <el-table-column
                    prop="val"
                    label="值"
                    show-overflow-tooltip>
                    <template slot-scope="scope">
                      <span v-if="scope.row.isDict">
                        <div v-for="(val, key) in scope.row.val" :key="key">
                          {{ key }}: {{ val }}
                        </div>
                      </span>
                      <span v-else>{{ scope.row.val }}</span>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </template>
          </div>
        </el-tab-pane>
        <el-tab-pane label="事件" name="events">
          <div class="msgClass">
            <el-table
              v-if="persistentVolumeEvents && persistentVolumeEvents.length > 0"
              :data="persistentVolumeEvents"
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
            <div v-else style="padding: 25px 15px;color: #909399; text-align: center">暂无事件发生</div>
          </div>
        </el-tab-pane>
      </el-tabs>

      <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
        <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
        <span slot="footer" class="dialog-footer">
          <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
          <el-button plain @click="updatePersistentVolume" size="small">确 定</el-button>
        </span>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { getPersistentVolume, updatePersistentVolume } from '@/api/persistent_volume'
import { listEvents } from '@/api/event'
import { Message } from 'element-ui'

export default {
  name: 'PersistentVolumeDetail',
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
      originPersistentVolume: {},
      selectContainer: '',
      eventLoading: true,
      activeNames: ["1"],
      persistentVolumeEvents: [],
      pvSpec: ['capacity', 'accessModes', 'claimRef', 'persistentVolumeReclaimPolicy', 
               'storageClassName', 'mountOptions', 'volumeMode', 'nodeAffinity']
    }
  },
  created() {
    this.fetchData()
  },
  watch: {},
  computed: {
    titleName: function() {
      return ['PersistentVolume', this.persistentVolumeName]
    },
    persistentVolumeName: function() {
      return this.$route.params ? this.$route.params.persistentVolumeName : ''
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    persistentVolume: function() {
      console.log(this.originPersistentVolume)
      return this.originPersistentVolume
    },
  },
  methods: {
    handleChange(val) {
        console.log(val);
    },
    fetchData: function() {
      this.originPersistentVolume = {}
      this.loading = true
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error('获取集群参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.persistentVolumeName) {
        Message.error('获取PersistentVolume名称参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      getPersistentVolume(cluster, this.persistentVolumeName).then(response => {
        this.loading = false
        this.originPersistentVolume = response.data
        listEvents(cluster, this.originPersistentVolume.metadata.uid).then(response => {
          this.eventLoading = false
          if (response.data) {
            this.persistentVolumeEvents = response.data.length > 0 ? response.data : []
          }
        }).catch(() => {
          this.eventLoading = false
        })
      }).catch(() => {
        this.loading = false
      })
    },
    getPersistentVolumeYaml: function() {
      if (!this.persistentVolumeName) {
        Message.error('获取PersistentVolume参数异常，请刷新重试')
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
      getPersistentVolume(cluster, this.persistentVolume.metadata.name, 'yaml')
        .then((response) => {
          this.yamlLoading = false
          this.yamlValue = response.data
        })
        .catch(() => {
          this.yamlLoading = false
        })
    },
    updatePersistentVolume: function() {
      if (!this.persistentVolume) {
        Message.error("获取PersistentVolume参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      updatePersistentVolume(cluster, this.persistentVolumeName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    dictToList: function(obj) {
      var l = []
      for(var k in obj) {
        var o = {
          key: k,
          val: obj[k],
          isDict: false,
        }
        console.log(Object.prototype.toString.call(obj[k]), obj[k])
        if(Object.prototype.toString.call(obj[k]) === '[object Object]') {
          o.isDict = true
        }
        l.push(o)
      }
      return l
    }
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
  width: 100%;
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
.msgClass .el-table::before {
  height: 0px;
}
</style>
