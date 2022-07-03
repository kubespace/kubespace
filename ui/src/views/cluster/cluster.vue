<template>
  <div>
    <div class="dashboard-container" v-loading="loading" style="margin-top: 20px;">
      <div style="margin: 10px 0px 10px;">集群状态</div>
      <el-row :gutter="20" style="margin-bottom: 10px;" class="row-class">
        <el-col :span="4">
            <OverviewCard icon="cluster" title="Cluster Version" :number="cluster_detail.cluster_versio" style="height: 100px;"/>
        </el-col>
        <el-col :span="4">
          <OverviewCard icon="cluster" title="CPU" :number="cluster_detail.cluster_cpu" style="height: 45px;"/>
          <OverviewCard icon="cluster" title="Memory" :number="cluster_detail.cluster_memory" style="height: 45px;margin-top: 10px;"/>
        </el-col>
        <el-col :span="4">
          <OverviewCard icon="cluster" title="Node" :number="cluster_detail.node_num" style="height: 100px;"/>
        </el-col>
        <el-col :span="4">
          <OverviewCard icon="cluster" title="Namespace" :number="cluster_detail.namespace_num" style="height:100px;"/>
        </el-col>
      </el-row>
      <div style="margin: 5px 0px 10px;">应用状态</div>
      <el-row :gutter="10" style="margin-bottom: 10px;" class="row-class">
          <el-col :span="4">
            <OverviewCard icon="workloads" title="Pod" :number="cluster_detail.pod_num" style="height: 100px;"/>
          </el-col>
          <el-col :span="3">
            <OverviewCard icon="workloads" title="Running" :number="cluster_detail.pod_running_num" style="height: 45px; color: #409EFF;"/>
            <OverviewCard icon="workloads" title="Pending" :number="cluster_detail.pod_pending_num" style="height: 45px; color: #E6A23C; margin-top: 10px;"/>
          </el-col>
          <el-col :span="3">
            <OverviewCard icon="workloads" title="Succeeded" :number="cluster_detail.pod_succeeded_num" style="height: 45px; color: #67C23A; "/>
            <OverviewCard icon="workloads" title="Failed" :number="cluster_detail.pod_failed_num" style="height: 45px; margin-top: 6px; color: #F56C6C;margin-top: 10px;"/>
          </el-col>
          <el-col :span="4">
            <OverviewCard icon="workloads" title="Deployment" :number="cluster_detail.deployment_num" style="height: 100px;"/>
          </el-col>
          <el-col :span="4">
            <OverviewCard icon="workloads" title="StatefulSet" :number="cluster_detail.statefulset_num" style="height: 100px;"/>
          </el-col>
          <el-col :span="4">
            <OverviewCard icon="workloads" title="DaemonSet" :number="cluster_detail.daemonset_num" style="height: 100px;"/>
          </el-col>
      </el-row>
       <div style="margin: 5px 0px 10px;">网络状态</div>
       <el-row :gutter="10" style="margin-bottom: 10px;" class="row-class">
          <el-col :span="4">
            <OverviewCard icon="network" title="Service" :number="cluster_detail.service_num" style="height: 100px;"/>
          </el-col>
          <el-col :span="4">
            <OverviewCard icon="network" title="Ingress" :number="cluster_detail.ingress_num" style="height: 100px;"/>
          </el-col>
       </el-row>
       <div style="margin: 5px 0px 10px;">存储状态</div>
       <el-row :gutter="10" style="margin-bottom: 10px;" class="row-class">
          <el-col :span="4">
            <OverviewCard icon="storage" title="Storage Class" :number="cluster_detail.storageclass_num" style="height: 100px;"/>
          </el-col>
          <el-col :span="3">
             <OverviewCard icon="storage" title="Available" :number="cluster_detail.pv_available_num" style="height: 45px; color: #409EFF;"/>
             <OverviewCard icon="storage" title="Bound" :number="cluster_detail.pv_bound_num" style="height: 45px; color: #67C23A;margin-top: 10px;"/>
          </el-col>
          <el-col :span="3">
             <OverviewCard icon="storage" title="Released" :number="cluster_detail.pv_released_num" style="height: 45px; color: #E6A23C;"/>
             <OverviewCard icon="storage" title="Failed" :number="cluster_detail.pv_failed_num" style="height: 45px; color: #F56C6C;margin-top: 10px;"/>
          </el-col>
          <el-col :span="4">
           <OverviewCard icon="storage" title="PV" :number="cluster_detail.pv_num" style="height: 100px;"/>
          </el-col>
          <el-col :span="4">
           <OverviewCard icon="storage" title="PVC" :number="cluster_detail.pvc_num" style="height: 100px;"/>
          </el-col>
      </el-row>

      <div style="margin: 5px 0px 10px;background-color: #f2f2f2;">事件</div>
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
  </div>
</template>

<script>
import { clusterDetail } from '@/api/cluster'
import { listEvents } from '@/api/event'
import { Message } from 'element-ui'
import OverviewCard from "../workspace/overviewCard.vue";
export default {
  name: 'cluster',
  data() {
    return {
      cluster_detail: {},
      originEvents: [],
      loading: true,
      maxHeight: window.innerHeight - 530,
    }
  },
  components: { OverviewCard },
  created() {
    this.fetchData()
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
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originCronJobs = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        clusterDetail(cluster).then(response => {
          this.loading = false
          this.cluster_detail = response.data
          console.log(this.cluster_detail)
        }).catch(() => {
          this.loading = false
        })
        listEvents(cluster).then(response => {
          this.originEvents = response.data ? response.data : []
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
    border: 1px solid #EBEEF5;
    padding: 5px 10px 10px;

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