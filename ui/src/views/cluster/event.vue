<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        :data="events"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort = "{prop: 'event_time', order: 'descending'}"
        @selection-change="handleSelectionChange"
        row-key="uid"
        >
        <!-- <el-table-column
          type="selection"
          width="45">
        </el-table-column> -->
        <el-table-column
          prop="type"
          label="类型"
          min-width="20"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
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
          min-width="40"
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
          min-width="100"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ scope.row.message ? scope.row.message : "—" }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="count"
          label="次数"
          min-width="20"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="event_time"
          label="触发时间"
          min-width="50"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ $dateFormat(scope.row.event_time) }}
            </span>
          </template>
        </el-table-column>
        <!-- <el-table-column
          label=""
          show-overflow-tooltip
          width="45">
          <template slot-scope="scope">
            <el-dropdown size="medium" >
              <el-link :underline="false"><svg-icon style="width: 1.3em; height: 1.3em;" icon-class="operate" /></el-link>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item @click.native.prevent="nameClick(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item>
                <el-dropdown-item @click.native.prevent="getEventYaml(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item @click.native.prevent="deleteEvents([{namespace: scope.row.namespace, name: scope.row.name}])">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                  <span style="margin-left: 5px;">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
        </el-table-column> -->
      </el-table>
    </div>
    <!-- <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
      <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
        <el-button plain @click="updateEvent()" size="small">确 定</el-button>
      </span>
    </el-dialog> -->
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listEvents, buildEvent } from '@/api/event'
import { Message } from 'element-ui'
// import { Yaml } from '@/views/components'

export default {
  name: 'Event',
  components: {
    Clusterbar,
    // Yaml
  },
  data() {
      return {
        // yamlDialog: false,
        // yamlNamespace: "",
        // yamlName: "",
        // yamlValue: "",
        // yamlLoading: true,
        cellStyle: {border: 0},
        titleName: ["Events"],
        maxHeight: window.innerHeight - 150,
        loading: true,
        originEvents: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delEvents: [],
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
    eventsWatch: function (newObj) {
      if (newObj) {
        let newUid = newObj.resource.metadata.uid
        let newRv = newObj.resource.metadata.resourceVersion
        if (newObj.event === 'add') {
          this.originEvents.push(buildEvent(newObj.resource))
        } else if (newObj.event === 'update') {
          for (let i in this.originEvents) {
            let d = this.originEvents[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newDp = buildEvent(newObj.resource)
                this.$set(this.originEvents, i, newDp)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originEvents = this.originEvents.filter(( { uid } ) => uid !== newUid)
        }
      }
    }
  },
  computed: {
    events: function() {
      let dlist = []
      for (let p of this.originEvents) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        
        dlist.push(p)
      }
      return dlist
    },
    eventsWatch: function() {
      return this.$store.getters["ws/eventsWatch"]
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originEvents = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listEvents(cluster).then(response => {
          this.loading = false
          this.originEvents = response.data ? response.data : []
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
