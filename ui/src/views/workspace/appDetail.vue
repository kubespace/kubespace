<template>
  <div>
    <clusterbar :titleName="titleName"/>
    <div class="dashboard-container detail-dashboard" ref="tableCot" v-loading="loading">
      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" inline class="pod-item" label-width="80px" style="margin: 15px 10px 30px 10px;">
          <el-form-item label="名称">
            <span>{{ originApp.name }}</span>
          </el-form-item>
          <el-form-item label="状态">
            <span :style="{color: statusColorMap[originApp.status]}">{{ statusNameMap[originApp.status] }}</span>
          </el-form-item>
          <el-form-item label="集群">
            <span>{{ originApp.cluster }}</span>
          </el-form-item>
          <el-form-item label="版本">
            <span>{{ originApp.package_version }}</span>
          </el-form-item>
          <el-form-item label="更新时间">
            <span>{{ $dateFormat(originApp.update_time) }}</span>
          </el-form-item>
          <el-form-item label="命名空间">
            <span>{{ originApp.namespace }}</span>
          </el-form-item>
        </el-form>
      </div>

      <div style="padding: 0px 8px;" v-if="['Running', 'RunningFault', 'NotReady'].indexOf(originApp.status) > -1">
        <div>Pods</div>
        <div class="msgClass" style="margin: 15px 10px 20px 10px;">
            <el-table
              ref="table"
              :data="pods"
              class="table-fix"
              tooltip-effect="dark"
              style="width: 100%"
              v-loading="loading"
              :cell-style="cellStyle"
              :default-sort = "{prop: 'name'}"
              >
              <el-table-column
                prop="name"
                label="名称"
                min-width="150"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>
                    {{ scope.row.name }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column
                prop="containerNum"
                label="容器"
                min-width="45"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <template v-if="scope.row.initContainers">
                  <el-tooltip :content="`${c.name} (${c.status})`" placement="top" v-for="c in scope.row.initContainers" :key="c.name">
                    <svg-icon style="margin-top: 7px;" :class="containerClass(c.status)" icon-class="square" />
                  </el-tooltip>
                  </template>
                  <el-tooltip :content="`${c.name} (${c.status})`" placement="top" v-for="c in scope.row.containers" :key="c.name">
                    <svg-icon style="margin-top: 7px;" :class="containerClass(c.status)" icon-class="square" />
                  </el-tooltip>
                </template>
              </el-table-column>
              <el-table-column
                prop="restarts"
                label="重启"
                min-width="45"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="node_name"
                label="节点"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="ip"
                label="IP"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="created"
                label="创建时间"
                min-width="100"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                prop="status"
                label="状态"
                min-width="60"
                show-overflow-tooltip>
              </el-table-column>
              <el-table-column
                label=""
                show-overflow-tooltip
                width="45">
                <template slot-scope="scope">
                  <el-dropdown size="medium" >
                    <el-link :underline="false"><svg-icon style="width: 1.3em; height: 1.3em;" icon-class="operate" /></el-link>
                    <el-dropdown-menu slot="dropdown">
                      <div @mouseout="logContainerShow = false;">
                        <el-dropdown-item @click.native.prevent="selectContainer = scope.row.containers[0].name; selectPodName = scope.row.name; logDialog = true;">
                          <div class="download">
                            <div>
                              <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="log" />
                              <span style="margin-left: 5px;">日志</span>
                            </div>
                            <div class="download-right" v-show="scope.row.containerNum > 1 && logContainerShow">
                              <div class="download-item" v-for="c in scope.row.init_containers" :key="c.name"
                                  @click="selectContainer = c.name; selectPodName = scope.row.name; logDialog = true;">
                                  {{ c.name }}
                              </div>
                              <div class="download-item" v-for="c in scope.row.containers" :key="c.name"
                                  @click="selectContainer = c.name; selectPodName = scope.row.name; logDialog = true;">
                                  {{ c.name }}
                              </div>
                            </div>
                          </div>
                        </el-dropdown-item>
                      </div>
                      <div @mouseover="termContainerShow = true;" @mouseout="termContainerShow = false;">
                        <el-dropdown-item @click.native.prevent="selectContainer = scope.row.containers[0].name; selectPodName = scope.row.name; terminalDialog = true;">
                          <div class="download">
                            <div>
                              <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="terminal" />
                              <span style="margin-left: 5px;">终端</span>
                            </div>
                            <div class="download-right" v-show="scope.row.containers.length > 1 && termContainerShow">
                              <div class="download-item" v-for="c in scope.row.containers" :key="c.name"
                                  @click="selectContainer = c.name; selectPodName = scope.row.name; terminalDialog = true;">
                                  {{ c.name }}
                              </div>
                            </div>
                          </div>
                        </el-dropdown-item>
                      </div>
                      <el-dropdown-item @click.native.prevent="deletePods([{namespace: scope.row.namespace, name: scope.row.name}])">
                        <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                        <span style="margin-left: 5px;">删除</span>
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </el-dropdown>
                </template>
              </el-table-column>
            </el-table>
          </div>
      </div>

      <el-tabs value="workloads" style="padding: 0px 8px;">
        <el-tab-pane label="Workloads" name="workloads">
          <div class="msgClass">
            <el-table
            ref="table"
            :data="workloads"
            class="table-fix"
            tooltip-effect="dark"
            style="width: 100%"
            :cell-style="cellStyle"
            :default-sort = "{prop: 'name'}"
            >
              <el-table-column
                prop="name"
                label="名称"
                min-width="10"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>{{ scope.row.metadata.name }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="kind"
                label="类型"
                min-width="10"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>{{ scope.row.kind }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="kind"
                label="副本数"
                min-width="4"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>{{ scope.row.spec.replicas }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop=""
                label="镜像"
                min-width="14"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span v-for="s in scope.row.spec.template.spec.containers" :key="s.name" class="back-class">
                    {{ s.image }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column v-if="originApp.status != 'UnInstall'"
                prop="image_pull_policy"
                label="Pods"
                min-width="5"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>{{ scope.row.status.readyReplicas }}/{{ scope.row.status.replicas }}</span>
                </template>
              </el-table-column>
              <el-table-column v-if="originApp.status != 'UnInstall'"
                prop=""
                label="创建时间"
                min-width="10"
                show-overflow-tooltip>
                <template slot-scope="scope">
                  <span>{{ $dateFormat(scope.row.metadata.creationTimestamp) }}</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
        <el-tab-pane label="Service" name="service">
        </el-tab-pane>
        <el-tab-pane label="ConfigMap" name="configmap">
        </el-tab-pane>
        <el-tab-pane label="Secret" name="secret">
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { getApp } from '@/api/project/apps'
import { containerClass, buildPods } from '@/api/pods'
import { Message } from 'element-ui'

export default {
  name: 'AppDetail',
  components: {
    Clusterbar
  },
  data() {
    return {
      titleName: ["应用管理"],
      search_name: '',
      users: [],
      cellStyle: {border: 0},
      maxHeight: window.innerHeight - 150,
      loading: true,
      originApp: {},
      // pods: [],
      containers: [],
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
    }
  },
  created() {
    this.fetchData()
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 150
        that.maxHeight = heightStyle
      })()
    }
  },
  computed: {
    projectId() {
      return this.$route.params.workspaceId
    },
    appId() {
      return this.$route.params.appId
    },
    pods() {
      let pods = []
      if(this.originApp.release) {
        for(let obj of this.originApp.release.objects) {
          if(obj.kind == 'Pod') pods.push(buildPods(obj))
        }
      }
      return pods
    },
    workloads() {
      let workloads = []
      if(this.originApp.release) {
        for(let obj of this.originApp.release.objects) {
          if(['Deployment', 'StatefulSet', 'DaemonSet', 'CronJob', 'Job'].indexOf(obj.kind) > -1){
            workloads.push(obj)
          }
        }
      }
      return workloads
    }
  },
  methods: {
    containerClass,
    fetchData() {
      if(!this.appId) {
        Message.error("获取应用id错误，请刷新重试")
        return
      }
      this.loading = true
      getApp(this.appId)
        .then((response) => {
          this.originApp = response.data || {};
          this.titleName = ["应用管理", this.originApp.name]
          this.loading = false
        })
        .catch(() => {
          this.loading = false
        })
    },
  },
}
</script>

<style lang="scss" scoped>

</style>
