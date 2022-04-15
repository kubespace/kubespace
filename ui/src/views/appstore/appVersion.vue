<template>
  <div>
    <clusterbar :titleName="titleName" :createFunc="openImportAppVersionDialog" createDisplay="导入新版本"
      :titleLink="['appstoreIndex']"/>
    <div class="dashboard-container" v-loading="loading">
      <el-card shadow="" class="appVersionCard">
        <el-row>
          <el-col :span="4" style="width: 80px;">
            <div class="app-version-img-wrapper" style="margin-left: 10px;">
              <el-image :src="originApp.icon ? 'data:image/png;base64,'+ originApp.icon : ''" fit="contain">
                <div slot="error" class="image-slot">
                  <!-- <span style="font-size: 32px;">{{ apps[index(row, col)].name[0].toUpperCase() }}</span> -->
                  <span style="font-size: 32px;"><i class="el-icon-s-grid"></i></span>
                </div>
              </el-image>
            </div>
          </el-col>
          <el-col :span="18">
            <div style="font-size: 18px; margin-top: 4px;">
              {{ originApp.name }}
            </div>
            <div style="margin-top: 2px;">
              <span style="font-size: 13px; color:#636a6e!important">应用类型: </span>
              <span style="font-size: 14px;">{{ typeMap[originApp.type] }}</span>
            </div>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="24" class="description">
            <span>
              {{ originApp.description }}
            </span>
          </el-col>
        </el-row>
      </el-card>

      <el-card class="appVersionTable">
        <el-table
          ref="multipleTable"
          :data="originApp.versions ? originApp.versions : []"
          class="table-fix"
          :cell-style="cellStyle"
          v-loading="loading"
          :default-sort = "{prop: 'create_time', order: 'descending'}"
          tooltip-effect="dark"
          style="width: 100%"
        >
          <el-table-column prop="name" label="chart版本" show-overflow-tooltip min-width="15">
            <template slot-scope="scope">
              <span >
                {{ scope.row.package_version }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="app_version" label="app版本" show-overflow-tooltip min-width="15">
          </el-table-column>
          <el-table-column prop="description" label="版本说明" show-overflow-tooltip min-width="15">
          </el-table-column>
          <el-table-column prop="owner" label="创建人" show-overflow-tooltip min-width="10">
            <template slot-scope="scope">
              {{ scope.row.create_user }}
            </template>
          </el-table-column>
          <el-table-column prop="update_time" label="创建时间" show-overflow-tooltip min-width="15">
            <template slot-scope="scope">
              {{ $dateFormat(scope.row.create_time) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template slot-scope="scope">
              <div class="tableOperate">
                <el-link :underline="false" type="primary" style="margin-right: 13px; color:#409EFF" :href="'/app/charts/'+scope.row.chart_path">下载</el-link>
                <el-link :disabled="!$editorRole()" :underline="false" type="danger" @click="handleDeleteVersion(scope.row.id, scope.row.package_version)">删除</el-link>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <el-dialog :title="'导入新版本'" :visible.sync="createFormVisible" top="5vh" width="70%"
      @close="closeImportAppDialog" :destroy-on-close="true" :close-on-click-modal="false" v-loading="dialogLoading">
      <div class="dialogContent" style="">
        <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
          <el-form-item v-if="form.name" label="应用名称" prop="" required style="margin-top: -10px;">
            {{ form.name }}
          </el-form-item>
          <el-form-item label="charts包" prop="" required>
            <el-upload
              class="appStoreUpload"
              drag
              ref="appUpload"
              :limit="1"
              :data="form"
              :on-success="fileResolve"
              action="/api/v1/appstore/resolve">
              <i class="el-icon-upload"></i>
              <div class="el-upload__text">将charts包文件拖到此处，或<em>点击上传</em></div>
            </el-upload>
            <span style="line-height: 20px;" v-if="resolveErrMsg">{{ resolveErrMsg }}</span>
          </el-form-item>
          <el-form-item v-if="form.package_version" label="chart版本" prop="" required style="margin-top: -10px;">
            {{ form.package_version }}
          </el-form-item>
          <el-form-item v-if="form.package_version" label="app版本" prop="" required style="margin-top: -10px;">
            {{ form.app_version }}
          </el-form-item>
          <el-form-item v-if="form.package_version" label="版本说明" prop="" required style="margin-top: -5px;">
            <el-input type="textarea" v-model="form.version_description" autocomplete="off" placeholder="请输入应用版本说明" size="small"></el-input>
          </el-form-item>
          
        </el-form>
      </div>
      <div slot="footer" class="dialogFooter" style="margin-top: -10px;">
        <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
        <el-button type="primary" @click="handleImportApp" >导 入</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { getStoreApp, deleteStoreAppVersion, createStoreApp } from '@/api/project/appStore'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'
let yaml = require('js-yaml')

export default {
  name: 'AppStoreVersions',
  components: {
    Clusterbar,
    Yaml
  },
  data() {
    return {
      cellStyle: {border: 0},
      titleName: ["应用商店"],
      maxHeight: window.innerHeight - 135,
      loading: true,
      dialogLoading: false,
      originApp: {},
      typeMap: {
        middleware: "中间件",
        component: "集群组件",
        ordinary_app: "普通应用"
      },
      createFormVisible: false,
      resolveErrMsg: '',
      form: {},
      rules: {},
    }
  },
  created() {
    this.fetchData();
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 135
        // console.log(heightStyle)
        that.maxHeight = heightStyle
      })()
    }
  },
  watch: {
  },
  computed: {
    appId() {
      return this.$route.params.appId
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      getStoreApp(this.appId, {with_versions: true}).then(response => {
        this.loading = false
        this.originApp = response.data
        this.titleName = ["应用商店", this.originApp.name]
      }).catch(() => {
        this.loading = false
      })
    },
    importAppVersion() {

    },
    handleDeleteVersion(id, version) {
      if(!id) {
        Message.error("获取应用版本id参数异常，请刷新重试");
        return
      }
      var msg = `请确认是否删除「${version}」此应用版本?`
      if(this.originApp.versions.length == 1) {
        msg += " 当前为最后一个版本，并同时删除该应用。"
      }

      this.$confirm(msg, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        deleteStoreAppVersion(this.appId, id).then(() => {
          Message.success("删除应用版本成功")
          if(this.originApp.versions.length == 1) {
            this.loading = false
            this.$router.push({name: 'appstoreIndex'})
          } else {
            this.fetchData()
          }
        }).catch((err) => {
          console.log(err)
        });
      }).catch(() => {       
      });
    },
    handleImportApp() {
      var appUpload = this.$refs.appUpload
      if(appUpload.uploadFiles.length == 0) {
        Message.error("请上传charts包")
        return
      }
      if(!this.form.name) {
        Message.error('应用名称为空')
        return
      }
      if(!this.form.version_description) {
        Message.error("请输入应用版本说明")
        return
      }
      var data = new FormData()
      data.append("file", appUpload.uploadFiles[0].raw)
      data.append('name', this.form.name)
      data.append('package_version', this.form.package_version)
      data.append('app_version', this.form.app_version)
      data.append('version_description', this.form.version_description)
      this.dialogLoading = true
      createStoreApp(data).then(response => {
        this.dialogLoading = false
        Message.success("导入应用新版本成功")
        this.createFormVisible = false
        this.fetchData()
      }).catch(() => {
        this.dialogLoading = false
      })
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
        this.form = {
          name: this.originApp.name,
          package_version: charts.package_version,
          app_version: charts.app_version,
          version_description: ''
        }
      }
    },
    openImportAppVersionDialog() {
      this.form = {
        name: this.originApp.name,
      }
      this.createFormVisible = true 
    },
    closeImportAppDialog() {
      this.form = {}
      this.resolveErrMsg = ''
    },
  }
}
</script>

<style lang="scss" scoped>
.dashboard {
  &-container {
    // padding: 10px 30px;
    // margin-right: 15px;
    // padding-right: 15px;
    // overflow: auto;
    margin-top: 20px;
  }
  &-text {
    font-size: 30px;
    line-height: 46px;
  }

  .table-fix {
    height: calc(100% - 100px);
  }
}
.description {
  margin-top: 8px;
  font-size: 14px;
  color: rgb(99, 106, 110);
  margin-left: 15px;
}

</style>

<style lang="scss">
.app-version-img-wrapper {
  .el-image {
    display: table-cell;
    vertical-align: middle;
    text-align: center;
    width: 50px;
    height: 50px;
    border: 2px solid rgba(65,117,152,0.1);
    box-shadow: 0 0 5px 0 rgba(65,117,152,0.2);
    border-radius: 50%!important;
    padding: 8px!important;
  }
}
.appVersionCard {
  .el-card__body {
    padding: 10px;
  }
}
.appVersionTable {
  margin-top: 20px;
  .el-card__body {
    padding: 0px;
  }
}
</style>
