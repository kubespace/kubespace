<template>
  <div v-loading="loading">
    <div class="dashboard-container detail-dashboard project-overview-baseinfo" ref="tableCot">
      <div style="padding: 5px 0px 0px;">
        <div style="margin-bottom: 10px;">基本信息</div>
        <el-form label-position="left" inline class="pod-item" label-width="80px" 
          style="margin: 3px 0px 10px 0px; border: 1px solid #EBEEF5; box-shadow: none; padding: 5px 20px;">
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
      </div>
      <div style="padding: 15px 0px 0px;">
        <div style="margin-bottom: 10px;">空间资源</div>
        <div class="pod-item" style="margin: 3px 0px 10px 0px; border: 1px solid #EBEEF5; box-shadow: none; padding: 5px 20px; font-size: 14px;">
          <div class="border-class">
            <el-row :gutter="5">
              <el-col :span="3">
                <el-card shadow="never" style="height: 140px">
                  <div style="text-align: center; height: 20px;">应用</div>
                  <div style="text-align: center; padding-top: 39px; font-size: 20px;">{{ originApps.length }}</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="never" style="height: 64px; color: #409EFF; margin-top: 3px;">
                  <div style="text-align: center; font-size: 13px">未安装</div>
                  <div style="text-align: center; padding-top: 4px;">{{ uninstallCnt }}</div>
                </el-card>
                <el-card shadow="never" style="height: 64px; margin-top: 6px; color: #E6A23C;">
                  <div style="text-align: center; font-size: 13px;">未就绪</div>
                  <div style="text-align: center; padding-top: 4px;">{{ notReadyCnt }}</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="never" style="height: 64px; color: #67C23A; margin-top: 3px;">
                  <div style="text-align: center; font-size: 13px">运行中</div>
                  <div style="text-align: center; padding-top: 4px;">{{ runningCnt }}</div>
                </el-card>
                <el-card shadow="never" style="height: 64px; margin-top: 6px; color: #F56C6C">
                  <div style="text-align: center; font-size: 13px;">运行故障</div>
                  <div style="text-align: center; padding-top: 4px;">{{ runningFaultCnt }}</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="never" style="height: 140px">
                  <div style="text-align: center; height: 20px; font-size: 13px;">ConfigMap</div>
                  <div style="text-align: center; padding-top: 39px; font-size: 20px;">{{ resource.config_map_num }}</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="never" style="height: 140px">
                  <div style="text-align: center; height: 20px; font-size: 13px;">Secret</div>
                  <div style="text-align: center; padding-top: 39px; font-size: 20px;">{{ resource.secret_num }}</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="never" style="height: 140px">
                  <div style="text-align: center; height: 20px; font-size: 13px;">Service</div>
                  <div style="text-align: center; padding-top: 39px; font-size: 20px;">{{ resource.service_num }}</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="never" style="height: 140px">
                  <div style="text-align: center; height: 20px; font-size: 13px;">Ingress</div>
                  <div style="text-align: center; padding-top: 39px; font-size: 20px;">{{ resource.ingress_num }}</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="never" style="height: 140px">
                  <div style="text-align: center; height: 20px; font-size: 13px;">PVC</div>
                  <div style="text-align: center; padding-top: 39px; font-size: 20px;">{{ resource.pvc_num }}</div>
                </el-card>
              </el-col>
            </el-row>
          </div>
        </div>
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
      maxHeight: window.innerHeight - 150,
      loading: true,
      project: {},
      originApps: []
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
        let heightStyle = window.innerHeight - 150
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
        // this.loading = false
      }).catch((err) => {
        console.log(err)
        // this.loading = false
      })
    },
    fetchApps() {
      this.loading = true
      listApps({scope_id: this.projectId, scope: "project_app"}).then((resp) => {
        let originApps = resp.data ? resp.data : []
        this.originApps = originApps
        this.loading = false
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
