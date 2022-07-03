<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="openImportAppDialog" createDisplay="导入应用"/>
    <div class="dashboard-container" :style="{'max-height': maxHeight + 'px'}" :max-height="maxHeight">
      <el-row :gutter="28" v-for="row of appRow" :key="row" style="margin-top: 15px;">
        <el-col :span="8" v-for="col of (row==appRow ? appCol : 3)" :key="row*3+col">
          <a @click="versionClick(apps[index(row, col)])">
            <el-card shadow="" style="height: 160px;">
              <el-row>
                <el-col :span="6">
                  <div class="img-wrapper">
                    <el-image :src="'data:image/png;base64,'+apps[index(row, col)].icon" fit="contain">
                      <div slot="error" class="image-slot">
                        <!-- <span style="font-size: 32px;">{{ apps[index(row, col)].name[0].toUpperCase() }}</span> -->
                        <span style="font-size: 45px;"><i class="el-icon-s-grid"></i></span>
                      </div>
                    </el-image>
                  </div>
                </el-col>
                <el-col :span="16">
                  <div style="font-size: 20px;">
                    {{ apps[index(row, col)].name }}
                  </div>
                  <div style="margin-top: 7px;">
                    <span style="font-size: 13px; color:#636a6e!important">应用类型: </span>
                    <span style="font-size: 14px;">{{ typeMap[apps[index(row, col)].type] }}</span>
                  </div>
                  <div style="margin-top: 3px;">
                    <span style="font-size: 13px; color:#636a6e!important">最新版本: </span>
                    <span style="font-size: 14px;">{{ apps[index(row, col)].latest_package_version }} / {{ apps[index(row, col)].latest_app_version }}</span>
                  </div>

                </el-col>
              </el-row>

              <el-row style="margin-top: 8px;">
                <el-col :span="24">
                  <el-tooltip class="item" effect="light" placement="top-start">
                    <div slot="content" style="width: 300px;">
                      {{ apps[index(row, col)].description }}
                    </div>
                    <span class="description-line">
                      {{ apps[index(row, col)].description }}
                    </span>
                  </el-tooltip>
                </el-col>
              </el-row>
            </el-card>
          </a>
        </el-col>
      </el-row>
    </div>

    <el-dialog title="导入应用" :visible.sync="createFormVisible" top="5vh" width="70%"
      @close="closeImportAppDialog" :destroy-on-close="true" :close-on-click-modal="false" v-loading="dialogLoading">
      <div class="dialogContent" style="">
        <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
          <el-form-item label="charts包" prop="" required>
            <el-upload
              class="appStoreUpload"
              drag
              ref="appUpload"
              :limit="1"
              :data="form"
              :on-success="fileResolve"
              :on-remove="fileRemove"
              action="/api/v1/appstore/resolve">
              <i class="el-icon-upload"></i>
              <div class="el-upload__text">将charts包文件拖到此处，或<em>点击上传</em></div>
            </el-upload>
            <span style="line-height: 20px;" v-if="resolveErrMsg">{{ resolveErrMsg }}</span>
          </el-form-item>
          <el-form-item v-if="form.name" label="应用名称" prop="" style="margin-top: -10px;">
            <div slot="label">
              <div class="app-version-img-wrapper" style="margin-left: 10px;">
                <el-image :src="form.icon" fit="contain">
                  <div slot="error" class="image-slot">
                    <span style="font-size: 32px;"><i class="el-icon-s-grid"></i></span>
                  </div>
                </el-image>
              </div>
            </div>
            <el-upload style="margin-top: 5px;"
              class="upload-demo"
              ref="iconUpload"
              action="/api/v1/appstore/resolve"
              :on-change="handlePictureCardPreview"
              :auto-upload="false"
              :show-file-list="false"
              >
              <el-button size="small" type="primary" style="border-radius: 0px;">上传应用图标</el-button>
            </el-upload>
          </el-form-item>
          <el-form-item v-if="form.name" label="应用名称" prop="" required style="margin-top: 0px;">
            <el-input v-model="form.name" autocomplete="off" placeholder="请输入应用名称" size="small"></el-input>
          </el-form-item>
          <el-form-item v-if="form.name" label="chart版本" prop="" required style="margin-top: 0px;">
            {{ form.package_version }}
          </el-form-item>
          <el-form-item v-if="form.name" label="app版本" prop="" required style="margin-top: 0px;">
            {{ form.app_version }}
          </el-form-item>
          <el-form-item v-if="form.name" label="应用类型" prop="secret_type" style="margin-top: 0px;" required>
            <el-radio-group v-model="form.type" name="middleware" size="small">
              <el-radio-button label="middleware">中间件</el-radio-button>
              <el-radio-button label="component">集群组件</el-radio-button>
              <el-radio-button label="ordinary_app">普通应用</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item v-if="form.name" label="应用描述" prop="" required style="margin-top: 0px;">
            <el-input type="textarea" v-model="form.description" autocomplete="off" placeholder="请输入应用描述" size="small"></el-input>
          </el-form-item>
          <el-form-item v-if="form.name" label="版本说明" prop="" required style="margin-top: 0px;">
            <el-input type="textarea" v-model="form.version_description" autocomplete="off" placeholder="请输入应用版本说明" size="small"></el-input>
          </el-form-item>
          
        </el-form>
      </div>
      <div slot="footer" class="dialogFooter" style="margin-top: 0px;">
        <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
        <el-button type="primary" @click="handleImportApp" >导 入</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listStoreApps, createStoreApp } from '@/api/project/appStore'
import { Message } from 'element-ui'

export default {
  name: 'AppStore',
  components: {
    Clusterbar,
  },
  data() {
      return {
        cellStyle: {border: 1},
        titleName: ["应用商店"],
        maxHeight: window.innerHeight - 135,
        loading: true,
        search_name: '',
        originApps: [],
        dialogLoading: false,
        typeMap: {
          middleware: "中间件",
          component: "集群组件",
          ordinary_app: "普通应用"
        },
        createFormVisible: false,
        resolveErrMsg: "",
        form: {
          icon: ''
        },
        rules: {}
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
    apps: function() {
      let dlist = []
      for (let p of this.originApps) {
        if (this.search_name && !p.name.includes(this.search_name)) continue
        
        dlist.push(p)
      }
      return dlist
    },
    appRow: function() {
      if(this.apps) {
        if(this.apps.length % 3 > 0) return parseInt(this.apps.length / 3) + 1;
        else return parseInt(this.apps.length / 3)
      }
      return 0
    },
    appCol: function() {
      if (this.apps) {
        if(this.apps.length % 3 > 0) return this.apps.length % 3;
        else return 3
      }
      return 0
    }
  },
  methods: {
    index: function(row, col) {
      return (row - 1) * 3 + col - 1
    },
    fetchData: function() {
      this.loading = true
      this.originApps = []
      listStoreApps().then(response => {
        this.loading = false
        let originApps = response.data || []
        originApps.sort((a, b) => {return a.name > b.name ? 1 : -1})
        this.$set(this, 'originApps', originApps)
      }).catch(() => {
        this.loading = false
      })
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    versionClick: function(app) {
      this.$router.push({name: 'appstoreVersions', params: {'appId': app.id}})
    },
    openImportAppDialog() {
      this.form = {
        icon: ''
      }
      this.createFormVisible = true 
    },
    closeImportAppDialog() {
      // this.form = {icon: "", type: "middleware"}
      this.resolveErrMsg = ''
    },
    handleImportApp() {
      var appUpload = this.$refs.appUpload
      if(appUpload.uploadFiles.length == 0) {
        Message.error("请上传charts包")
        return
      }
      if(!this.form.name) {
        Message.error('请输入应用名称')
        return
      }
      if(!this.form.description) {
        Message.error("请输入应用描述")
        return
      }
      if(!this.form.version_description) {
        Message.error("请输入应用版本说明")
        return
      }
      if(!this.form.type) {
        Message.error('请选择应用类型')
        return
      }
      var data = new FormData()
      data.append("file", appUpload.uploadFiles[0].raw)
      data.append('name', this.form.name)
      data.append('package_version', this.form.package_version)
      data.append('app_version', this.form.app_version)
      data.append('description', this.form.description)
      data.append('version_description', this.form.version_description)
      data.append('type', this.form.type)
      if(this.form.iconFile) {
        data.append('icon', this.form.iconFile)
      }
      this.dialogLoading = true
      createStoreApp(data).then(response => {
        this.dialogLoading = false
        Message.success("导入应用成功")
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
          icon: this.form.icon,
          name: charts.package_name,
          package_version: charts.package_version,
          app_version: charts.app_version,
          description: charts.description,
          version_description: '',
          type: 'middleware'
        }
      }
    },
    fileRemove(file) {
      this.form = {
        icon: '',
      }
    },
    handlePictureCardPreview(file) {
      this.$refs.iconUpload.clearFiles()
      if(!file.raw.type || file.raw.type.indexOf('image') < 0) {
        Message.error('请选择图片类型文件')
        return
      }
      var that = this
      var reader = new FileReader()
      reader.readAsDataURL(file.raw);
      reader.onload = function(e) {
        that.$set(that.form, 'icon', this.result)
        that.form.iconFile = file.raw
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.dashboard {
  &-container {
    // padding: 10px 30px;
    margin-right: 15px;
    padding-right: 15px;
    margin-left: 15px;
    padding-left: 15px;
    padding-bottom: 10px;;
    overflow: auto;
  }
  &-text {
    font-size: 30px;
    line-height: 46px;
  }

  .table-fix {
    height: calc(100% - 100px);
  }
}
.description-line {
  font-size: 14px;
  color: rgb(99, 106, 110) !important;
  margin-top: 5px;
  white-space: normal;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
  -webkit-line-clamp: 2;
}

</style>

<style lang="scss">
.img-wrapper {
  .el-image {
    display: table-cell;
    vertical-align: middle;
    text-align: center;
    width: 70px;
    height: 70px;
    border: 2px solid rgba(65,117,152,0.1);
    box-shadow: 0 0 5px 0 rgba(65,117,152,0.2);
    border-radius: 50%!important;
    padding: 6px!important;

    img {
      max-width: calc(100% - 10px);
      max-height: calc(100% - 10px);
    }
  }
}
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
.app-version-img-wrapper {
  .image-slot {
    line-height: 0px;
  }
  .el-image {
    display: table-cell;
    vertical-align: middle;
    text-align: center;
    width: 50px;
    height: 50px;
    border: 2px solid rgba(65,117,152,0.1);
    box-shadow: 0 0 5px 0 rgba(65,117,152,0.2);
    border-radius: 50%!important;
    padding: 5px!important;
    line-height: 0px;

    img {
      max-width: 100%;
      max-height: 100%;
    }
  }
}
</style>
