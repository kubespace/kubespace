<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="projectId ? undefined : nsSearch" :nameFunc="nameSearch" 
      :createFunc="openCreateFormDialog"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="ingresses"
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
          min-width="50"
          show-overflow-tooltip>
          <!-- <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.namespace, scope.row.name)">
              {{ scope.row.name }}
            </span>
          </template> -->
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="40"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="rules"
          label="主机名"
          min-width="75"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ getIngressHosts(scope.row.rules) }}
            </span>
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
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 10px;" @click="openUpdateFormDialog(scope.row.namespace, scope.row.name)">编辑</el-link>
              <el-link :disabled="!$editorRole()" :underline="false" type="danger" @click="handleDeleteIngresses([{namespace: scope.row.namespace, name: scope.row.name}])">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog :title="updateFormVisible ? '编辑Ingress' : '创建Ingress'" :visible.sync="createFormVisible"
      @close="closeFormDialog" :destroy-on-close="true" width="70%" :close-on-click-modal="false" top="8vh">
      <div v-loading="dialogLoading">
        <div class="dialogContent" style="margin: 0px;">
          <el-form :model="ingress.metadata" :rules="rules" ref="form" label-position="left" label-width="105px">
            <el-form-item label="名称" prop="name" autofocus required>
              <el-input v-model="ingress.metadata.name" style="width: 50%;" autocomplete="off" 
                placeholder="只能包含小写字母数字以及-和.,数字或者字母开头或结尾" size="small" :disabled="updateFormVisible"></el-input>
            </el-form-item>
            <el-form-item label="命名空间" required>
              <span v-if="namespace">{{ namespace }}</span>
              <!-- <el-input v-else :disabled="updateFormVisible" v-model="ingress.metadata.namespace" style="width: 50%;"  autocomplete="off" placeholder="请输入空间描述" size="small"></el-input> -->
              <el-select v-else :disabled="updateFormVisible" v-model="ingress.metadata.namespace" placeholder="请选择命名空间"
                size="small" style="width: 50%;" >
                <el-option
                  v-for="item in namespaces"
                  :key="item.name"
                  :label="item.name"
                  :value="item.name">
                </el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="Annotations" prop="" >
              <el-row style="margin-bottom: 5px; margin-top: 2px;">
                <el-col :span="11" style="background-color: #F5F7FA; padding-left: 10px;">
                  <div class="border-span-header">
                    <span  class="border-span-content">*</span>Key
                  </div>
                </el-col>
                <el-col :span="12" style="background-color: #F5F7FA">
                  <div class="border-span-header">
                    Value
                  </div>
                </el-col>
                <!-- <el-col :span="5"><div style="width: 100px;"></div></el-col> -->
              </el-row>
              <el-row style="padding-top: 0px;" v-for="(d, i) in ingress.metadata.annotations" :key="i">
                <el-col :span="11">
                  <div class="border-span-header">
                    <el-input v-model="d.key" size="small" style="padding-right: 10px" placeholder="Key"></el-input>
                  </div>
                </el-col>
                <el-col :span="12">
                  <div class="border-span-header">
                    <el-input v-model="d.value" size="small" placeholder="Value"></el-input>
                  </div>
                </el-col>
                <el-col :span="1" style="padding-left: 10px">
                  <el-button circle size="mini" style="padding: 5px;" 
                    @click="ingress.metadata.annotations.splice(i, 1)" icon="el-icon-close"></el-button>
                </el-col>
              </el-row>
              <el-row>
                <el-col :span="23">
                <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px; border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
                  @click="ingress.metadata.annotations.push({})" icon="el-icon-plus">添加注解</el-button>
                </el-col>
              </el-row>
            </el-form-item>
            <el-form-item label="路由规则" prop="" :required="true">
              <el-card style="margin-bottom: 10px;" v-for="(r, idx) in ingress.spec.rules" :key="idx">
                <el-row>
                  <el-col :span="20">
                    <el-form label-width="60px">
                      <el-form-item label="Host" prop="name" autofocus>
                        <el-input v-model="r.host" style="width: 50%;" autocomplete="off" 
                          placeholder="请输入Host主机名，如kubespace.cn" size="small"></el-input>
                      </el-form-item>
                    </el-form>
                  </el-col>
                  <el-col :span="4">
                    <div style="float: right;">
                      <el-button style="float: right; padding: 3px 0" type="text"
                        @click="ingress.spec.rules.splice(idx, 1)">删除</el-button>
                    </div>
                  </el-col>
                </el-row>
                
                <el-row style="margin-bottom: 5px; margin-top: 0px;">
                  <el-col :span="6" style="background-color: #F5F7FA; padding-left: 10px;">
                    <div class="border-span-header">
                      路径
                    </div>
                  </el-col>
                  <el-col :span="6" style="background-color: #F5F7FA">
                    <div class="border-span-header">
                      路径类型
                    </div>
                  </el-col>
                  <el-col :span="6" style="background-color: #F5F7FA">
                    <div class="border-span-header">
                      后端服务
                    </div>
                  </el-col>
                  <el-col :span="5" style="background-color: #F5F7FA">
                    <div class="border-span-header">
                      服务端口
                    </div>
                  </el-col>
                  <!-- <el-col :span="5"><div style="width: 100px;"></div></el-col> -->
                </el-row>
                <el-row style="padding-top: 0px;" v-for="(d, i) in r.http.paths" :key="i">
                  <el-col :span="6">
                    <div class="border-span-header">
                      <el-input v-model="d.path" size="small" style="padding-right: 10px" placeholder="后端路径"></el-input>
                    </div>
                  </el-col>
                  <el-col :span="6">
                    <div class="border-span-header">
                      <el-select v-model="d.pathType" placeholder="路径类型" @change="serviceChange(d)"
                        size="small" style="width: 100%; padding-right: 10px" >
                        <el-option value="ImplementationSpecific" label="ImplementationSpecific"></el-option>
                        <el-option value="Exact" label="Exact"></el-option>
                        <el-option value="Prefix" label="Prefix"></el-option>
                      </el-select>
                    </div>
                  </el-col>
                  <el-col :span="6">
                    <div class="border-span-header">
                      <el-select v-model="d.service" placeholder="后端服务" @change="serviceChange(d)"
                        size="small" style="width: 100%; padding-right: 10px" clearable allow-create filterable>
                        <el-option
                          v-for="item in namespaceService"
                          :key="item.name"
                          :label="item.name"
                          :value="item.name">
                        </el-option>
                      </el-select>
                    </div>
                  </el-col>
                  <el-col :span="5">
                    <div class="border-span-header">
                      <el-select v-model="d.port" placeholder="服务端口"
                        size="small" style="width: 100%;" clearable allow-create filterable>
                        <el-option
                          v-for="item in d.ports || []"
                          :key="item.port"
                          :label="item.port"
                          :value="item.port">
                        </el-option>
                      </el-select>
                    </div>
                  </el-col>
                  <el-col :span="1" style="padding-left: 10px">
                    <el-button circle size="mini" style="padding: 5px;" 
                      @click="r.http.paths.splice(i, 1)" icon="el-icon-close"></el-button>
                  </el-col>
                </el-row>
                <el-row>
                  <el-col :span="23">
                    <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px; border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
                      @click="r.http.paths.push({})" icon="el-icon-plus">添加路径</el-button>
                  </el-col>
                </el-row>
              </el-card>
              <el-button style="width: 50%; border-radius: 0px; padding: 9px 15px; border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
                @click="ingress.spec.rules.push({host: '', http: {paths: []}})" icon="el-icon-plus">添加规则</el-button>
            </el-form-item>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter" style="margin-top: 25px;">
          <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="updateFormVisible ? handleUpdateIngress() : handleCreateIngress()" >确 定</el-button>
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
  name: 'Ingress',
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
      titleName: ["Ingresses"],
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      originIngresses: [],
      search_ns: [],
      search_name: '',
      delFunc: undefined,
      delIngresses: [],
      dialogLoading: false,
      createFormVisible: false,
      updateFormVisible: false,
      ingressGroup: "extensions",
      ingress: {
        kind: "Ingress",
        apiVersion: "extensions/v1beta1",
        metadata: {
          name: "",
          annotations: [],
        },
        spec: {
          backend: {},
          rules: [],
          tls: [],
        },
      },
      rules: {},
      namespaces: [],
      services: [],
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
    ingresses: function() {
      let dlist = []
      for (let p of this.originIngresses) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        if (p.conditions && p.conditions.length > 0) {
          p.conditions.sort()
        } else {
          p.conditions = []
        }
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
    },
    namespaceService: function() {
      let namespace = this.namespace
      if (namespace == '') {
        namespace = this.ingress.metadata.namespace
      }
      if(!namespace) return []
      let services = []
      for(let s of this.services) {
        if(s.namespace == namespace) {
          services.push(s)
        }
      }
      return services
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originIngresses = []
      const cluster = this.$store.state.cluster
      let params = {namespace: this.namespace}
      if(this.projectId) params['label_selector'] = {"matchLabels": projectLabels()}
      if (cluster) {
        listResource(cluster, ResType.Ingress, params).then(response => {
          this.loading = false
          let originIngresses = response.data.ingresses ? response.data.ingresses : []
          this.$set(this, 'originIngresses', originIngresses)
          this.ingressGroup = response.data.group
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
    buildIngresses: function(ingress) {
      if (!ingress) return
      let p = {
        uid: ingress.metadata.uid,
        namespace: ingress.metadata.namespace,
        name: ingress.metadata.name,
        backend: ingress.spec.backend,
        tls: ingress.spec.tls,
        rules: ingress.spec.rules,
        resource_version: ingress.metadata.resourceVersion,
        created: ingress.metadata.creationTimestamp
      }
      return p
    },
    nameClick: function(namespace, name) {
      this.$router.push({name: 'ingressDetail', params: {namespace: namespace, ingressName: name}})
    },
    getIngress: function(namespace, name) {
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
      this.dialogLoading = true
      getResource(cluster, ResType.Ingress, namespace, name, ).then(response => {
        let ingress = response.data
        // let rules = []
        for(let r of ingress.spec.rules) {
          for(let p of r.http.paths) {
            if(p.backend){
              if(this.ingressGroup == 'extensions') {
                p['service'] = p.backend.serviceName
                p['port'] = p.backend.servicePort
              } else if(p.backend.service) {
                p['service'] = p.backend.service.name
                if(p.backend.service.port) {
                  p['port'] = p.backend.service.port.number || p.backend.service.port.name
                }
              }
            }
          }
        }
        let annotations = []
        if(ingress.metadata.annotations){
          for(let k in ingress.metadata.annotations) {
            annotations.push({key: k, value: ingress.metadata.annotations[k]})
          }
        }
        ingress.metadata.annotations = annotations
        this.ingress = ingress
        this.dialogLoading = false
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    handleDeleteIngresses: function(ingresses) {
      let cs = ''
      for(let c of ingresses) {
        cs += `${c.namespace}/${c.name},`
      }
      cs = cs.substr(0, cs.length - 1)
      this.$confirm(`请确认是否删除「${cs}」Ingress?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        delResource(this.cluster, ResType.Ingress, {resources: ingresses}, {project_id: this.projectId}).then(() => {
          Message.success("删除Ingress成功")
          this.fetchData()
        }).catch(() => {
          // console.log(e)
        })
      }).catch(() => {       
      });
    },
    handleUpdateIngress: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      let ingress = JSON.parse(JSON.stringify(this.ingress))
      let err = this.transferIngressRules(ingress)
      if(err) {
        Message.error(err)
        return
      }
      let yamlStr = yaml.dump(ingress)
      this.dialogLoading = true
      updateResource(cluster, ResType.Ingress, ingress.metadata.namespace, ingress.metadata.name, yamlStr, {project_id: this.projectId}).then(() => {
        Message.success("更新Ingress成功")
        this.dialogLoading = false
        this.createFormVisible = false
        this.fetchData()
      }).catch(() => {
        // console.log(e) 
        this.dialogLoading = false
      })
    },
    _delIngressesFunc: function() {
      if (this.delIngresses.length > 0){
        let delIngresses = []
        for (var p of this.delIngresses) {
          delIngresses.push({namespace: p.namespace, name: p.name})
        }
        this.deleteIngresses(delIngresses)
      }
    },
    handleSelectionChange(val) {
      this.delIngresses = val;
      if (val.length > 0){
        this.delFunc = this._delIngressesFunc
      } else {
        this.delFunc = undefined
      }
    },
    getIngressHosts(rules) {
      if (!rules) return ''
      let hosts = []
      for(let r of rules) {
        hosts.push(r.host)
      }
      return hosts.join(',')
    },
    transferIngressRules(ingress) {
      let rules = []
      for(let r of ingress.spec.rules) {
        let rule = {host: r.host, http: {paths: []}}
        for(let p of r.http.paths) {
          let path = {
            path: p.path
          }
          let port = p.port
          try{
            port = parseInt(p.port)
          }catch(e) {
            port = p.port
          }
          if(p.pathType) path['pathType'] = p.pathType
          if(this.ingressGroup == 'extensions') {
            path['backend'] = {serviceName: p.service}
            if(port) path['backend']['servicePort'] = port
          } else {
            path['backend'] = {service: {name: p.service}}
            if(port) {
              if(isNaN(port)){
                path['backend']['service']['port'] = {'name': port}
              } else {
                path['backend']['service']['port'] = {'number': port}
              }
            }
            if(!p.pathType) path['pathType'] = 'ImplementationSpecific'
          }
          rule.http.paths.push(path)
        }
        rules.push(rule)
      }
      ingress.spec.rules = rules
      
      let annotations = {}
      for(let a of ingress.metadata.annotations) {
        if(!a.key) {
          return '注解Key不能为空'
        }
        annotations[a.key] = a.value || ''
      }
      ingress.metadata.annotations = annotations
    },
    handleCreateIngress: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      let ingress = JSON.parse(JSON.stringify(this.ingress))
      if(this.namespace){
        ingress.metadata.namespace = this.namespace
      }
      if(!ingress.metadata.namespace) {
        Message.error("命名空间不能为空")
        return
      }
      if(this.projectId) {
        ingress.metadata.labels = projectLabels()
      }
      let err = this.transferIngressRules(ingress)
      if(err) {
        Message.error(err)
        return
      }
      // console.log(this.ingressGroup)
      if(this.ingressGroup == 'networking.k8s.io') {
        ingress.apiVersion = 'networking.k8s.io/v1'
      }
      
      let yamlStr = yaml.dump(ingress)
      this.dialogLoading = true
      createResource(cluster, yamlStr, {project_id: this.projectId}).then(() => {
        Message.success("创建Ingress成功")
        this.dialogLoading = false
        this.createFormVisible = false
        this.fetchData()
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    fetchServices: function() {
      let params = {namespace: this.namespace}
      if (this.cluster) {
        listResource(this.cluster, ResType.Service, params).then(response => {
          this.services = response.data || []
        }).catch(() => {
        })
      } else if(!this.projectId) {
        Message.error("获取集群异常，请刷新重试")
      }
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
      if(this.services.length == 0) {
        this.fetchServices()
      }
      this.createFormVisible = true
    },
    openUpdateFormDialog(namespace, name) {
      this.createFormVisible = true
      this.updateFormVisible = true
      if(!this.namespace && this.namespaces.length == 0) {
        this.fetchNamespace()
      }
      if(this.services.length == 0) {
        this.fetchServices()
      }
      this.getIngress(namespace, name)
    },
    closeFormDialog() {
      this.createFormVisible = false
      this.updateFormVisible = false
      this.ingress = {
        kind: "Ingress",
        apiVersion: "extensions/v1beta1",
        metadata: {
          name: "",
          annotations: [],
        },
        spec: {
          backend: {},
          rules: [],
          tls: [],
        },
      }
    },
    serviceChange(path) {
      let service = {}
      for(let s of this.namespaceService) {
        if(s.name == path.service) {
          service = s
          break
        }
      }
      path.ports = service.ports
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
