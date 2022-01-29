<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch"
    :createFunc="createFunc" createDisplay="安装应用"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        :data="releases"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort = "{prop: 'name'}"
        row-key="uid"
        >
        <el-table-column
          prop="name"
          label="名称"
          min-width="25"
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
          min-width="25"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="chart_version"
          label="Chart版本"
          min-width="30"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="app_version"
          label="App版本"
          min-width="23"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="version"
          label="修订"
          min-width="15"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="status"
          label="状态"
          min-width="20"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="last_deployed"
          label="上次部署"
          min-width="25"
          show-overflow-tooltip>
          <template slot-scope="scope">
              {{ dateFormat(scope.row.last_deployed) }}
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
                <el-dropdown-item v-if="$updatePerm()" @click.native.prevent="yamlDialog = true; getUpdateRelease(scope.row.name, scope.row.namespace)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">升级</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$updatePerm()" @click.native.prevent="deleteRelease(scope.row)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                  <span style="margin-left: 5px;">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog title="升级" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="45px"
      @close="yamlDialog=false; updateValues={name: '', namespace: '', config: '', values: ''}; yamlChange=1;">
      <el-button-group style="margin-bottom: 10px;">
        <el-button :type="yamlChange ? 'primary': ''" size="small" @click="yamlChange=1;  ">当前配置</el-button>
        <el-button :type="yamlChange ? '': 'primary'" size="small" @click="yamlChange=0;  yamlValue=updateValues.values">原始values(只读)</el-button>
      </el-button-group>
      <template v-if="yamlDialog">
        <yaml v-show="yamlChange" v-model="updateValues.config" :loading="yamlLoading" :readOnly="false"></yaml>
        <yaml v-show="!yamlChange" v-model="yamlValue" :loading="yamlLoading" :readOnly="true"></yaml>
      </template>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
        <el-button plain @click="updateNode()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listReleases, deleteRelease, getRelease } from '@/api/app'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'
import { dateFormat } from '@/utils/utils'

export default {
  name: 'Application',
  components: {
    Clusterbar,
    Yaml
  },
  data() {
      return {
        yamlDialog: false,
        yamlName: "",
        yamlValue: "",
        yamlLoading: true,
        cellStyle: {border: 0},
        titleName: ["Applications"],
        maxHeight: window.innerHeight - 150,
        loading: true,
        originReleases: [],
        search_ns: [],
        search_name: '',
        yamlChange: 1,
        updateValues: {
          name: "",
          namespace: "",
          config: "",
          values: "",
        }
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
  },
  computed: {
    releases: function() {
      let dlist = []
      for (let p of this.originReleases) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        
        dlist.push(p)
      }
      return dlist
    }
  },
  methods: {
    dateFormat,
    fetchData: function() {
      this.loading = true
      const cluster = this.$store.state.cluster
      if (cluster) {
        listReleases(cluster).then(response => {
          this.originReleases = response.data
          this.loading = false
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
    getUpdateRelease: function(name, namespace) {
      this.yamlLoading = true
      const cluster = this.$store.state.cluster
      this.updateValues = {config: "", values: "", name: name, namespace: namespace}
      this.yamlValue = ""
      if (cluster) {
        var params = {
          name: name,
          namespace: namespace,
          get_option: 'values'
        }
        getRelease(cluster, params).then(response => {
          this.yamlLoading = false
          this.updateValues.config = response.data.config
          this.updateValues.values = response.data.values
          this.yamlValue = this.updateValues.config
        }).catch(() => {
          this.yamlLoading = false
        })
      } else {
        this.yamlLoading = false
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
      this.$router.push({name: 'appDetail', params: {namespace: namespace, appName: name}})
    },
    deleteRelease: function(release) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      let params = {
        name: release.name,
        namespace: release.namespace,
      }
      this.$confirm('是否确认删除应用' + release.name + '?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        deleteRelease(cluster, params).then(() => {
          Message.success("删除成功")
          this.fetchData();
        }).catch(() => {
          // console.log(e)
        })
      })
    },
    createFunc() {
      this.$router.push({name: 'appCreate'})
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
