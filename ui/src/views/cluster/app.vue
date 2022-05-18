<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch"
    :createFunc="openImportStoreAppDialog" createDisplay="导入集群组件"/>
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
        <el-table-column prop="name" label="名称" show-overflow-tooltip min-width="15">
          <!-- <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.id)">
              {{ scope.row.name }}
            </span>
          </template> -->
        </el-table-column>
        <el-table-column prop="package_version" label="版本" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ scope.row.app_version.from == 'space' ? scope.row.app_version.package_version : scope.row.app_version.package_version + ' / ' + scope.row.app_version.app_version }}
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" show-overflow-tooltip min-width="15">
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
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 15px;" 
                v-if="scope.row.status=='UnInstall'" @click="openInstallFormDialog(scope.row)">安装</el-link>
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 15px;" 
                v-if="scope.row.status!='UnInstall'" @click="openInstallFormDialog(scope.row, true)">升级</el-link>
              
              <el-dropdown style="font-size: 13px;">
                <span class="el-dropdown-link operator-btn" :style="{color: !$editorRole() ? '#a0cfff' : ''}">
                  更多操作
                </span>
                <el-dropdown-menu slot="dropdown">
                  <el-dropdown-item>
                    <el-link :disabled="!$editorRole()" :underline="false" type="danger" style="font-weight: 400"
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

    <el-dialog title="应用商店导入" :visible.sync="importStoreFormVisible"
      @close="importStoreFormVisible=false;importStoreAppForm={}" :destroy-on-close="true" :close-on-click-modal="false">
      <div v-loading="dialogLoading">
        <div class="dialogContent" style="">
          <el-form :model="importStoreAppForm" :rules="rules" ref="form" label-position="left" label-width="105px">
            <div v-loading="fetchVersionLoading">
              <el-form-item label="命名空间" prop="" :required="true">
                <el-select v-model="importStoreAppForm.namespace" placeholder="请选择要安装的命名空间" size="small" style="width: 100%;">
                  <el-option
                    v-for="item in namespaces"
                    :key="item.name"
                    :label="item.name"
                    :value="item.name">
                  </el-option>
                </el-select>
              </el-form-item>
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
        <div slot="footer" class="dialogFooter" style="padding-top: 10px;">
          <el-button @click="importStoreFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="handelImportStoreApp" >导 入</el-button>
        </div>
      </div>
    </el-dialog>

    <el-dialog :title="updateFormVisible ? '升级应用' : '安装应用'" :visible.sync="installFormVisible"
      @close="closeFormDialog" :destroy-on-close="true" :close-on-click-modal="false" top="2vh" width="70%">
      <div v-loading="dialogLoading">
        <div class="dialogContent" style="margin-top: -10px;">
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
                    :label="item.package_version + ' / ' + item.app_version"
                    :value="item.id">
                  </el-option>
                </el-select>
              </el-form-item>
              
              <div>
                <yaml v-model="form.values"></yaml>
              </div>
            </div>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter" style="padding-top: 20px;">
          <el-button @click="installFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="updateFormVisible ? handleInstallApp(true) : handleInstallApp()" >
            {{ updateFormVisible ? '升 级' : '安 装' }}
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
// import { listReleases, deleteRelease, getRelease } from '@/api/app'
import { listApps, listAppStatus, listAppVersions, installApp, destroyApp, deleteApp, importStoreApp } from "@/api/project/apps";
import { Message } from 'element-ui'
import { listStoreApps } from '@/api/project/appStore'
import { listNamespace } from '@/api/namespace'
import { Yaml } from '@/views/components'
import yaml from 'js-yaml'

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
      dialogLoading: false,
      cellStyle: {border: 0},
      titleName: ["Components"],
      maxHeight: window.innerHeight - 150,
      loading: true,
      originApps: [],
      appVersions: [],
      namespaces: [],
      installFormVisible: false,
      updateFormVisible: false,
      search_ns: [],
      search_name: '',
      yamlChange: 1,
      rules: {},
      form: {
        id: "",
        name: "",
        app_version_id: "",
        values_dict: {},
        values: '',
        from: ''
      },
      updateValues: {
        name: "",
        namespace: "",
        config: "",
        values: "",
      },
      importStoreFormVisible: false,
      importStoreAppForm: {
        storeAppId: '',
        storeAppVersion: '',
        namespace: "default",
      },
      fetchVersionLoading: false,
      storeApps: [],
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
        "import_app": "导入应用",
        "component": "集群组件"
      },
    }
  },
  created() {
    this.fetchApps()
    this.fetchNamespace()
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
      for (let p of this.originApps) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        
        dlist.push(p)
      }
      return dlist
    },
    clusterId() {
      return parseInt(this.$store.state.cluster)
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
    fetchApps() {
      this.loading = true
      listApps({scope_id: this.clusterId, scope: "component"}).then((resp) => {
        let originApps = resp.data ? resp.data : []
        this.$set(this, 'originApps', originApps)
        this.loading = false
        // this.getAppStatus()
      }).catch((err) => {
        this.loading = false
      })
    },
    fetchNamespace: function() {
      this.namespaces = []
      if (this.$store.state.cluster) {
        listNamespace(this.$store.state.cluster).then(response => {
          this.namespaces = response.data
          this.namespaces.sort((a, b) => {return a.name > b.name ? 1 : -1})
        }).catch((err) => {
          console.log(err)
        })
      } else {
        Message.error("获取集群异常，请刷新重试")
      }
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
        Message.error("获取销毁集群组件id失败，请刷新重试");
        return
      }
      this.$confirm(`请确认是否销毁「${name}」此集群组件?`, '提示', {
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
      this.$confirm(`请确认是否删除「${name}」此集群组件以及所有版本?`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          deleteApp(id).then(() => {
            Message.success("删除应用成功")
            this.fetchApps()
          }).catch((err) => {
            console.log(err)
          });
        }).catch(() => {       
        });
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
    listStoreApps: function() {
      this.dialogLoading = true
      listStoreApps({with_versions: true}).then(response => {
        this.storeApps = []
        for(let a of response.data || []) {
          if(a.type == 'component') this.storeApps.push(a)
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
        scope: "component",
        scope_id: this.clusterId,
        namespace: this.importStoreAppForm.namespace,
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
.replicaDialog {
  .el-form-item {
    margin-bottom: 10px;
  }
  .el-dialog--center .el-dialog__body {
    padding: 5px 25px;
  }
}
</style>
