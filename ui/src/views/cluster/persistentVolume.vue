<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <el-table
        ref="multipleTable"
        :data="persistentVolume"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort="{ prop: 'name' }"
        @selection-change="handleSelectionChange"
        row-key="uid"
      >
        <el-table-column type="selection" width="45"> </el-table-column>
        <el-table-column
          prop="name"
          label="名称"
          min-width="40"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.name)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="capacity"
          label="容量"
          min-width="20"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="reclaim_policy"
          label="重声明策略"
          min-width="30"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="access_modes"
          label="访问模式"
          min-width="35"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="claim"
          label="存储声明"
          min-width="45"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span v-if="scope.row.claim">
              {{ scope.row.claim_namespace + '/' + scope.row.claim }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="storage_class"
          label="存储类"
          min-width="35"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="status"
          label="状态"
          min-width="25"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="create_time"
          label="创建时间"
          min-width="45"
          show-overflow-tooltip>
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.create_time) }}
          </template>
        </el-table-column>
        <el-table-column label="" show-overflow-tooltip width="45">
          <template slot-scope="scope">
            <el-dropdown size="medium">
              <el-link :underline="false">
                <svg-icon style="width: 1.3em; height: 1.3em" icon-class="operate"/>
              </el-link>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item @click.native.prevent="nameClick(scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em;"
                    icon-class="detail"/>
                  <span style="margin-left: 5px">详情</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="getPersistentVolumeYaml(scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em;"
                    icon-class="edit"/>
                  <span style="margin-left: 5px">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="deletePvs([{name: scope.row.name}])">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em;"
                    icon-class="delete"/>
                  <span style="margin-left: 5px">删除</span>
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
        <el-button plain @click="updatePv()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from "@/views/components";
import { Message } from "element-ui";
import { ResType, listResource, getResource, delResource, updateResource, createResource } from '@/api/cluster/resource'

export default {
  name: "PersistentVolume",
  components: {
    Clusterbar,
    Yaml,
  },
  data() {
    return {
      titleName: ["PersistentVolumes"],
      originPersistentVolumes: [],
      search_name: "",
      search_ns: [],
      cellStyle: { border: 0 },
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      yamlDialog: false,
      yamlName: "",
      yamlValue: "",
      yamlLoading: true,
      delFunc: undefined,
      delPvs: [],
    }
  },
  created() {
    this.fetchData();
  },
  watch: {
    cluster: function() {
      this.fetchData()
    }
  },
  computed: {
    persistentVolume: function () {
      let data = [];
      for (let c of this.originPersistentVolumes) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(c.namespace) < 0) continue;
        if (this.search_name && !c.name.includes(this.search_name)) continue;
        data.push(c);
      }
      return data;
    },
    cluster() {
      return this.$store.state.cluster
    }
  },
  methods: {
    nameClick: function (name) {
      this.$router.push({
        name: "pvDetail",
        params: { persistentVolumeName: name },
      });
    },
    nsSearch: function (vals) {
      this.search_ns = [];
      for (let ns of vals) {
        this.search_ns.push(ns);
      }
    },
    nameSearch: function (val) {
      this.search_name = val;
    },
    fetchData: function () {
      this.loading = true;
      this.originConfigMaps = [];
      const cluster = this.$store.state.cluster;
      if (cluster) {
        listResource(cluster, ResType.PersistentVolume)
          .then((response) => {
            this.loading = false;
            this.originPersistentVolumes = response.data ? response.data : []
          })
          .catch(() => {
            this.loading = false;
          });
      } else {
        this.loading = false;
        Message.error("获取集群异常，请刷新重试.");
      }
    },
    getPersistentVolumeYaml: function (name) {
      const cluster = this.$store.state.cluster;
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试");
        return;
      }
      if (!name) {
        Message.error("获取名称参数异常，请刷新重试");
        return;
      }
      this.yamlLoading = true;
      this.yamlDialog = true;
      getResource(cluster, ResType.PersistentVolume, name, "yaml")
        .then((response) => {
          this.yamlLoading = false;
          this.yamlValue = response.data;
          this.yamlName = name;
        })
        .catch(() => {
          this.yamlLoading = false;
        });
    },
    updatePv: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!this.yamlName) {
        Message.error("获取存储卷参数异常，请刷新重试")
        return
      }
      updateResource(cluster, ResType.PersistentVolume, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    deletePvs: function(pvs) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( pvs.length <= 0 ){
        Message.error("请选择要删除的存储卷")
        return
      }
      let params = {
        resources: pvs
      }
      delResource(cluster, ResType.PersistentVolume, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    _delPvsFunc: function() {
      if (this.delPvs.length > 0){
        let delPvs = []
        for (var p of this.delPvs) {
          delPvs.push({name: p.name})
        }
        this.deletePvs(delPvs)
      }
    },
    handleSelectionChange(val) {
      this.delPvs = val;
      if (val.length > 0){
        this.delFunc = this._delPvsFunc
      } else {
        this.delFunc = undefined
      }
    },
  },
};
</script>
<style lang="scss" scoped>

.name-class {
  cursor: pointer;
}
.name-class:hover {
  color: #409eff;
}
</style>