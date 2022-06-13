<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" createDisplay="创建应用">
      <div slot="right-btn" style="display: inline-block">
        <el-button size="small" type="primary" @click="openCreateApp" icon="el-icon-plus" :disabled="!$editorRole()">
          创建应用
        </el-button>
        <!-- <el-button @click="openImportStoreAppDialog" type="primary" size="small" v-if="$editorRole()">
          导入应用
        </el-button> -->
        <el-dropdown style="margin-left: 15px;" v-if="$editorRole()">
          <el-button type="primary" size="small">
            导入应用<i class="el-icon-arrow-down el-icon--right"></i>
          </el-button>
          <el-dropdown-menu slot="dropdown">
            <el-dropdown-item @click.native="openImportStoreAppDialog">导入应用商店</el-dropdown-item>
            <el-dropdown-item @click.native="openImportCustomDialog">导入自定义</el-dropdown-item>
          </el-dropdown-menu>
        </el-dropdown>
      </div>
    </clusterbar>
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
          <template slot-scope="scope">
            {{ scope.row.app_version.from == 'space' ? scope.row.app_version.package_version : scope.row.app_version.package_version + ' / ' + scope.row.app_version.app_version }}
          </template>
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
        <el-table-column label="操作" width="170">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 10px"
                v-if="scope.row.status=='UnInstall'" @click="openInstallFormDialog(scope.row)">安装</el-link>
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 10px"
                v-if="scope.row.status!='UnInstall'" @click="openInstallFormDialog(scope.row, true)">升级</el-link>
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 10px"
                v-if="scope.row.app_version.from=='space'" @click="openEditApp(scope.row.app_version_id)">编辑</el-link>
              
              <el-dropdown style="font-size: 13px;">
                <span class="el-dropdown-link operator-btn" :style="{color: !$editorRole() ? '#a0cfff' : ''}">
                  更多操作
                </span>
                <el-dropdown-menu slot="dropdown" :disabled="!$editorRole()">
                  <el-dropdown-item>
                    <el-link :disabled="!$editorRole()" :underline="false" class="operator-btn" type="primary" style="font-weight: 400"
                      @click="openCloneAppDiloag(scope.row, 'project_app')">克隆</el-link>
                  </el-dropdown-item>
                  <el-dropdown-item>
                    <el-link :disabled="!$editorRole()" :underline="false" class="operator-btn" type="primary" style="font-weight: 400"
                      @click="openCloneAppDiloag(scope.row, 'store_app')">发布</el-link>
                  </el-dropdown-item>
                  <el-dropdown-item>
                    <el-link :disabled="!$editorRole()" :underline="false" class="operator-btn" type="primary" style="font-weight: 400"
                      @click="openAppVersion(scope.row.id)">历史版本</el-link>
                  </el-dropdown-item>
                  <el-dropdown-item>
                    <el-link :disabled="!$editorRole()" :underline="false" class="operator-btn" type="danger" style="font-weight: 400"
                      v-if="scope.row.status!='UnInstall'" @click="handleDestroyApp(scope.row.id, scope.row.name)">销毁</el-link>
                    <el-link :disabled="!$editorRole()" :underline="false" type="danger" style="font-weight: 400" v-if="scope.row.status=='UnInstall'"
                      @click="handleDeleteApp(scope.row.id, scope.row.name)">删除</el-link>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="updateFormVisible ? '升级应用' : '安装应用'" :visible.sync="installFormVisible"
        @close="closeFormDialog" :destroy-on-close="true" :close-on-click-modal="false" top="5vh" width="70%">
        <div v-loading="dialogLoading">
          <div class="dialogContent projectApp" style="margin-top: -10px;">
            <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
              <el-form-item label="应用名称" prop="" autofocus>
                <span>{{ form.name }}</span>
              </el-form-item>
              <div v-loading="fetchVersionLoading">
                <el-form-item label="安装版本" prop="" :required="true" style="margin-top: 0px;">
                  <el-select v-model="form.app_version_id" placeholder="请选择应用版本" size="small" style="width: 50%;"
                    @change="changeInstallAppVersion">
                    <el-option
                      v-for="item in appVersions"
                      :key="item.id"
                      :label="item.package_version"
                      :value="item.id">
                    </el-option>
                  </el-select>
                </el-form-item>
                <template v-if="form.from == 'space'">
                  <el-row style="line-height: 38px; margin-top: 10px;">
                    <el-col :span="7"><div style="padding-left: 10px; background-color: rgb(245, 247, 250);color: #909399">负载容器</div></el-col>
                    <el-col :span="11"><div style="background-color: rgb(245, 247, 250);color: #909399">镜像</div></el-col>
                    <el-col :span="6"><div style="background-color: rgb(245, 247, 250);color: #909399">标签</div></el-col>
                  </el-row>
                  <div v-for="(v, k) of form.values_dict.workloads ? form.values_dict.workloads : {}" :key="k">
                    <el-row style="margin-top: 10px;" v-for="(cv, ck) of v.containers" :key="ck">
                      <el-col :span="7">
                        <div style="padding-left: 10px;padding-top: 6px; padding-right: 3px;">{{ k + "/" + ck }}</div>
                      </el-col>
                      <el-col :span="11" style="padding-right: 15px">
                        <el-input v-model="cv.image" autocomplete="off" placeholder="请输入容器镜像" size="small"></el-input>
                      </el-col>
                      <el-col :span="6">
                        <el-input v-model="cv.tag" autocomplete="off" placeholder="镜像Tag" size="small"></el-input>
                      </el-col>
                    </el-row>
                  </div>
                </template>
                <div v-if="form.from == 'import'">
                  <yaml v-model="form.values"></yaml>
                </div>
              </div>
            </el-form>
          </div>
          <div slot="footer" class="dialogFooter" style="padding-top: 25px;">
            <el-button @click="installFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
            <el-button type="primary" @click="updateFormVisible ? handleInstallApp(true) : handleInstallApp()" >
              {{ updateFormVisible ? '升 级' : '安 装' }}
            </el-button>
          </div>
        </div>
      </el-dialog>

      <el-dialog title="应用商店导入" :visible.sync="importStoreFormVisible"
        @close="importStoreFormVisible=false;importStoreAppForm={}" :destroy-on-close="true" :close-on-click-modal="false">
        <div v-loading="dialogLoading">
          <div class="dialogContent" style="">
            <el-form :model="importStoreAppForm" :rules="rules" ref="form" label-position="left" label-width="105px">
              <div v-loading="fetchVersionLoading">
                <el-form-item label="应用名称" prop="" :required="true">
                  <el-select v-model="importStoreAppForm.storeAppId" placeholder="请选择应用" size="small" style="width: 100%;"
                    @change="storeAppChange" filterable>
                    <el-option
                      v-for="item in storeApps"
                      :key="item.id"
                      :label="item.name"
                      :value="item.id">
                    </el-option>
                  </el-select>
                </el-form-item>
                <el-form-item label="应用版本" prop="" :required="true">
                  <el-select v-model="importStoreAppForm.storeAppVersion" placeholder="请选择应用版本" size="small" style="width: 100%;">
                    <el-option
                      v-for="item in storeAppVersions || []"
                      :key="item.id"
                      :label="item.package_version + ' / ' + item.app_version"
                      :value="item.id">
                    </el-option>
                  </el-select>
                </el-form-item>
              </div>
            </el-form>
          </div>
          <div slot="footer" class="dialogFooter" style="padding-top: 20px;">
            <el-button @click="importStoreFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
            <el-button type="primary" @click="handelImportStoreApp" >导 入</el-button>
          </div>
        </div>
      </el-dialog>

      <el-dialog :title="cloneForm.scope == 'project_app' ? '克隆应用' : '发布应用'" :visible.sync="cloneFormVisible"
        @close="cloneFormVisible=false;cloneForm={}" :destroy-on-close="true" :close-on-click-modal="false">
        <div v-loading="dialogLoading">
          <div class="dialogContent" style="">
            <el-form :model="cloneForm" :rules="rules" ref="form" label-position="left" label-width="105px">
              <el-form-item v-if="cloneForm.scope == 'project_app'" label="工作空间" prop="" :required="true">
                <el-select v-model="cloneForm.project_id" placeholder="请选择工作空间" size="small" style="width: 100%;">
                  <el-option
                    v-for="item in projects"
                    :key="item.id"
                    :label="item.name"
                    :value="item.id">
                  </el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="应用名称" prop="" :required="true">
                <el-input v-model="cloneForm.name" placeholder="请输入应用名称" size="small"></el-input>
              </el-form-item>
              <el-form-item label="应用版本" prop="" autofocus>
                <span>{{ cloneForm.package_version }}</span>
              </el-form-item>
            </el-form>
          </div>
          <div slot="footer" class="dialogFooter" style="padding-top: 10px;">
            <el-button @click="cloneFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
            <el-button type="primary" @click="handleDuplicateApp" >{{ cloneForm.scope == 'project_app' ? '克 隆' : '发 布' }}</el-button>
          </div>
        </div>
      </el-dialog>

      <el-dialog title="导入自定义应用" :visible.sync="importCustomVisible" top="5vh" width="70%"
        @close="closeImportCustomDialog" :destroy-on-close="true" :close-on-click-modal="false" v-loading="dialogLoading">
        <div class="dialogContent importCustomApp" style="">
          <el-form :model="importCustomForm" abel-position="left" label-width="105px">
            <el-form-item label="charts包" prop="" required>
              <el-upload
                class="appStoreUpload"
                drag
                ref="appUpload"
                :limit="1"
                :data="importCustomForm"
                :on-success="fileResolve"
                :on-remove="fileRemove"
                action="/api/v1/appstore/resolve">
                <i class="el-icon-upload"></i>
                <div class="el-upload__text">将charts包文件拖到此处，或<em>点击上传</em></div>
              </el-upload>
              <span style="line-height: 20px;" v-if="resolveErrMsg">{{ resolveErrMsg }}</span>
            </el-form-item>
            
            <el-form-item v-if="importCustomForm.name" label="应用名称" prop="" required style="margin-top: 0px;">
              <el-input v-model="importCustomForm.name" autocomplete="off" placeholder="请输入应用名称" size="small"></el-input>
            </el-form-item>
            <el-form-item v-if="importCustomForm.name" label="chart版本" prop="" required style="margin-top: 0px;">
              {{ importCustomForm.package_version }}
            </el-form-item>
            <el-form-item v-if="importCustomForm.name" label="app版本" prop="" required style="margin-top: 0px;">
              {{ importCustomForm.app_version }}
            </el-form-item>
            <el-form-item v-if="importCustomForm.name" label="应用类型" prop="secret_type" style="margin-top: 0px;" required>
              <el-radio-group v-model="importCustomForm.type" name="middleware" size="small">
                <el-radio-button label="ordinary_app">普通应用</el-radio-button>
                <el-radio-button label="middleware">中间件</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item v-if="importCustomForm.name" label="应用描述" prop="" required style="margin-top: 0px;">
              <el-input type="textarea" v-model="importCustomForm.description" autocomplete="off" placeholder="请输入应用描述" size="small"></el-input>
            </el-form-item>
            <el-form-item v-if="importCustomForm.name" label="版本说明" prop="" required style="margin-top: 0px;">
              <el-input type="textarea" v-model="importCustomForm.version_description" autocomplete="off" placeholder="请输入应用版本说明" size="small"></el-input>
            </el-form-item>
            
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter" style="margin-top: 0px;">
          <el-button @click="importCustomVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="handleImportCustomApp" >导 入</el-button>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar, Yaml } from "@/views/components";
import { listProjects } from "@/api/project/project";
import { listApps, listAppStatus, listAppVersions, installApp, destroyApp, deleteApp, importStoreApp, duplicateApp, importCustomApp } from "@/api/project/apps";
import { Message } from "element-ui";
import { listStoreApps } from '@/api/project/appStore'
import yaml from 'js-yaml'

export default {
  name: "projectApps",
  components: {
    Clusterbar,
    Yaml,
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
      dialogLoading: false,
      refreshStatusTimer: undefined,
      form: {
        id: "",
        name: "",
        app_version_id: "",
        values_dict: {},
        values: '',
        from: ''
      },
      rules: {
        // name: [{ required: true, message: '请输入空间名称', trigger: 'blur' },],
      },
      statusNameMap: {
        "UnInstall": "未安装",
        "NotReady": "未就绪",
        "RunningFault": "运行故障",
        "Running": "运行中"
      },
      statusColorMap: {
        "UnInstall": "",
        "NotReady": "#E6A23C",
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
      importStoreFormVisible: false,
      importStoreAppForm: {
        storeAppId: '',
        storeAppVersion: '',
      },
      storeApps: [],
      cloneFormVisible: false,
      cloneForm: {

      },
      projects: [],
      appStatusSSE: undefined,
      importCustomVisible: false,
      importCustomForm: {

      },
      resolveErrMsg: "",
    };
  },
  created() {
    this.fetchApps();
  },
  beforeDestroy() {
    if(this.appStatusSSE) this.appStatusSSE.disconnect()
    if(this.refreshStatusTimer) {
      clearTimeout(this.refreshStatusTimer)
    }
  },
  computed: {
    projectId() {
      return this.$route.params.workspaceId
    },
    storeAppVersions() {
      if(!this.importStoreAppForm.storeAppId) return []
      if(this.storeApps.length == 0) return []
      for(let a of this.storeApps) {
        if(a.id == this.importStoreAppForm.storeAppId && a.versions) return a.versions
      }
      return []
    },
  },
  methods: {
    nameClick: function(id) {
      this.$router.push({name: 'workspaceAppDetail', params: {'appId': id}})
    },
    fetchApps() {
      this.loading = true
      listApps({scope_id: this.projectId, scope: "project_app"}).then((resp) => {
        let originApps = resp.data ? resp.data : []
        this.$set(this, 'originApps', originApps)
        this.loading = false
        // this.getAppStatus()
        if(!this.appStatusSSE) this.fetchAppStatusSSE()
      }).catch((err) => {
        this.loading = false
      })
    },
    handleInstallApp(upgrade) {
      if(!this.form.id) {
        Message.error("获取安装应用失败，请刷新重试");
        return
      }
      if(!this.form.app_version_id) {
        Message.error("请选择要安装的应用版本");
        return
      }
      let values = this.form.values
      if(this.form.from == 'space') {
        let values_dict = JSON.parse(JSON.stringify(this.form.values_dict))
        for(let wk in values_dict.workloads || {}) {
          for(let ck in values_dict.workloads[wk].containers || {}) {
            let image = values_dict.workloads[wk].containers[ck].image
            let tag = values_dict.workloads[wk].containers[ck].tag
            if(!image) {
              Message.error("请输入应用容器镜像")
              return
            }
            if(tag) image += ":"+tag
            values_dict.workloads[wk].containers[ck].image = image
            if(tag != undefined) {
              delete values_dict.workloads[wk].containers[ck].tag
            }
          }
        }
        values = yaml.dump(values_dict)
      }
      let data = {
        project_app_id: this.form.id, 
        app_version_id: this.form.app_version_id, 
        values: values,
        upgrade: upgrade ? true : false
      }
      this.dialogLoading = true
      installApp(data).then(() => {
        this.dialogLoading = false
        this.installFormVisible = false;
        Message.success("安装应用成功")
        this.fetchApps()
      }).catch((err) => {
        this.dialogLoading = false
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
    handleDeleteApp(id, name) {
      if(!id) {
        Message.error("获取应用id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${name}」此应用以及所有版本?`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          this.loading = true
          deleteApp(id).then(() => {
            this.loading = false
            Message.success("删除应用成功")
            this.fetchApps()
          }).catch((err) => {
            this.loading = false
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
        values_dict: {},
        values: '',
        from: app.app_version.from,
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
          let values_dict = yaml.load(v.values)
          this.form.values = v.values
          this.form.from = v.from
          if(v.from == 'space') {
            for(let wk in values_dict.workloads || {}) {
              for(let ck in values_dict.workloads[wk].containers || {}) {
                let imageTag = values_dict.workloads[wk].containers[ck].image
                if(imageTag) {
                  let s = imageTag.split(":")
                  values_dict.workloads[wk].containers[ck].image = s[0]
                  if(s.length > 1) {
                    values_dict.workloads[wk].containers[ck].tag = s[1]
                  }
                }
              }
            }
          }
          this.form.values_dict = values_dict
          return
        }
      }
    },
    closeFormDialog() {
      this.updateFormVisible = false; 
      this.installFormVisible = false;
    },
    openCreateApp() {
      this.$router.push({name: 'workspaceCreateApp'})
    },
    openEditApp(id) {
      this.$router.push({name: 'workspaceEditApp', params: {appVersionId: id}})
    },
    fetchAppStatusSSE() {
      let url = `/api/v1/project/apps/status_sse?scope=project_app&scope_id=${this.projectId}`
      this.appStatusSSE = this.$sse.create({
        url: url,
        includeCredentials: false,
        format: 'plain'
      });
      this.appStatusSSE.on("message", (res) => {
        // console.log(res)
        if(res && res != "\n") {
          let data = JSON.parse(res)
          // console.log(data)
          if(data.code == 'Success') {
            let mapStatus = data.data
            // console.log(mapStatus)
            for(let i in this.originApps) {
              let app = this.originApps[i]
              if(app.name in mapStatus) {
                this.$set(app, 'status', mapStatus[app.name].status)
                // app.status = mapStatus[app.name].status
                this.$set(this.originApps, i, app)
              }
            } 
          }
        }
      })
      this.appStatusSSE.connect().then(() => {
        console.log('[info] connected', 'system')
      }).catch(() => {
        console.log('[error] failed to connect', 'system')
      })
      this.appStatusSSE.on('error', () => { // eslint-disable-line
        console.log('[error] disconnected, automatically re-attempting connection', 'system')
      })
    },
    getAppStatus() {
      listAppStatus({scope_id: this.projectId, scope: "project_app"}).then((resp) => {
        let mapStatus = resp.data
        console.log(mapStatus)
        for(let i in this.originApps) {
          let app = this.originApps[i]
          if(app.name in mapStatus) {
            this.$set(app, 'status', mapStatus[app.name].status)
            // app.status = mapStatus[app.name].status
            this.$set(this.originApps, i, app)
            console.log(this.originApps)
          }
        }
        this.refreshAppStatus()
      }).catch(() => {
        this.refreshAppStatus()
      })
    },
    refreshAppStatus() {
      let that = this
      if(this.refreshStatusTimer) {
        clearTimeout(this.refreshStatusTimer)
      }
      this.refreshStatusTimer = setTimeout(function () {
          that.getAppStatus()
      }, 10000);
    },
    listStoreApps: function() {
      this.dialogLoading = true
      listStoreApps({with_versions: true}).then(response => {
        this.storeApps = []
        for(let a of response.data || []) {
          if(a.type != 'component') this.storeApps.push(a)
        }
        this.storeApps.sort((a, b) => {return a.name > b.name ? 1 : -1})
        this.dialogLoading = false
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    openImportStoreAppDialog() {
      this.importStoreFormVisible = true
      if(this.storeApps.length == 0) {
        this.listStoreApps()
      }
    },
    storeAppChange(val) {
      if(this.storeAppVersions.length > 0) {
        this.importStoreAppForm.storeAppVersion = this.storeAppVersions[0].id
      } else {
        this.importStoreAppForm.storeAppVersion = ''
      }
    },
    handelImportStoreApp() {
      var data = {
        scope: "project_app",
        scope_id: parseInt(this.projectId),
        store_app_id: this.importStoreAppForm.storeAppId,
        app_version_id: this.importStoreAppForm.storeAppVersion
      }
      if(!data.store_app_id) {
        Message.error("请选择应用")
        return
      }
      if(!data.app_version_id) {
        Message.error("请选择应用")
        return
      }
      this.dialogLoading = true
      importStoreApp(data).then(response => {
        this.dialogLoading = false
        Message.success("导入应用成功")
        this.importStoreFormVisible = false
        this.fetchApps()
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    fetchProjects() {
      this.dialogLoading = true
      listProjects().then(response => {
        this.projects = response.data
        this.projects.sort((a, b) => {return a.name > b.name ? 1 : -1})
        this.dialogLoading = false
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    openCloneAppDiloag(app, scope) {
      this.cloneForm = {
        app_id: app.id,
        name: app.name,
        package_version: app.app_version.package_version,
        version_id: app.app_version_id,
        scope: scope
      }
      if(scope == 'project_app' && this.projects.length == 0) {
        this.fetchProjects()
      }
      this.cloneFormVisible = true
    },
    openAppVersion(id) {
      this.$router.push({name: 'workspaceAppVersion', params: {'appId': id}})
    },
    handleDuplicateApp() {
      if(!this.cloneForm.scope) {
        Message.error("克隆参数scope错误，请刷新重试")
        return
      }
      var data = {
        scope: this.cloneForm.scope
      }
      if(this.cloneForm.scope == 'project_app') {
        if(!this.cloneForm.project_id) {
          Message.error("克隆参数project_id错误，请刷新重试")
          return
        }
        data['scope_id'] = this.cloneForm.project_id
      }
      if(!this.cloneForm.app_id) {
        Message.error("克隆参数app_id错误，请刷新重试")
        return
      }
      data['app_id'] = this.cloneForm.app_id
      if(!this.cloneForm.version_id) {
        Message.error("克隆参数version_id错误，请刷新重试")
        return
      }
      data["version_id"] = this.cloneForm.version_id
      if(!this.cloneForm.name) {
        Message.error("克隆参数name错误，请刷新重试")
        return
      }
      data["name"] = this.cloneForm.name
      
      this.dialogLoading = true
      duplicateApp(data).then(response => {
        if(this.cloneForm.scope=='project_app') {
          Message.success("克隆应用到工作空间成功")
        } else {
          Message.success("发布应用到应用商店成功")
        }
        this.dialogLoading = false
        this.cloneFormVisible = false
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    openImportCustomDialog() {
      this.importCustomForm = {
        icon: ''
      }
      this.importCustomVisible = true 
    },
    closeImportCustomDialog() {
      // this.form = {icon: "", type: "middleware"}
      this.resolveErrMsg = ''
    },
    fileResolve(response, file, flist) {
      this.resolveErrMsg = ''
      if(response.code != "Success") {
        file.status = 'error'
        this.resolveErrMsg = response.msg
        Message.error(response.msg)
        this.$refs.appUpload.clearFiles()
      } else {
        let charts = response.data
        this.importCustomForm = {
          name: charts.package_name,
          package_version: charts.package_version,
          app_version: charts.app_version,
          description: charts.description,
          version_description: '',
          type: 'ordinary_app'
        }
      }
    },
    fileRemove(file) {
      this.importCustomForm = {
      }
    },
    handleImportCustomApp() {
      var appUpload = this.$refs.appUpload
      if(appUpload.uploadFiles.length == 0) {
        Message.error("请上传charts包")
        return
      }
      if(!this.importCustomForm.name) {
        Message.error('请输入应用名称')
        return
      }
      if(!this.importCustomForm.description) {
        Message.error("请输入应用描述")
        return
      }
      if(!this.importCustomForm.version_description) {
        Message.error("请输入应用版本说明")
        return
      }
      if(!this.importCustomForm.type) {
        Message.error('请选择应用类型')
        return
      }
      var data = new FormData()
      data.append("scope", "project_app")
      data.append("scope_id", this.projectId)
      data.append("file", appUpload.uploadFiles[0].raw)
      data.append('name', this.importCustomForm.name)
      data.append('package_version', this.importCustomForm.package_version)
      data.append('app_version', this.importCustomForm.app_version)
      data.append('description', this.importCustomForm.description)
      data.append('version_description', this.importCustomForm.version_description)
      data.append('type', this.importCustomForm.type)
      this.dialogLoading = true
      importCustomApp(data).then(response => {
        this.dialogLoading = false
        Message.success("导入自定义应用成功")
        this.importCustomVisible = false
        this.fetchApps()
      }).catch(() => {
        this.dialogLoading = false
      })
    },
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

</style>
<style lang="scss">
.importCustomApp {
  .appStoreUpload {
    .el-upload-dragger {
      height: 100px;

      .el-icon-upload {
        margin: 10px 0px 0px;
      }
    }
    .el-upload-list__item:first-child {
      margin-top: 0px;
    }
    .el-upload__text {
      line-height: 20px;
      margin-bottom: 10px;
      margin-top: -7px;
    }
  }
}
</style>
