<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc" 
      :createFunc="createFunc" createDisplay="创建"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="deployments"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort = "{prop: 'name'}"
        @selection-change="handleSelectionChange"
        row-key="uid"
        >
        <el-table-column
          type="selection"
          width="45">
        </el-table-column>
        <el-table-column
          prop="name"
          label="名称"
          min-width="100"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.namespace, scope.row.name)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="60"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="ready_replicas"
          label="Pods"
          min-width="40"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ scope.row.ready_replicas }}/{{ scope.row.status_replicas }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="replicas"
          label="副本"
          min-width="30"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="strategy"
          label="更新策略"
          min-width="55"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="conditions"
          label="状态"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.conditions">
              <span v-for="c in scope.row.conditions" :key="c">
                <span v-if="c === 'Available'" style="color: rgb(102, 177, 255)">{{ c }}</span>
                <span v-if="c === 'Progressing'" style="color: #67C23A"> {{ c }}</span>
                <span v-if="c === 'ReplicaFailure'" style="color: #F56C6C"> {{ c }}</span>
              </span>
            </template>
          </template>
        </el-table-column>
        <el-table-column
          prop="created"
          label="创建时间"
          min-width="70"
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
                <el-dropdown-item @click.native.prevent="nameClick(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$updatePerm()"
                  @click.native.prevent="update_replicas_deployment = scope.row; update_replicas = scope.row.replicas; replicaDialog = true;">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="scale" />
                  <span style="margin-left: 5px;">副本</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$updatePerm()" @click.native.prevent="getDeploymentYaml(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$deletePerm()" 
                  @click.native.prevent="deleteDeployments([{namespace: scope.row.namespace, name: scope.row.name}])">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                  <span style="margin-left: 5px;">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
      <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
        <el-button plain @click="updateDeployment()" size="small">确 定</el-button>
      </span>
    </el-dialog>
    <el-dialog title="扩缩容" :visible.sync="replicaDialog" :close-on-click-modal="false" width="380px" top="14%" class="replicaDialog" >
      <div slot="title">
        <span style="line-height: 24px; font-size: 18px; color: #303133;">扩缩容:</span>
        <span style="line-height: 24px; font-size: 15px; color: #606266;">{{ update_replicas_deployment ? update_replicas_deployment.name : '' }}</span>
      </div>
      <!-- <label style="margin-bottom: 10px;">{{ update_replicas_deployment ? update_replicas_deployment.name : '' }}</label> -->
      <el-form ref="form" label-position="left" label-width="100px">
        <el-form-item label="当前副本数">
          <label>{{ update_replicas_deployment ? update_replicas_deployment.replicas : 0 }}</label>
        </el-form-item>
      </el-form>
      <el-form ref="form" label-position="left" label-width="100px">
        <el-form-item label="期望副本数">
          <el-input-number size="medium" v-model="update_replicas" :min="1" :max="100"></el-input-number>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="replicaDialog = false" size="small">取 消</el-button>
        <el-button plain @click="updateDeploymentObj({deployment: update_replicas_deployment, replicas: update_replicas})" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listDeployments, getDeployment, deleteDeployments, updateDeployment, updateDeploymentObj } from '@/api/deployment'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'

export default {
  name: 'Deployment',
  components: {
    Clusterbar,
    Yaml
  },
  data() {
      return {
        replicaDialog: false,
        yamlDialog: false,
        yamlNamespace: "",
        yamlName: "",
        yamlValue: "",
        yamlLoading: true,
        cellStyle: {border: 0},
        titleName: ["Deployments"],
        maxHeight: window.innerHeight - 150,
        loading: true,
        originDeployments: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delDeployment: [],
        update_replicas: 0,
        update_replicas_deployment: null,
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
        // console.log(heightStyle)
        that.maxHeight = heightStyle
      })()
    }
  },
  watch: {
    deploymentsWatch: function (newObj) {
      if (newObj) {
        let newUid = newObj.resource.metadata.uid
        let newRv = newObj.resource.metadata.resourceVersion
        if (newObj.event === 'add') {
          this.originDeployments.push(this.buildDeployments(newObj.resource))
        } else if (newObj.event === 'update') {
          for (let i in this.originDeployments) {
            let d = this.originDeployments[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newDp = this.buildDeployments(newObj.resource)
                this.$set(this.originDeployments, i, newDp)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originDeployments = this.originDeployments.filter(( { uid } ) => uid !== newUid)
        }
      }
    }
  },
  computed: {
    deployments: function() {
      let dlist = []
      for (let p of this.originDeployments) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        if (p.conditions && p.conditions.length > 0) p.conditions.sort()
        dlist.push(p)
      }
      return dlist
    },
    deploymentsWatch: function() {
      return this.$store.getters["ws/deploymentWatch"]
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originDeployments = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listDeployments(cluster).then(response => {
          this.loading = false
          this.originDeployments = response.data
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
    buildDeployments: function(deployment) {
      if (!deployment) return
      var conditions = []
      for (let c of deployment.status.conditions) {
        if (c.status === "True") {
          conditions.push(c.type)
        }
      }
      let p = {
        uid: deployment.metadata.uid,
        namespace: deployment.metadata.namespace,
        name: deployment.metadata.name,
        replicas: deployment.spec.replicas,
        status_replicas: deployment.status.replicas || 0,
        ready_replicas: deployment.status.readyReplicas || 0,
        update_replicas: deployment.status.updateReplicas,
        available_replicas: deployment.status.availableReplicas,
        unavailable_replicas: deployment.status.unavailabelReplicas,
        resource_version: deployment.metadata.resourceVersion,
        strategy: deployment.spec.strategy.type,
        conditions: conditions,
        created: deployment.metadata.creationTimestamp
      }
      return p
    },
    nameClick: function(namespace, name) {
      this.$router.push({name: 'deploymentDetail', params: {namespace: namespace, deploymentName: name}})
    },
    getDeploymentYaml: function(namespace, name) {
      this.yamlNamespace = ""
      this.yamlName = ""
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!namespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (!name) {
        Message.error("获取Deployment名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getDeployment(cluster, namespace, name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteDeployments: function(deployments) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( deployments.length <= 0 ){
        Message.error("请选择要删除的Deployment")
        return
      }
      let params = {
        resources: deployments
      }
      deleteDeployments(cluster, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    updateDeployment: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!this.yamlNamespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (!this.yamlName) {
        Message.error("获取Deployment参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      updateDeployment(cluster, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
      }).catch(() => {
        // console.log(e) 
      })
    },
    updateDeploymentObj: function(update_obj) {
      console.log(update_obj)
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!update_obj || !update_obj.deployment) {
        Message.error("获取更新参数异常，请刷新重试")
        return
      }
      let deployment = update_obj.deployment
      if (!deployment.namespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (!deployment.name) {
        Message.error("获取Deployment参数异常，请刷新重试")
        return
      }
      let update_params = {}
      if (update_obj.replicas) {
        if (update_obj.replicas < 1) {
          Message.error("副本数不能小于1，请重新输入")
          return
        }
        if (parseInt(update_obj.replicas) === parseInt(deployment.replicas)) {
          Message.error("副本数与当前值相同，请重新输入")
          return
        }
        update_params["replicas"] = update_obj.replicas
      }
      if (Object.keys(update_params).length === 0) {
        Message.error("更新参数为空")
        return
      }
      updateDeploymentObj(cluster, deployment.namespace, deployment.name, update_params).then(() => {
        Message.success("更新成功")
        this.replicaDialog = false;
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delDeploymentsFunc: function() {
      if (this.delDeployments.length > 0){
        let delDeployments = []
        for (var p of this.delDeployments) {
          delDeployments.push({namespace: p.namespace, name: p.name})
        }
        this.deleteDeployments(delDeployments)
      }
    },
    handleSelectionChange(val) {
      this.delDeployments = val;
      if (val.length > 0){
        this.delFunc = this._delDeploymentsFunc
      } else {
        this.delFunc = undefined
      }
    },
    createFunc() {
      this.$router.push({name: 'deploymentCreate'})
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
