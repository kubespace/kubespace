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
import { ResType, listResource, watchResource } from '@/api/cluster/resource'
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
        maxHeight: window.innerHeight - this.$contentHeight,
        loading: true,
        originEvents: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delEvents: [],
        clusterSSE: undefined
      }
  },
  created() {
    this.fetchData()
  },
  beforeDestroy() {
    if(this.clusterSSE) this.clusterSSE.disconnect()
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
    events: function() {
      let dlist = []
      for (let p of this.originEvents) {
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
      this.originEvents = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listResource(cluster, ResType.Event).then(response => {
          this.loading = false
          this.originEvents = response.data ? response.data : []
          this.fetchSSE()
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
    fetchSSE() {
      if(!this.clusterSSE) {
        this.clusterSSE = watchResource(this.$sse, this.cluster, ResType.Event, this.sseWatch, {process: true})
      }
    },
    sseWatch(newObj) {
      if (newObj) {
        console.log(newObj)
        let newUid = newObj.resource.uid
        let newRv = newObj.resource.resource_version
        if (newObj.event === 'add') {
          for(let i in this.originEvents) {
            let d = this.originEvents[i]
            if (d.uid === newUid) return
          }
          this.originEvents.push(newObj.resource)
        } else if (newObj.event === 'update') {
          for (let i in this.originEvents) {
            let d = this.originEvents[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newDp = newObj.resource
                this.$set(this.originEvents, i, newDp)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originEvents = this.originEvents.filter(( { uid } ) => uid !== newUid)
        }
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
