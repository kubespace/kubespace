<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="projectId ? undefined : nsSearch" 
      :nameFunc="nameSearch"  :createFunc="openCreateFormDialog"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="services"
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
          prop="name"
          label="名称"
          min-width="40"
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
          min-width="30"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="type"
          label="类型"
          min-width="25"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="cluster_ip"
          label="ClusterIP"
          min-width="35"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="ports"
          label="端口"
          min-width="30"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.ports && scope.row.ports.length > 0">
              <span>{{ getPortsDisplay(scope.row.ports) }}</span>
            </template>
          </template>
        </el-table-column>
        <!-- <el-table-column
          prop="external_ip"
          label="ExternalIP"
          min-width="40"
          show-overflow-tooltip>
        </el-table-column> -->
        <el-table-column
          prop="selector"
          label="选择器"
          min-width="55"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.selector">
              <span v-for="(val, key) in scope.row.selector" :key="key" class="back-class">
                {{ key + '=' + val }} 
              </span>
            </template>
            <!-- <span v-else>--</span> -->
          </template>
        </el-table-column>
        <el-table-column
          prop="created"
          label="创建时间"
          min-width="40"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ $dateFormat(scope.row.created) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" show-overflow-tooltip width="110px">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 15px;" @click="openUpdateFormDialog(scope.row.namespace, scope.row.name)">编辑</el-link>
              <el-link :disabled="!$editorRole()" :underline="false" type="danger" @click="handleDeleteServices([{namespace: scope.row.namespace, name: scope.row.name}])">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog :title="updateFormVisible ? '编辑Service' : '创建Service'" :visible.sync="createFormVisible"
      @close="closeFormDialog" :destroy-on-close="true" width="70%" :close-on-click-modal="false">
      <div v-loading="dialogLoading">
        <div class="dialogContent" style="">
          <el-form :model="service.metadata" :rules="rules" ref="form" label-position="left" label-width="105px">
            <el-form-item label="服务名称" prop="name" required>
              <el-input :disabled="updateFormVisible" v-model="service.metadata.name" style="width: 50%" placeholder="请输入服务名称" size="small"></el-input>
            </el-form-item>
            <el-form-item label="命名空间" prop="" required="">
              <span v-if="namespace">{{ namespace }}</span>
              <el-select v-else :disabled="updateFormVisible" v-model="service.metadata.namespace" placeholder="请选择命名空间"
                size="small" style="width: 50%;" >
                <el-option
                  v-for="item in namespaces"
                  :key="item.name"
                  :label="item.name"
                  :value="item.name">
                </el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="服务类型" style="width: 100%" prop="" required>
              <el-radio-group v-model="service.spec.type"  size="small">
                <el-radio-button label="ClusterIP"></el-radio-button>
                <el-radio-button label="NodePort"></el-radio-button>
                <el-radio-button label="LoadBalancer"></el-radio-button>
                <el-radio-button label="ExternalName"></el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="Pod选择器" prop="" required>
              <div style="margin-bottom: 5px;" v-for="(l, i) in service.spec.selector" :key="i">
                <el-input size="small" v-model="l.key" style="width: 25%;" placeholder="Key"></el-input> = 
                <el-input size="small" v-model="l.value" style="width: 25%;" placeholder="Value"></el-input>
                <el-button size="mini" circle style="padding: 5px; margin-left: 10px;" @click="service.spec.selector.splice(i, 1)" 
                  icon="el-icon-close"></el-button>
              </div>
              <el-button plain size="small" @click="service.spec.selector.push({key: '', value: ''})" icon="el-icon-plus"
                style="border-radius: 0px;">添加</el-button>
            </el-form-item>

            <el-form-item label="端口配置" required >
              <el-row style="margin-bottom: 5px; margin-top: 8px;">
                <el-col :span="service.spec.type == 'NodePort' ? 5 : 6" style="background-color: #F5F7FA; padding-left: 10px;">
                  <div class="border-span-header">
                    名称
                  </div>
                </el-col>
                <el-col :span="service.spec.type == 'NodePort' ? 5 : 6" style="background-color: #F5F7FA">
                  <div class="border-span-header">
                    <span  class="border-span-content">*</span>服务端口
                  </div>
                </el-col>
                <el-col :span="service.spec.type == 'NodePort' ? 5 : 6" style="background-color: #F5F7FA;">
                  <div class="border-span-header">
                    <span  class="border-span-content">*</span>容器端口
                  </div>
                </el-col>
                <el-col v-if="service.spec.type == 'NodePort'" :span="5" style="background-color: #F5F7FA">
                  <!-- <div class="border-span-header"> -->
                    NodePort
                  <!-- </div> -->
                </el-col>
                <el-col :span="service.spec.type == 'NodePort' ? 3 : 5" style="background-color: #F5F7FA">
                  <div class="border-span-header">
                    协议
                  </div>
                </el-col>
                <!-- <el-col :span="5"><div style="width: 100px;"></div></el-col> -->
              </el-row>
              <el-row style="padding-top: 0px;" v-for="(item, idx) in service.spec.ports" :key="idx">
                <el-col :span="service.spec.type == 'NodePort' ? 5 : 6">
                  <div class="border-span-header">
                    <el-input v-model="item.name" size="small" style="padding-right: 10px" placeholder="服务端口名称"></el-input>
                  </div>
                </el-col>
                <el-col :span="service.spec.type == 'NodePort' ? 5 : 6">
                  <div class="border-span-header">
                    <el-input-number :controls="false" v-model="item.port" size="small" style="width:100%;padding-right: 10px" placeholder="服务暴露端口"></el-input-number>
                  </div>
                </el-col>
                <el-col :span="service.spec.type == 'NodePort' ? 5 : 6">
                  <div class="border-span-header">
                    <el-input-number :controls="false" v-model="item.targetPort" size="small" style="width:100%;padding-right: 10px" placeholder="容器访问端口，如:80"></el-input-number>
                  </div>
                </el-col>
                <el-col v-if="service.spec.type == 'NodePort'" :span="5">
                  <!-- <div class="border-span-header"> -->
                    <el-input-number :controls="false" v-model="item.nodePort" size="small" style="width:100%;padding-right: 10px" placeholder="宿主机暴露端口"></el-input-number>
                  <!-- </div> -->
                </el-col>
                <el-col :span="service.spec.type == 'NodePort' ? 3 : 5">
                  <div class="border-span-header">
                    <el-select v-model="item.protocol" placeholder="端口所属协议" size="small">
                      <el-option label="TCP" value="TCP"></el-option>
                      <el-option label="UDP" value="UDP"></el-option>
                      <el-option label="SCTP" value="SCTP"></el-option>
                    </el-select>
                  </div>
                </el-col>
                <el-col :span="1" style="padding-left: 10px">
                  <el-button circle size="mini" style="padding: 5px;" 
                    @click="service.spec.ports.splice(idx, 1)" icon="el-icon-close"></el-button>
                </el-col>
              </el-row>
              <el-row>
                <el-col :span="service.spec.type == 'NodePort' ? 23 : 23">
                <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px; border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
                  @click="service.spec.ports.push({protocol: 'TCP'})" icon="el-icon-plus">添加服务端口</el-button>
                </el-col>
              </el-row>
            </el-form-item>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter" style="margin-top: 20px;">
          <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="updateFormVisible ? handleUpdateService() : handleCreateService()" >确 定</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { ResType, listResource, getResource, delResource, updateResource, createResource } from '@/api/cluster/resource'
import { projectLabels } from '@/api/project/project'
import { Message } from 'element-ui'
import yaml from 'js-yaml'

export default {
  name: 'Service',
  components: {
    Clusterbar,
  },
  data() {
    return {
      yamlDialog: false,
      yamlNamespace: "",
      yamlName: "",
      yamlValue: "",
      yamlLoading: true,
      cellStyle: {border: 0},
      titleName: ["Services"],
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      originServices: [],
      search_ns: [],
      search_name: '',
      delFunc: undefined,
      delServices: [],
      dialogLoading: false,
      createFormVisible: false,
      updateFormVisible: false,
      service: {
        kind: "Service",
        apiVersion: "v1",
        metadata: {
          name: "",
        },
        spec: {
          ports: [],
          selector: [{}],
          type: 'ClusterIP',
        }
      },
      rules: {},
      namespaces: [],
    }
  },
  created() {
    this.fetchData()
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - this.$contentHeight
        // console.log(heightStyle)
        that.maxHeight = heightStyle
      })()
    }
  },
  watch: {
    cluster: function() {
      this.fetchData()
    }
  },
  computed: {
    services: function() {
      let dlist = []
      for (let p of this.originServices) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        
        dlist.push(p)
      }
      return dlist
    },
    projectId() {
      return this.$route.params.workspaceId
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    namespace: function() {
      return this.$store.state.namespace || ''
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      const cluster = this.$store.state.cluster
      let params = {namespace: this.namespace}
      if(this.projectId) params['label_selector'] = {"matchLabels": projectLabels()}
      if (cluster) {
        listResource(cluster, ResType.Service, params, {project_id: this.projectId}).then(response => {
          this.loading = false
          let originServices = response.data || []
          this.$set(this, 'originServices', originServices)
        }).catch(() => {
          this.loading = false
        })
      } else if(!this.projectId) {
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
    buildServices: function(service) {
      if (!service) return
      let p = {
        uid: service.metadata.uid,
        namespace: service.metadata.namespace,
        name: service.metadata.name,
        type: service.spec.type,
        cluster_ip: service.spec.clusterIP,
        ports: service.spec.ports,
        external_ip: service.spec.externalIPs,
        selector: service.spec.selector,
        resource_version: service.metadata.resourceVersion,
        created: service.metadata.creationTimestamp
      }
      return p
    },
    nameClick: function(namespace, name) {
      let routeName = this.projectId ? 'workspaceServiceDetail' : 'serviceDetail'
      this.$router.push({name: routeName, params: {namespace: namespace, serviceName: name}})
    },
    getService: function(namespace, name) {
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
        Message.error("获取Service名称参数异常，请刷新重试")
        return
      }
      this.dialogLoading = true
      getResource(cluster, ResType.Service, namespace, name, '', {project_id: this.projectId}).then(response => {
        let service = response.data
        let selector = []
        for(let k in service.spec.selector) {
          selector.push({key: k, value: service.spec.selector[k]})
        }
        service.spec.selector = selector
        this.service = service
        this.dialogLoading = false
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    handleDeleteServices: function(services) {
      let cs = ''
      for(let c of services) {
        cs += `${c.namespace}/${c.name},`
      }
      cs = cs.substr(0, cs.length - 1)
      this.$confirm(`请确认是否删除「${cs}」Service?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        delResource(this.cluster, ResType.Service, {resources: services}, {project_id: this.projectId}).then(() => {
          Message.success("删除Service成功")
          this.loading = false
          this.fetchData()
        }).catch((err) => {
          this.loading = false
        });
      }).catch(() => {       
      });
    },
    handleCreateService: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      // if(this.service.spec.selector.length == 0) {
      //   Message.error("Pod选择器不能为空")
      //   return
      // }
      let service = JSON.parse(JSON.stringify(this.service))
      if(this.namespace){
        service.metadata.namespace = this.namespace
      }
      if(!service.metadata.namespace) {
        Message.error("命名空间不能为空")
        return
      }
      if(this.projectId) {
        service.metadata.labels = projectLabels()
      }
      let selector = {}
      for(let s of service.spec.selector) {
        if(!s.key) {
          Message.error("Pod选择器Key不能为空")
          return
        }
        selector[s.key] = s.value
      }
      service.spec.selector = selector
      for(let p of service.spec.ports) {
        try{
          p['port'] = parseInt(p.port)
          p['targetPort'] = parseInt(p.targetPort)
          if(p.nodePort) p['nodePort'] = parseInt(p.nodePort)
        } catch(e) {
          Message.error("端口错误")
          return
        }
      }
      let yamlStr = yaml.dump(service)
      this.dialogLoading = true
      createResource(cluster, yamlStr, {project_id: this.projectId}).then(() => {
        Message.success("创建Service成功")
        this.dialogLoading = false
        this.createFormVisible = false
        this.fetchData()
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    handleUpdateService: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      let service = JSON.parse(JSON.stringify(this.service))
      let selector = {}
      for(let s of service.spec.selector) {
        if(!s.key) {
          Message.error("Pod选择器Key不能为空")
          return
        }
        selector[s.key] = s.value
      }
      service.spec.selector = selector
      for(let p of service.spec.ports) {
        try{
          p['port'] = parseInt(p.port)
          if(p.targetPort) p['targetPort'] = parseInt(p.targetPort)
          if(p.nodePort) p['nodePort'] = parseInt(p.nodePort)
        } catch(e) {
          Message.error("端口错误")
          return
        }
      }
      let yamlStr = yaml.dump(service)
      this.dialogLoading = true
      updateResource(cluster, ResType.Service, this.service.metadata.namespace, this.service.metadata.name, yamlStr, {project_id: this.projectId}).then(() => {
        this.dialogLoading = false
        this.createFormVisible = false
        Message.success("编辑Service成功")
        this.fetchData()
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    _delServicesFunc: function() {
      if (this.delServices.length > 0){
        let delServices = []
        for (var p of this.delServices) {
          delServices.push({namespace: p.namespace, name: p.name})
        }
        this.handleDeleteServices(delServices)
      }
    },
    handleSelectionChange(val) {
      this.delServices = val;
      if (val.length > 0){
        this.delFunc = this._delServicesFunc
      } else {
        this.delFunc = undefined
      }
    },
    getPortsDisplay(ports) {
      if (!ports) return ''
      var pd = []
      for (let p of ports) {
        var pds = p.port
        if (p.nodePort) {
          pds += ':' + p.nodePort
        }
        pds += '/' + p.protocol
        pd.push(pds)
      }
      return pd.join(',')
    },
    fetchNamespace: function() {
      this.namespaces = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listResource(cluster, ResType.Namespace).then(response => {
          this.namespaces = response.data
          this.namespaces.sort((a, b) => {return a.name > b.name ? 1 : -1})
        }).catch((err) => {
          console.log(err)
        })
      } else {
        Message.error("获取集群异常，请刷新重试")
      }
    },
    openCreateFormDialog() {
      if(!this.namespace && this.namespaces.length == 0) {
        this.fetchNamespace()
      }
      this.createFormVisible = true
    },
    openUpdateFormDialog(namespace, name) {
      this.createFormVisible = true
      this.updateFormVisible = true
      if(!this.namespace && this.namespaces.length == 0) {
        this.fetchNamespace()
      }
      this.getService(namespace, name)
    },
    closeFormDialog() {
      this.createFormVisible = false
      this.updateFormVisible = false
      this.service = {
        kind: "Service",
        apiVersion: "v1",
        metadata: {
          name: "",
        },
        spec: {
          ports: [],
          selector: [{}],
          type: 'ClusterIP',
        }
      }
    }
  }
}
</script>

<style lang="scss" scoped>

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
