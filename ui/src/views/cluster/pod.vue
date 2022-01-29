<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="pods"
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
          min-width="170"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.namespace, scope.row.name)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="90"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="containerNum"
          label="容器"
          min-width="65"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.init_containers">
            <el-tooltip :content="`${c.name} (${c.status})`" placement="top" v-for="c in scope.row.init_containers" :key="c.name">
              <svg-icon style="margin-top: 7px;" :class="containerClass(c.status)" icon-class="square" />
            </el-tooltip>
            </template>
            <el-tooltip :content="`${c.name} (${c.status})`" placement="top" v-for="c in scope.row.containers" :key="c.name">
              <svg-icon style="margin-top: 7px;" :class="containerClass(c.status)" icon-class="square" />
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column
          prop="restarts"
          label="重启"
          min-width="45"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="node_name"
          label="节点"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="ip"
          label="IP"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="controlled"
          label="控制器"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="created"
          label="创建时间"
          min-width="135"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="status"
          label="状态"
          min-width="60"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span :class="podStatusClass(scope.row.status)" style="font-weight: 450">{{ scope.row.status }}</span>
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
                <el-dropdown-item @click.native.prevent="nameClick(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$updatePerm()" @click.native.prevent="getPodYaml(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$deletePerm()" @click.native.prevent="deletePods([{namespace: scope.row.namespace, name: scope.row.name}])">
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
        <el-button plain @click="updatePod()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listPods, getPod, deletePods, updatePod, buildPods } from '@/api/pods'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'

export default {
  name: 'Pod',
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
        titleName: ["Pods"],
        maxHeight: window.innerHeight - 150,
        loading: true,
        originPods: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delPods: [],
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
    podsWatch: function (newObj) {
      if (newObj) {
        let newUid = newObj.resource.metadata.uid
        let newRv = newObj.resource.metadata.resourceVersion
        if (newObj.event === 'add') {
          this.originPods.push(buildPods(newObj.resource))
        } else if (newObj.event === 'update') {
          for (let i in this.originPods) {
            let p = this.originPods[i]
            if (p.uid === newUid && p.resource_version < newRv) {
              let newPod = buildPods(newObj.resource)
              this.$set(this.originPods, i, newPod)
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originPods = this.originPods.filter(( { uid } ) => uid !== newUid)
        }
      }
    }
  },
  computed: {
    pods: function() {
      let plist = []
      for (let p of this.originPods) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        plist.push(p)
      }
      return plist
    },
    podsWatch: function() {
      return this.$store.getters["ws/podWatch"]
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originPods = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listPods(cluster).then(response => {
          this.loading = false
          this.originPods = response.data
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
    nameClick: function(namespace, name) {
      this.$router.push({name: 'podsDetail', params: {namespace: namespace, podName: name}})
    },
    containerClass: function(status) {
      if (status === 'running') return 'running-class'
      if (status === 'terminated') return 'terminate-class'
      if (status === 'waiting') return 'waiting-class'
    },
    getPodYaml: function(namespace, podName) {
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
      if (!podName) {
        Message.error("获取Pod名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getPod(cluster, namespace, podName, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlName = podName
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deletePods: function(pods) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( pods.length <= 0 ){
        Message.error("请选择要删除的POD")
        return
      }
      let params = {
        resources: pods
      }
      deletePods(cluster, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    updatePod: function() {
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
        Message.error("获取POD参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      updatePod(cluster, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delPodsFunc: function() {
      console.log('delete ', this.delPods)
      if (this.delPods.length > 0){
        let delPods = []
        for (var p of this.delPods) {
          delPods.push({namespace: p.namespace, name: p.name})
        }
        this.deletePods(delPods)
      }
    },
    handleSelectionChange(val) {
      console.log(val);
      this.delPods = val;
      if (val.length > 0){
        this.delFunc = this._delPodsFunc
      } else {
        this.delFunc = undefined
      }
      // this.multipleSelection = val;
    },
    podStatusClass(status) {
      if(status=='Running') return 'running-class'
      if(status=='Pending') return 'waiting-class'
      if(status=='Succeeded') return 'succeeded-class'
      if(status=='Failed') return 'error-class'
      if(status=='Unknown') return 'terminate-class'
    }
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

.error-class {
  color: #F56C6C;
}

.succeeded-class {
  color: #409EFF;
}
</style>

<style>
.el-dialog__body {
  padding-top: 5px;
  padding-bottom: 5px;
}
</style>
