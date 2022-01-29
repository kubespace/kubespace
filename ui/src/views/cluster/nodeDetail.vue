<template>
  <div>
    <clusterbar :titleName="titleName" :editFunc="getNodeYaml" />
    <div class="dashboard-container" v-loading="loading" v-if="node">
      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" class="pod-item" label-width="120px" style="margin: 15px 10px 20px 10px;">
          <el-form-item label="名称">
            <span>{{ node.metadata.name }}</span>
          </el-form-item>
          <el-form-item label="创建时间">
            <span>{{ node.metadata.creationTimestamp }}</span>
          </el-form-item>
          <el-form-item label="IP地址">
            <span>{{ nodeInternalIp(node) }}</span>
          </el-form-item>
          <el-form-item label="版本">
            <span>{{ node.status.nodeInfo.kubeletVersion }}</span>
          </el-form-item>
          <el-form-item label="操作系统">
            <span>{{ node.status.nodeInfo.osImage }}</span>
          </el-form-item>
          <el-form-item label="容器运行时">
            <span>{{ node.status.nodeInfo.containerRuntimeVersion }}</span>
          </el-form-item>
        </el-form>
      </div>

      <el-tabs value="conditions" style="padding: 0px 8px;">
        <el-tab-pane label="状态" name="conditions">
          <div class="msgClass">
            <el-table
              v-if="node.status.conditions && node.status.conditions.length > 0"
              :data="node.status.conditions"
              class="table-fix"
              tooltip-effect="dark"
              style="width: 100%"
              :cell-style="cellStyle"
              :default-sort = "{prop: 'lastProbeTime'}"
              >
              <el-table-column
                prop="type"
                label="类型"
                min-width="30"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="status"
                label="状态"
                min-width="20"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="reason"
                label="原因"
                min-width="50"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>
                    {{ scope.row.reason ? scope.row.reason : "—" }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column
                prop="message"
                label="信息"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>
                    {{ scope.row.message ? scope.row.message : "—" }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column
                label="触发时间"
                min-width="40"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>
                    {{ scope.row.lastProbeTime ? scope.row.lastProbeTime : scope.row.lastTransitionTime }}
                  </span>
                </template>
              </el-table-column>
            </el-table>
            <div v-else style="padding: 25px 15px ; color: #909399; text-align: center">暂无数据</div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="地址" name="addresses">
          <div class="msgClass">
            <el-table
              :data="node.status.addresses"
              class="table-fix"
              tooltip-effect="dark"
              style="width: 100%"
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
                prop="address"
                label="地址"
                min-width=""
                show-overflow-tooltip>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
        <el-tab-pane label="信息" name="nodeInfo">
          <div class="msgClass">
            <el-form label-position="left" class="pod-item" label-width="180px" style="box-shadow: 0 0 0 0; margin-top: 0px;">
              <template v-for="(val, key) in node.status.nodeInfo">
                <el-form-item :label="key" :key="key">
                  <span>{{ val }}</span>
                </el-form-item>
              </template>
            </el-form>
          </div>
        </el-tab-pane>
        <el-tab-pane label="污点" name="taints">
          <div class="msgClass">
            <el-table
              :data="node.spec.taints"
              class="table-fix"
              tooltip-effect="dark"
              style="width: 100%"
              :cell-style="cellStyle"
              :default-sort = "{prop: 'event_time', order: 'descending'}"
              >
              <el-table-column
                prop="key"
                label="键"
                min-width="25"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="value"
                label="值"
                min-width="25"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="effect"
                label="影响"
                min-width="25"
                show-overflow-tooltip>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
        <el-tab-pane label="镜像" name="images">
          <div class="msgClass">
            <el-table
              :data="node.status.images"
              class="table-fix"
              tooltip-effect="dark"
              style="width: 100%"
              :cell-style="cellStyle"
              :default-sort = "{prop: 'event_time', order: 'descending'}"
              >
              <el-table-column
                prop="names"
                label="名称"
                min-width=""
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <div v-for="n in scope.row.names" :key="n">
                    {{ n }}
                  </div>
                </template>
              </el-table-column>
              <el-table-column
                prop="sizeBytes"
                label="大小"
                min-width="15"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>{{ sizeStr(scope.row.sizeBytes) }}</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
        <el-tab-pane label="事件" name="events">
          <div class="msgClass">
            <el-table
              v-if="nodeEvents && nodeEvents.length > 0"
              :data="nodeEvents"
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
                    {{ scope.row.reason ? scope.row.reason : "—" }}
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
                    {{ scope.row.message ? scope.row.message : "—" }}
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
          <el-button plain @click="updateNode" size="small">确 定</el-button>
        </span>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { getNode, updateNode } from '@/api/nodes'
import { listEvents } from '@/api/event'
import { Message } from 'element-ui'

export default {
  name: 'nodeDetail',
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
      originNode: {},
      selectContainer: '',
      eventLoading: true,
      nodeEvents: []
    }
  },
  created() {
    this.fetchData()
  },
  watch: {},
  computed: {
    titleName: function() {
      return ['Node', this.nodeName]
    },
    nodeName: function() {
      return this.$route.params ? this.$route.params.nodeName : ''
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    node: function() {
      console.log(this.originNode)
      return this.originNode
    },
  },
  methods: {
    fetchData: function() {
      this.originNode = {}
      this.loading = true
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error('获取集群参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.nodeName) {
        Message.error('获取node名称参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      getNode(cluster, this.nodeName).then(response => {
        this.loading = false
        this.originNode = response.data
        listEvents(cluster, this.originNode.metadata.uid).then(response => {
          this.eventLoading = false
          if (response.data) {
            this.nodeEvents = response.data.length > 0 ? response.data : []
          }
        }).catch(() => {
          this.eventLoading = false
        })
      }).catch(() => {
        this.loading = false
      })
    },
    getNodeYaml: function() {
      if (!this.nodeName) {
        Message.error('获取node参数异常，请刷新重试')
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
      getNode(cluster, this.nodeName, 'yaml')
        .then((response) => {
          this.yamlLoading = false
          this.yamlValue = response.data
        })
        .catch(() => {
          this.yamlLoading = false
        })
    },
    updateNode: function() {
      if (!this.node) {
        Message.error("获取node参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      updateNode(cluster, this.nodeName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    nodeInternalIp: function(node) {
      for (let a of node.status.addresses) {
		    if (a.type == "InternalIP") {
			    return a.address
		    }
      }
      return ''
    },
    sizeStr: function(size) {
      let s = size
      if(s / 1024 < 1) {
        return size + 'B'
      }
      s /= 1024
      if (s / 1024 < 1) {
        return s.toFixed(2) + 'KB'
      }
      s /= 1024
      return s.toFixed(2) + 'MB'
    }
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
  margin: 20px 5px 30px 5px;
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
