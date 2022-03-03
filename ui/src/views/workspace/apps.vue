<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="openCreateApp" createDisplay="创建应用"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="originApps"
        class="table-fix"
        :cell-style="cellStyle"
        v-loading="loading"
        :default-sort = "{prop: 'name'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column prop="name" label="名称" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.id)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="package_version" label="版本" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="type" label="类型" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ typeNameMap[scope.row.type] }}
          </template>
        </el-table-column>
        <el-table-column prop="update_user" label="操作人" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="update_time" label="更新时间" show-overflow-tooltip min-width="20">
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.update_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            <div class="status-class" :style="{'border-color': statusColorMap[scope.row.status], 'background-color': statusColorMap[scope.row.status]}"></div>
            <span :style="{'font-weight': 430}">{{ statusNameMap[scope.row.status] }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :underline="false" class="operator-btn"
                v-if="scope.row.status=='UnInstall'" @click="openInstallFormDialog(scope.row)">安装</el-link>
              <el-link :underline="false" class="operator-btn"
                v-if="scope.row.status!='UnInstall'" @click="openInstallFormDialog(scope.row)">升级</el-link>
              <el-link :underline="false" class="operator-btn"
                @click="openEditApp(scope.row.app_version_id)">编辑</el-link>
              <el-link :underline="false" class="operator-btn" style="color: #F56C6C"
                v-if="scope.row.status!='UnInstall'" @click="handleDestroyApp(scope.row.id, scope.row.name)">销毁</el-link>
              <el-link :underline="false" style="color: #F56C6C" v-if="scope.row.status=='UnInstall'"
                @click="handleDeleteApp(scope.row.id, scope.row.name)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="updateFormVisible ? '升级应用' : '安装应用'" :visible.sync="installFormVisible"
      @close="closeFormDialog" :destroy-on-close="true" :close-on-click-modal="false">
        <div v-loading="installLoading">
          <div class="dialogContent" style="">
            <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
              <el-form-item label="名称" prop="" autofocus>
                <span>{{ form.name }}</span>
              </el-form-item>
              <div v-loading="fetchVersionLoading">
                <el-form-item label="安装版本" prop="" :required="true">
                  <el-select v-model="form.app_version_id" placeholder="请选择应用版本" size="small" style="width: 100%;"
                    @change="changeInstallAppVersion">
                    <el-option
                      v-for="item in appVersions"
                      :key="item.id"
                      :label="item.package_version"
                      :value="item.id">
                    </el-option>
                  </el-select>
                </el-form-item>
                <el-row style="margin-bottom: 17px;">
                  <el-col :span="7"><div style="color: #909399">负载容器</div></el-col>
                  <el-col :span="12"><div style="color: #909399">镜像</div></el-col>
                  <el-col :span="5"><div style="color: #909399">标签</div></el-col>
                </el-row>
                <div v-for="(v, k) of form.values_dict.workloads ? form.values_dict.workloads : {}" :key="k">
                  <el-row style="margin-bottom: 17px;" v-for="(cv, ck) of v.containers" :key="ck">
                    <el-col :span="7">
                      <div style="padding-top: 6px; padding-right: 3px;">{{ k + "/" + ck }}</div>
                    </el-col>
                    <el-col :span="12" style="padding-right: 15px">
                      <el-input v-model="cv.image" autocomplete="off" placeholder="请输入容器镜像" size="small"></el-input>
                    </el-col>
                    <el-col :span="5">
                      <el-input v-model="cv.tag" autocomplete="off" placeholder="镜像Tag" size="small"></el-input>
                    </el-col>
                  </el-row>
                </div>
              </div>
            </el-form>
          </div>
          <div slot="footer" class="dialogFooter" style="padding-top: 25px;">
            <el-button @click="installFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
            <el-button type="primary" @click="updateFormVisible ? handleUpdateApp() : handleInstallApp()" >
              {{ installFormVisible ? '安 装' : '升 级' }}
            </el-button>
          </div>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { listApps, listAppVersions, installApp, destroyApp } from "@/api/project/apps";
import { Message } from "element-ui";
import yaml from 'js-yaml'

export default {
  name: "projectApps",
  components: {
    Clusterbar,
  },
  mounted: function () {
    const that = this;
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 150;
        that.maxHeight = heightStyle;
      })();
    };
  },
  data() {
    return {
      maxHeight: window.innerHeight - 150,
      cellStyle: { border: 0 },
      titleName: ["应用管理"],
      loading: true,
      installFormVisible: false,
      updateFormVisible: false,
      clusters: [],
      namespaces: [],
      appVersions: [],
      fetchVersionLoading: false,
      installLoading: false,
      form: {
        id: "",
        name: "",
        app_version_id: "",
        values_dict: {},
      },
      rules: {
        // name: [{ required: true, message: '请输入空间名称', trigger: 'blur' },],
      },
      statusNameMap: {
        "UnInstall": "未安装",
        "UnReady": "未就绪",
        "RunningFault": "运行故障",
        "Running": "运行中"
      },
      statusColorMap: {
        "UnInstall": "",
        "UnReady": "#E6A23C",
        "RunningFault": "#F56C6C",
        "Running": "#67C23A"
      },
      typeNameMap: {
        "ordinary_app": "普通应用",
        "middleware": "中间件",
        "import_app": "导入应用"
      },
      originApps: [],
      search_name: "",
    };
  },
  created() {
    this.fetchApps();
  },
  computed: {
    projectId() {
      return this.$route.params.workspaceId
    },
  },
  methods: {
    nameClick: function(id) {
      this.$router.push({name: 'workspaceOverview', params: {'workspaceId': id}})
    },
    fetchApps() {
      this.loading = true
      listApps({project_id: this.projectId}).then((resp) => {
        this.originApps = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    handleInstallApp() {
      if(!this.form.id) {
        Message.error("获取安装应用失败，请刷新重试");
        return
      }
      if(!this.form.app_version_id) {
        Message.error("请选择要安装的应用版本");
        return
      }
      let values = yaml.dump(this.form.values_dict)
      let data = {
        project_app_id: this.form.id, 
        app_version_id: this.form.app_version_id, 
        values: values,
      }
      this.installLoading = true
      installApp(data).then(() => {
        this.installLoading = false
        this.installFormVisible = false;
        Message.success("安装应用成功")
        this.fetchApps()
      }).catch((err) => {
        this.installLoading = false
      });
    },
    handleDestroyApp(id, name) {
      if(!id) {
        Message.error("获取销毁应用id失败，请刷新重试");
        return
      }
      this.$confirm(`请确认是否销毁「${name}」此应用?`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
      }).then(() => {
        this.loading = true
        destroyApp({project_app_id: id}).then(() => {
          Message.success("销毁应用成功")
          this.fetchApps()
        }).catch((err) => {
          console.log(err)
          this.loading = false
        });
      }).catch(() => {       
      });
    },
    handleUpdateApp() {
      if(!this.form.id) {
        Message.error("获取空间id参数异常，请刷新重试");
        return
      }
      let project = {
        name: this.form.name, 
        description: this.form.description, 
        owner: this.form.owner
      }
      updateSecret(this.form.id, project).then(() => {
        this.installFormVisible = false;
        Message.success("更新项目空间成功")
        this.fetchApps()
      }).catch((err) => {
        console.log(err)
      });
    },
    handleDeleteApp(id, name) {
      if(!id) {
        Message.error("获取密钥id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${name}」此项目空间?`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          deleteSecret(id).then(() => {
            Message.success("删除项目空间成功")
            this.fetchApps()
          }).catch((err) => {
            console.log(err)
          });
        }).catch(() => {       
        });
    },
    nameSearch(val) {
      this.search_name = val;
    },
    openInstallFormDialog(app, isUpdate) {
      if(isUpdate) this.updateFormVisible = true
      this.appVersions = []
      this.form = {
        id: app.id,
        name: app.name,
        app_version_id: app.app_version_id,
        values_dict: {}
      }
      this.installFormVisible = true;
      this.fetchVersionLoading = true;
      listAppVersions({scope: "project_app", scope_id: app.id}).then((resp) => {
        this.appVersions = resp.data ? resp.data : []
        this.changeInstallAppVersion(app.app_version_id)
        this.fetchVersionLoading = false
      }).catch((err) => {
        this.fetchVersionLoading = false
      })
    },
    changeInstallAppVersion(app_version_id) {
      this.form.values_dict = {}
      if(!this.appVersions || this.appVersions.length <= 0) return
      for(let v of this.appVersions) {
        if(v.id == app_version_id) {
          this.form.values_dict = yaml.load(v.values)
          return
        }
      }
    },
    closeFormDialog() {
      this.form = {
        id: "",
        name: "",
        app_version_id: "",
        values_dict: {}
      }
      this.appVersions = []
      this.updateFormVisible = false; 
      this.installFormVisible = false;
    },
    openCreateApp() {
      this.$router.push({name: 'workspaceCreateApp'})
    },
    openEditApp(id) {
      this.$router.push({name: 'workspaceEditApp', params: {appVersionId: id}})
    }
  },
};
</script>


<style lang="scss" scoped>
@import "~@/styles/variables.scss";

.table-fix {
  height: calc(100% - 100px);
}

.status-class {
  height: 13px; 
  width: 13px; 
  border: 1px solid; 
  border-color:rgb(121, 123, 129); 
  background-color: rgb(121, 123, 129);  
  display: inline-block;
  vertical-align: middle; 
  border-radius: 25px; 
  margin: 0px 5px 3px 0px;
}

.operator-btn {
  margin-right: 15px;
  color:#0c81f5
}

</style>
