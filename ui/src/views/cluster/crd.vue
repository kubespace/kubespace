<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        :data="crds"
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
          min-width="28"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="version"
          label="版本"
          min-width="18"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="scope"
          label="scope"
          min-width="18"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="create_time"
          label="创建时间"
          min-width="12"
          show-overflow-tooltip>
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.create_time) }}
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
                <el-dropdown-item @click.native.prevent="nameClick(scope.row)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="getCrdYaml(scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <!-- <el-dropdown-item @click.native.prevent="deleteNodes([{namespace: scope.row.namespace, name: scope.row.name}])">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                  <span style="margin-left: 5px;">删除</span>
                </el-dropdown-item> -->
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
        <el-button plain @click="updateNode()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { ResType, listResource, getResource } from '@/api/cluster/resource'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'

export default {
  name: 'CRD',
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
        titleName: ["CRD"],
        maxHeight: window.innerHeight - this.$contentHeight,
        loading: true,
        originCrds: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delCrds: [],
      }
  },
  created() {
    this.fetchData()
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
    crds: function() {
      let dlist = []
      for (let p of this.originCrds) {
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
      this.originCrds = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listResource(cluster, ResType.CRD).then(response => {
          this.loading = false
          this.originCrds = response.data || []
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    nameClick: function(obj) {
      this.$router.push({name: 'cr', params: {group: obj.group, resource: obj.resource, version: obj.version}})
    },
    getCrdYaml: function(name) {
      this.yamlName = ""
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!name) {
        Message.error("获取Node名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getResource(cluster, ResType.CRD, "", name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteCrds: function(nodes) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( nodes.length <= 0 ){
        Message.error("请选择要删除的Nodes")
        return
      }
      let params = {
        resources: nodes
      }
    },
  }
}
</script>
