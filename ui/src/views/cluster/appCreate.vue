<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-row :gutter="28" v-for="row of appRow" :key="row" style="margin-top: 15px;">
        <el-col :span="8" v-for="col of (row==appRow ? appCol : 3)" :key="row*3+col">
          
          <a @click="installClick(apps[index(row, col)])">
          <el-card shadow="hover" style="height: 160px;">
            <el-row>
              <el-col :span="8">
                <div class="img-wrapper">
                  <el-image :src="apps[index(row, col)].icon">
                    <div slot="error" class="image-slot">
                      <span style="font-size: 32px; padding-left: 15px;">{{ apps[index(row, col)].name[0].toUpperCase() }}</span>
                    </div>
                  </el-image>
                </div>
              </el-col>
              <el-col :span="16">
                <div style="font-size: 20px;">
                  {{ apps[index(row, col)].name }}
                </div>
                <div style="margin-top: 10px;">
                  <span style="font-size: 13px; color:#636a6e!important">Chart版本: </span>
                  <span style="font-size: 14px;">{{ apps[index(row, col)].chart_version }}</span>
                </div>
                <div style="margin-top: 3px;">
                  <span style="font-size: 13px; color:#636a6e!important">App版本: </span>
                  <span style="font-size: 14px;">{{ apps[index(row, col)].app_version }}</span>
                </div>

              </el-col>
            </el-row>

            <el-row style="margin-top: 14px;">
              <el-col :span="24">
                <span style="font-size: 14px; color:#636a6e!important; margin-top: 5px;">
                  {{ apps[index(row, col)].description }}
                </span>
              </el-col>
            </el-row>
          </el-card>
          </a>
        </el-col>
      </el-row>
    </div>
    <el-dialog :title="'安装' + installApp.name" :visible.sync="installDialog" :close-on-click-modal="false" width="60%" top="45px"
      @close="installApp = {}; installDialog = false; yamlValue = '';">
      <div v-loading="installLoading">
        <el-form ref="form" :inline="true" :model="installData" label-width="80px" size="small">
          <el-form-item label="发布名称">
            <el-input v-model="installData.name"></el-input>
          </el-form-item>
          <el-form-item label="命名空间">
            <el-select v-model="installData.namespace" placeholder="">
              <el-option :label="n.name" :value="n.name" v-for="n in namespaces" :key="n.name"></el-option>
            </el-select>
          </el-form-item>
        </el-form>
        <yaml v-if="installDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
        <!-- <span slot="footer" class="dialog-footer"> -->
          <div style="padding: 10px 0px 20px; text-align: right;">
          <el-button plain @click="installDialog = false" size="small">取 消</el-button>
          <el-button plain @click="install()" size="small">确 定</el-button>
          </div>
        <!-- </span> -->
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listApps, getApp, createApp } from '@/api/app'
import { listNamespace } from '@/api/namespace'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'
import { dateFormat } from '@/utils/utils'
let yaml = require('js-yaml')

export default {
  name: 'Application',
  components: {
    Clusterbar,
    Yaml
  },
  data() {
      return {
        installDialog: false,
        yamlName: "",
        yamlValue: "",
        yamlLoading: true,
        cellStyle: {border: 0},
        titleName: ["Applications"],
        maxHeight: window.innerHeight - 150,
        loading: true,
        originApps: [],
        installApp: {},
        installData: {
          'name': '',
          'namespace': 'default',
        },
        search_ns: [],
        search_name: '',
        namespaces: [],
        delFunc: undefined,
        installLoading: false,
      }
  },
  created() {
    this.fetchData();
    this.fetchNamespace();
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
    apps: function() {
      let dlist = []
      for (let p of this.originApps) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
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
    dateFormat,
    index: function(row, col) {
      return (row - 1) * 3 + col - 1
    },
    fetchData: function() {
      this.loading = true
      this.originApps = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listApps().then(response => {
          this.loading = false
          this.originApps = response.data || []
          this.originApps.sort((a, b) => {return a.name > b.name ? 1 : -1})
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.loading = false
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
    installClick: function(app) {
      this.installApp = app;
      this.installDialog = true;
      this.installData.name = app.name;
      this.getAppValue(app.name, app.chart_version)
    },
    getAppValue: function(name, chart_version) {
      if (!name) {
        Message.error("获取应用名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getApp(name, chart_version).then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data.values
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    install: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( !this.installData.name ){
        Message.error("请输入发布名称")
        return
      }
      if ( !this.installData.namespace ){
        Message.error("请选择命名空间")
        return
      }
      try {
        var values = yaml.load(this.yamlValue, 'utf8');
        console.log(values)
      } catch (e) {
        console.log(e);
        Message.error("解析values失败: " + e)
        return
      }
      let params = {
        name: this.installApp.name,
        namespace: this.installData.namespace,
        chart_version: this.installApp.chart_version,
        release_name: this.installData.name,
        values: values,
      }
      this.installLoading = true;
      createApp(cluster, params).then(() => {
        Message.success("安装成功")
        this.installLoading = false
      }).catch((e) => {
        console.log(e)
        this.installLoading = false
      })
    },
    createFunc() {
      
    },
    fetchNamespace: function() {
      this.namespaces = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listNamespace(cluster).then(response => {
          this.namespaces = response.data
          this.namespaces.sort((a, b) => {return a.name > b.name ? 1 : -1})
        }).catch((err) => {
          console.log(err)
        })
      } else {
        Message.error("获取集群异常，请刷新重试")
      }
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

.img-wrapper {
  position: relative;
  width: 70px;
  height: 70px;
  background-color: #fff;
  border: 2px solid rgba(65,117,152,0.1);
  box-shadow: 0 0 5px 0 rgba(65,117,152,0.2);
  border-radius: 50%!important;
  padding: .5rem!important;
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
