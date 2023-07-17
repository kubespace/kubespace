<template>
  <div>
    <div class="dashboard-container" v-loading="loading" style="margin-top: 15px;">
      <el-row :gutter="10" style="margin-bottom: 10px;" class="row-class">
        <el-col :span="18">
          <div class="border-class">
            <div style="margin: 5px 0px 10px;">集群状态</div>
            <el-row :gutter="5">
              <el-col :span="4">
                <el-card shadow="hover" style="height: 67px">
                  <div class="cluster-content">{{ cluster_detail.cluster_version }}</div>
                  <div class="cluster-title">Cluster Version</div>
                </el-card>
              </el-col>
              <el-col :span="5">
                <el-card shadow="hover" style="height: 67px;">
                  <el-row>
                    <el-col :span="9" style="border-right: 1px solid #EBEEF5">
                      <div class="cluster-content">{{ cluster_detail.cluster_cpu }}</div>
                      <div class="cluster-title">CPU</div>
                    </el-col>
                    <el-col :span="15">
                      <div class="cluster-content">{{ cluster_detail.cluster_memory }}</div>
                      <div class="cluster-title">Memory</div>
                    </el-col>
                  </el-row>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="hover" style="height: 67px">
                  <div class="cluster-content">{{ cluster_detail.node_num }}</div>
                  <div class="cluster-title">Node</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="hover" style="height: 67px">
                  <div class="cluster-content">{{ cluster_detail.namespace_num }}</div>
                  <div class="cluster-title">Namespace</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="hover" style="height: 67px">
                  <div class="cluster-content">{{ cluster_detail.serviceaccount_num }}</div>
                  <div class="cluster-title">ServiceAccount</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="hover" style="height: 67px">
                  <div class="cluster-content">{{ cluster_detail.event_num }}</div>
                  <div class="cluster-title">Event</div>
                </el-card>
              </el-col>
              <el-col :span="3">
                <el-card shadow="hover" style="height: 67px">
                  <div class="cluster-content">{{ cluster_detail.crd_num }}</div>
                  <div class="cluster-title">CRD</div>
                </el-card>
              </el-col>
            </el-row>
          </div>
        </el-col>

        <el-col :span="6">
          <div class="border-class">
            <div style="margin: 5px 0px 10px;">配置状态</div>
            <el-row :gutter="5">
              <el-col :span="12">
                <el-card shadow="hover" style="height: 67px">
                  <div class="cluster-content">{{ cluster_detail.config_map_num }}</div>
                  <div class="cluster-title">Configmap</div>
                </el-card>
              </el-col>
              <el-col :span="12">
                <el-card shadow="hover" style="height: 67px">
                  <div class="cluster-content">{{ cluster_detail.secret_num }}</div>
                  <div class="cluster-title">Secret</div>
                </el-card>
              </el-col>
            </el-row>
          </div>
        </el-col>
      </el-row>

      <el-row :gutter="10" style="margin-bottom: 10px;" class="row-class">
        <el-col :span="12">
          <div class="border-class">
            <div style="margin: 5px 0px 10px;">应用状态</div>
            <el-row :gutter="5">
              <el-col :span="4">
                <el-card shadow="hover" style="height: 100px">
                  <div class="cluster-content app-content">{{ cluster_detail.pod_num }}</div>
                  <div class="cluster-title app-title">Pod</div>
                </el-card>
              </el-col>
              <el-col :span="4">
                <el-card shadow="never" style="height: 44px; color: #409EFF; margin-top: 3px;">
                  <div class="cluster-content pod-content">{{ cluster_detail.pod_running_num }}</div>
                  <div class="cluster-title pod-title">Running</div>
                </el-card>
                <el-card shadow="never" style="height: 44px; margin-top: 6px; color: #E6A23C;">
                  <div class="cluster-content pod-content">{{ cluster_detail.pod_pending_num }}</div>
                  <div class="cluster-title pod-title">Pending</div>
                </el-card>
              </el-col>
              <el-col :span="4">
                <el-card shadow="never" style="height: 44px; color: #67C23A; margin-top: 3px;">
                  <div class="cluster-content pod-content">{{ cluster_detail.pod_succeeded_num }}</div>
                  <div class="cluster-title pod-title">Succeeded</div>
                </el-card>
                <el-card shadow="never" style="height: 44px; margin-top: 6px; color: #F56C6C">
                  <div class="cluster-content pod-content">{{ cluster_detail.pod_failed_num }}</div>
                  <div class="cluster-title pod-title">Failed</div>
                </el-card>
              </el-col>
              <el-col :span="4">
                <el-card shadow="hover" style="height: 100px">
                  <div class="cluster-content app-content">{{ cluster_detail.deployment_num }}</div>
                  <div class="cluster-title app-title">Deployment</div>
                </el-card>
              </el-col>
              <el-col :span="4">
                <el-card shadow="hover" style="height: 48px;">
                  <div class="cluster-content pod-content app-two-content">{{ cluster_detail.statefulset_num }}</div>
                  <div class="cluster-title pod-title">StatefulSet</div>
                </el-card>
                <el-card shadow="hover" style="height: 48px; margin-top: 4px;">
                  <div class="cluster-content pod-content app-two-content">{{ cluster_detail.daemonset_num }}</div>
                  <div class="cluster-title pod-title">DaemonSet</div>
                </el-card>
              </el-col>
              <el-col :span="4">
                <el-card shadow="hover" style="height: 48px;">
                  <div class="cluster-content pod-content app-two-content">{{ cluster_detail.job_num }}</div>
                  <div class="cluster-title pod-title">Job</div>
                </el-card>
                <el-card shadow="hover" style="height: 48px; margin-top: 4px;">
                  <div class="cluster-content pod-content app-two-content">{{ cluster_detail.cronjob_num }}</div>
                  <div class="cluster-title pod-title">CronJob</div>
                </el-card>
              </el-col>
            </el-row>
          </div>
        </el-col>

        <el-col :span="4">
          <div class="border-class">
            <div style="margin: 5px 0px 10px;">网络状态</div>
            <el-row :gutter="5">
              
              <el-col :span="12">
                <el-card shadow="hover" style="height: 100px">
                  <div class="cluster-content app-content">{{ cluster_detail.service_num }}</div>
                  <div class="cluster-title app-title">Service</div>
                </el-card>
              </el-col>
              <el-col :span="12">
                <el-card shadow="hover" style="height: 100px">
                  <div class="cluster-content app-content">{{ cluster_detail.ingress_num }}</div>
                  <div class="cluster-title app-title">Ingress</div>
                </el-card>
              </el-col>
            </el-row>
          </div>
        </el-col>

        <el-col :span="8">
          <div class="border-class">
            <div style="margin: 5px 0px 10px;">存储状态</div>
            <el-row :gutter="5">
              <el-col :span="5">
                <el-card shadow="hover" style="height: 100px">
                  <div class="cluster-content app-content">{{ cluster_detail.storageclass_num }}</div>
                  <div class="cluster-title app-title">Storage Class</div>
                </el-card>
              </el-col>
              <el-col :span="5">
                <el-card shadow="hover" style="height: 100px">
                  <div class="cluster-content app-content">{{ cluster_detail.pv_num }}</div>
                  <div class="cluster-title app-title">PV</div>
                </el-card>
              </el-col>
              <el-col :span="9">
                <el-row :gutter="5">
                  <el-col :span="12">
                    <el-card shadow="never" style="height: 44px; color: #409EFF; margin-top: 3px;">
                      <div class="cluster-content pod-content">{{ cluster_detail.pv_available_num }}</div>
                      <div class="cluster-title pod-title">Available</div>
                    </el-card>
                    <el-card shadow="never" style="height: 44px; color: #67C23A; margin-top: 6px;">
                      <div class="cluster-content pod-content">{{ cluster_detail.pv_bound_num }}</div>
                      <div class="cluster-title pod-title">Bound</div>
                    </el-card>
                  </el-col>
                  <el-col :span="12">
                    <el-card shadow="never" style="height: 45px; color: #E6A23C; margin-top: 3px;">
                      <div class="cluster-content pod-content">{{ cluster_detail.pv_released_num }}</div>
                      <div class="cluster-title pod-title">Released</div>
                    </el-card>
                    <el-card shadow="never" style="height: 45px; color: #F56C6C; margin-top: 6px;">
                      <div class="cluster-content pod-content">{{ cluster_detail.pv_failed_num }}</div>
                      <div class="cluster-title pod-title">Failed</div>
                    </el-card>
                  </el-col>
                </el-row>
              </el-col>
              <el-col :span="5">
                <el-card shadow="hover" style="height: 100px">
                  <div class="cluster-content app-content">{{ cluster_detail.pvc_num }}</div>
                  <div class="cluster-title app-title">PVC</div>
                </el-card>
              </el-col>
            </el-row>
          </div>
        </el-col>
      </el-row>

      <el-row>
        <div class="border-class event-class" >
          <div style="margin: 5px 0px 10px;">事件</div>
          <el-timeline v-if="events && events.length > 0" :style="{height: maxHeight + 'px', 'overflow': 'auto'}" :max-height="maxHeight">
            <template v-for="e of events">
              <el-timeline-item :key="e.uid" :timestamp="$dateFormat(e.event_time)" placement="top">
                <el-card shadow="never">
                  <p class="event-title">{{ e.object.kind }}/{{ e.object.name }}</p>
                  <p class="event-body">{{ e.message }}</p>
                </el-card>
              </el-timeline-item>
            </template>
          </el-timeline>
          <div v-else style=" padding: 10px 15px 25px; color: #909399; text-align: center">暂无事件发生</div>
        </div>
      </el-row>
    </div>
  </div>
</template>

<script>
import { ResType, listResource, getResource } from '@/api/cluster/resource'
import { Message } from 'element-ui'

export default {
  name: 'cluster',
  data() {
    return {
      cluster_detail: {},
      originEvents: [],
      loading: true,
      maxHeight: window.innerHeight - 410,
    }
  },
  components: {
  },
  created() {
    this.fetchData()
  },
  watch: {
    cluster: function() {
      this.fetchData()
    }
  },
  computed: {
    events: function() {
      let dlist = []
      for (let p of this.originEvents) {
        dlist.push(p)
      }
      dlist.sort((a, b) => {
        return a.event_time < b.event_time ? 1 : -1
      })
      return dlist
    },
    cluster() {
      return this.$store.state.cluster
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originCronJobs = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        let clusterOk, eventOk = false
        getResource(cluster, ResType.Cluster).then(response => {
          // this.loading = false
          this.cluster_detail = response.data
          clusterOk = true
          if (clusterOk && eventOk) {
            this.loading = false
          }
        }).catch(() => {
          this.loading = false
        })
        listResource(cluster, ResType.Event).then(response => {
          this.originEvents = response.data ? response.data : []
          eventOk = true
          if (clusterOk && eventOk) {
            this.loading = false
          }
        }).catch(() => {
        })
      } else {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
  }
}
</script>

<style lang="scss" scoped>
@import "~@/styles/variables.scss";
.border-class {
    //border: 1px solid #EBEEF5;
    padding: 5px 10px 10px;
    background-color: #fff;
    border-radius: 10px;

    .event-title {
        font-size: 14px;
        font-weight: 400;
        color: #1f2f3d;
    }
    .event-body {
        font-size: 14px;
        color: #5e6d82;
    }
}

.cluster-content {
  text-align: center; 
  font-size: 18px;
  margin-top: 0px;
}

.cluster-title {
  text-align: center; 
  margin-top: 6px; 
  font-size: 13px; 
  opacity: 0.6; 
  height: 20px;
}

.app-content {
  margin-top: 16px;
}

.app-title {
  margin-top: 10px;
}

.pod-content {
  font-size: 15px;
  margin-top: -5px;
}

.pod-title {
  font-size: 12px;
  margin-top: 1px;
}
.app-two-content {
  margin-top: -3px;
}
</style>
<style>
.row-class .el-card__body {
    padding: 10px 5px;
    /* height: 100px; */
}
.event-class .el-card__body {
    padding: 5px 20px;
    /* height: 100px; */
}
</style>