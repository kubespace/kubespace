<template>
    <div class="dashboard-container detail-dashboard" ref="tableCot" v-loading="loading">

      <el-card shadow="never" style="border-radius: 10px; border: 0px;">
        <div slot="header" style="">基本信息</div>
        <el-form label-position="left" inline class="pod-item" label-width="80px" 
          style="margin: -10px 0px; box-shadow: none; padding: 0px;">
          <el-form-item label="空间名称">
            <span>{{ project.name }}</span>
          </el-form-item>
          <el-form-item label="绑定集群">
            <span>{{ project.cluster ? project.cluster.name1 : '' }}</span>
          </el-form-item>
          <el-form-item label="创建时间">
            <span>{{ $dateFormat(project.create_time) }}</span>
          </el-form-item>
          <el-form-item label="负责人">
            <span>{{ project.owner }}</span>
          </el-form-item>
          <el-form-item label="命名空间">
            <span>{{ project.namespace }}</span>
          </el-form-item>
        </el-form>
      </el-card>

      <div style="background-color: #fff; padding: 20px; margin: 15px 0px; border-radius: 10px;">
        <div style="margin: 0px 0px 15px 0px;">空间资源</div>
        
          <div class="border-class">
            <el-row :gutter="5">
              <el-col :span="3">
                <el-card shadow="hover" class="cluster-card">
                  <div class="cluster-content">{{ originApps.length }}</div>
                  <div class="cluster-title">应用</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="never" style="height: 44px; color: #409EFF; margin-top: 3px;">
                  <div class="cluster-content app-content">{{ uninstallCnt }}</div>
                  <div class="cluster-title app-title">未安装</div>
                </el-card>
                <el-card shadow="never" style="height: 44px; margin-top: 6px; color: #E6A23C;">
                  <div class="cluster-content app-content">{{ notReadyCnt }}</div>
                  <div class="cluster-title app-title">未就绪</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="never" style="height: 44px; color: #67C23A; margin-top: 3px;">
                  <div class="cluster-content app-content">{{ runningCnt }}</div>
                  <div class="cluster-title app-title">运行中</div>
                </el-card>
                <el-card shadow="never" style="height: 44px; margin-top: 6px; color: #F56C6C">
                  <div class="cluster-content app-content">{{ runningFaultCnt }}</div>
                  <div class="cluster-title app-title">运行故障</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="hover" class="cluster-card">
                  <div class="cluster-content">{{ resource.config_map_num }}</div>
                  <div class="cluster-title">ConfigMap</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="hover" class="cluster-card">
                  <div class="cluster-content">{{ resource.secret_num }}</div>
                  <div class="cluster-title">Secret</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="hover" class="cluster-card">
                  <div class="cluster-content">{{ resource.service_num }}</div>
                  <div class="cluster-title">Service</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="hover" class="cluster-card">
                  <div class="cluster-content">{{ resource.ingress_num }}</div>
                  <div class="cluster-title">Ingress</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="hover" class="cluster-card">
                  <div class="cluster-content">{{ resource.pvc_num }}</div>
                  <div class="cluster-title">PVC</div>
                </el-card>
              </el-col>
            </el-row>
          </div>
      </div>
    </div>
</template>

<script>
import { getProject } from '@/api/project/project'
import { listApps } from '@/api/project/apps'
import { Message } from 'element-ui'

export default {
  name: 'ProjectOverview',
  components: {
  },
  data() {
    return {
      titleName: ["项目空间"],
      search_name: '',
      cellStyle: {border: 0},
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      project: {},
      originApps: [],
      projectLoadingOk: false,
      appLoadingOk: false,
    }
  },
  created() {
    this.fetchProject()
    this.fetchApps()
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - this.$contentHeight
        console.log(heightStyle)
        that.maxHeight = heightStyle
      })()
    }
  },
  computed: {
    projectId() {
      return this.$route.params.workspaceId
    },
    runningCnt() {
      return this.appStatusCnt('Running')
    },
    uninstallCnt() {
      return this.appStatusCnt('UnInstall')
    },
    notReadyCnt() {
      return this.appStatusCnt('NotReady')
    },
    runningFaultCnt() {
      return this.appStatusCnt('RunningFault')
    },
    resource() {
      if(this.project.resource) return this.project.resource
      return {}
    }
  },
  methods: {
    fetchProject() {
      this.loading = true
      getProject(this.projectId,).then((resp) => {
        this.project = resp.data ? resp.data : {}
        this.projectLoadingOk = true
        if(this.projectLoadingOk && this.appLoadingOk) {
          this.loading = false
        }
      }).catch((err) => {
        this.loading = false
      })
    },
    fetchApps() {
      this.loading = true
      listApps({scope_id: this.projectId, scope: "project"}).then((resp) => {
        let originApps = resp.data ? resp.data : []
        this.originApps = originApps
        this.appLoadingOk = true
        if(this.projectLoadingOk && this.appLoadingOk) {
          this.loading = false
        }
      }).catch((err) => {
        this.loading = false
      })
    },
    appStatusCnt(status) {
      let c = 0
      for(let a of this.originApps) {
        if(a.status == status) c++
      }
      return c
    }
  },
}
</script>

<style lang="scss" scoped>
.project-overview-baseinfo {
  padding: 20px 0px;
}

.cluster-card {
  height: 100px;
}

.cluster-content {
  text-align: center; 
  font-size: 18px;
  margin-top: 10px;
}

.cluster-title {
  text-align: center; 
  margin-top: 10px; 
  font-size: 13px; 
  opacity: 0.6; 
  height: 20px;
}

.app-content {
  font-size: 15px;
  margin-top: -16px;
}

.app-title {
  font-size: 12px;
  margin-top: 3px;
}
.app-two-content {
  margin-top: -3px;
}
</style>

<style lang="scss">
.project-overview-baseinfo{
  .el-form-item__label{
    line-height: 28px;
  }
  .el-form-item__content {
    line-height: 28px;
  }
}
</style>
