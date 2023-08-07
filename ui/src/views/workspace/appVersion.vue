<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :titleLink="['workspaceApp']" />
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="originVersions"
        class="table-fix"
        :cell-style="cellStyle"
        v-loading="loading"
        :default-sort = "{prop: 'name'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column prop="name" label="名称" show-overflow-tooltip min-width="13">
          <template slot-scope="scope">
            <span>
              {{ scope.row.package_name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="package_version" label="chart版本" show-overflow-tooltip min-width="12">
        </el-table-column>
        <el-table-column prop="app_version" label="app版本" show-overflow-tooltip min-width="12">
        </el-table-column>
        <el-table-column prop="description" label="版本说明" show-overflow-tooltip min-width="20">
        </el-table-column>
        <el-table-column prop="create_user" label="创建人" show-overflow-tooltip min-width="10">
          <template slot-scope="scope">
            {{ scope.row.create_user }}
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.create_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :disabled="!$editorRole(scope.row.id)" :underline="false" type="primary" style="margin-right: 13px" @click="openEditApp(scope.row)">编辑</el-link>
              <el-link :underline="false" type="primary" style="margin-right: 13px" :href="'/api/v1/project/apps/download?path='+scope.row.chart_path">下载</el-link>
              <el-link v-if="originApp.app_version_id && scope.row.id != originApp.app_version_id" :disabled="!$editorRole(scope.row.id)" :underline="false" type="danger" @click="handleDeleteAppVersion(scope.row.id, scope.row.package_version)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>

    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { listAppVersions, getApp, deleteAppVersion } from "@/api/project/apps";
import { Message } from "element-ui";

export default {
  name: "appVersion",
  components: {
    Clusterbar,
  },
  mounted: function () {
    const that = this;
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - this.$contentHeight;
        that.maxHeight = heightStyle;
      })();
    };
  },
  data() {
    return {
      maxHeight: window.innerHeight - this.$contentHeight,
      cellStyle: { border: 0 },
      titleName: ["应用管理"],
      loading: true,
      originVersions: [],
      originApp: {},
      search_name: "",
    };
  },
  created() {
    this.getApp();
    this.fetchVersions();
  },
  computed: {
    projectId() {
      return this.$route.params.workspaceId
    },
    appId() {
      return this.$route.params.appId
    },
  },
  methods: {
    fetchVersions() {
      this.loading = true
      listAppVersions({scope: 'project_app', scope_id: this.appId}).then((resp) => {
        this.originVersions = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    getApp() {
      getApp(this.appId).then((response) => {
        this.originApp = response.data || {};
        this.titleName = ["应用管理", this.originApp.name]
      }).catch(() => {
      })
    },
    
    nameSearch(val) {
      this.search_name = val;
    },
    openEditApp(appVersion) {

      if(appVersion.from == 'space') {
        this.$router.push({name: 'workspaceEditApp', params: {appVersionId: appVersion.id}})
      } else {
        this.$router.push({name: 'workspaceEditImportApp', params: {appVersionId: appVersion.id}})
      }
      // this.$router.push({name: 'workspaceEditApp', params: {appVersionId: id}})
    },
    handleDeleteAppVersion(id, package_version) {
      if(!id) {
        Message.error("获取应用版本id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${package_version}」此应用版本?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        deleteAppVersion(id).then(() => {
          this.loading = false
          Message.success("删除应用版本成功")
          this.fetchVersions()
        }).catch((err) => {
          this.loading = false
          console.log(err)
        });
      }).catch(() => {       
      });
    },
    
  },
};
</script>


<style lang="scss" scoped>
@import "~@/styles/variables.scss";

.table-fix {
  height: calc(100% - 100px);
}

</style>
