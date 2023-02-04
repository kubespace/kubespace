<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <el-table
        ref="multipleTable"
        :data="storageClass"
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
          min-width="40"
          show-overflow-tooltip>
          <template slot-scope="scope">
              {{ scope.row.name }}
          </template>
        </el-table-column>
        <el-table-column
          prop="provisioner"
          label="供应者"
          min-width="45"
          show-overflow-tooltip>
        </el-table-column>   
        <el-table-column
          prop="reclaim_policy"
          label="重声明策略"
          min-width="25"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="binding_mode"
          label="绑定模式"
          min-width="40"
          show-overflow-tooltip>
        </el-table-column>       
        <el-table-column
          prop="create_time"
          label="创建时间"
          min-width="30"
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
                <!-- <el-dropdown-item @click.native.prevent="nameClick(scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item> -->
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="getStorageClassYaml(scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="deleteScs([{name: scope.row.name}])">
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
        <el-button plain @click="updateSc()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { Message } from 'element-ui'
import { ResType, listResource, getResource, delResource, updateResource, createResource } from '@/api/cluster/resource'
import { listStorageClass, getStorageClass, updateStorageClass,
deleteStorageClasses, buildSc } from '@/api/storage_class'

export default {
  name: "StorageClass",
  components: {
    Clusterbar,
    Yaml
  },
  data() {
    return {
      titleName: ["StorageClasses"],
      originStorageClass: [],
      search_name: "",
      search_ns: [],
      cellStyle: {border: 0},
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      yamlDialog: false,
      yamlName: "",
      yamlValue: "",
      yamlLoading: true,
      delFunc: undefined,
      delScs: [],
    }
  },
  created() {
    this.fetchData()
  },
  watch: {
    cluster: function() {
      this.fetchData()
    }
  },
  computed: {
    storageClass: function() {
      let data = []
      for (let c of this.originStorageClass) {
          if (this.search_ns.length > 0 && this.search_ns.indexOf(c.namespace) < 0) continue
          if (this.search_name && !c.name.includes(this.search_name)) continue
          data.push(c)
      }
      return data
    },
    cluster() {
      return this.$store.state.cluster
    }
  },
  methods: {
    nameClick: function(name) {
      console.log(name)
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
    fetchData: function() {
        this.loading = true
        this.originStorageClass = []
        const cluster = this.$store.state.cluster
        if (cluster) {
          listResource(cluster, ResType.StorageClass).then(response => {
                this.loading = false
                this.originStorageClass = response.data ? response.data : []
            }).catch(() => {
                this.loading = false
            })
        } else {
            this.loading = false
            Message.error("获取集群异常，请刷新重试.")
        }
    },
    getStorageClassYaml: function(name) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!name) {
        Message.error("获取StorageClass名称参数异常，请刷新重试")
        return
      }
      this.yamlLoading = true
      this.yamlDialog = true
      getResource(cluster, ResType.StorageClass, "", name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    updateSc: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!this.yamlName) {
        Message.error("获取存储卷参数异常，请刷新重试")
        return
      }
      this.yamlLoading = true
      updateResource(cluster, ResType.StorageClass, "", this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
        this.yamlLoading = false
        this.yamlDialog = false
        this.fetchData()
      }).catch(() => {
        // console.log(e)
        this.yamlLoading = false
      })
    },
    deleteScs: function(scs) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( scs.length <= 0 ){
        Message.error("请选择要删除的存储类")
        return
      }
      let params = {
        resources: scs
      }
      delResource(cluster, ResType.StorageClass, params).then(() => {
        Message.success("删除成功")
        this.fetchData()
      }).catch(() => {
        // console.log(e)
      })
    },
    _delScsFunc: function() {
      if (this.delScs.length > 0){
        let delScs = []
        for (var p of this.delScs) {
          delScs.push({name: p.name})
        }
        this.deleteScs(delScs)
      }
    },
    handleSelectionChange(val) {
      this.delScs = val;
      if (val.length > 0){
        this.delFunc = this._delScsFunc
      } else {
        this.delFunc = undefined
      }
    },
  }
}
</script>
