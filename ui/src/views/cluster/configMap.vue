<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="projectId ? undefined : nsSearch" 
      :nameFunc="nameSearch" :createFunc="openCreateFormDialog"/>
    <div class="dashboard-container">
      <el-table
        ref="multipleTable"
        :data="configMaps"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort="{ prop: 'name' }"
      >
        <el-table-column
          prop="name"
          label="名称"
          min-width="60"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            <span>
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="40"
          show-overflow-tooltip
        >
        </el-table-column>
        <el-table-column
          prop="keys"
          label="配置项"
          min-width="90"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            <template v-for="(v, k) in scope.row.data">
              <el-tooltip :key="k" class="item" effect="light" placement="right-end">
                <div slot="content" style="max-width: 400px;white-space: pre-wrap;">
                  {{ v }}
                </div>
                <span class="back-class">
                  {{ k }}
                </span>
              </el-tooltip>
            </template>
          </template>
        </el-table-column>
        <el-table-column
          prop="create_time"
          label="创建时间"
          min-width="45"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.create_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" show-overflow-tooltip width="110px">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 15px;" @click="openUpdateFormDialog(scope.row.namespace, scope.row.name)">编辑</el-link>
              <el-link :disabled="!$editorRole()" :underline="false" type="danger" @click="handleDeleteConfigMap([{namespace: scope.row.namespace, name: scope.row.name}])">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog :title="updateFormVisible ? '编辑ConfigMap' : '创建ConfigMap'" :visible.sync="createFormVisible"
      @close="closeFormDialog" :destroy-on-close="true" width="70%" :close-on-click-modal="false">
      <div v-loading="dialogLoading">
        <div class="dialogContent" style="margin: 0px;">
          <el-form :model="configMap.metadata" :rules="rules" ref="form" label-position="left" label-width="105px">
            <el-form-item label="名称" prop="name" autofocus required>
              <el-input v-model="configMap.metadata.name" style="width: 50%;" autocomplete="off" 
                placeholder="只能包含小写字母数字以及-和.,数字或者字母开头或结尾" size="small" :disabled="updateFormVisible"></el-input>
            </el-form-item>
            <el-form-item label="命名空间" prop="description">
              <span v-if="namespace">{{ namespace }}</span>
              <!-- <el-input v-else :disabled="updateFormVisible" v-model="configMap.metadata.namespace" style="width: 50%;"  autocomplete="off" placeholder="请输入空间描述" size="small"></el-input> -->
              <el-select v-else :disabled="updateFormVisible" v-model="configMap.metadata.namespace" placeholder="请选择命名空间"
                size="small" style="width: 50%;" >
                <el-option
                  v-for="item in namespaces"
                  :key="item.name"
                  :label="item.name"
                  :value="item.name">
                </el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="配置项" prop="" :required="true">
              <el-row style="margin-bottom: 5px; margin-top: 8px;">
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
              <el-row style="padding-top: 0px;" v-for="(d, i) in configMap.data" :key="i">
                <el-col :span="11">
                  <div class="border-span-header">
                    <el-input v-model="d.key" size="small" style="padding-right: 10px" placeholder="配置项Key"></el-input>
                  </div>
                </el-col>
                <el-col :span="12">
                  <div class="border-span-header">
                    <el-input type="textarea" style="border-radius: 0px; margin-bottom: 5px;" v-model="d.value" size="small" placeholder="配置项Value"></el-input>
                  </div>
                </el-col>
                <el-col :span="1" style="padding-left: 10px">
                  <el-button circle size="mini" style="padding: 5px;" 
                    @click="configMap.data.splice(i, 1)" icon="el-icon-close"></el-button>
                </el-col>
              </el-row>
              <el-row>
                <el-col :span="23">
                <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px; border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
                  @click="configMap.data.push({})" icon="el-icon-plus">添加配置项</el-button>
                </el-col>
              </el-row>
            </el-form-item>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter" style="margin-top: 25px;">
          <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="updateFormVisible ? handleUpdateConfigMap() : handleCreateConfigMap()" >确 定</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { Message } from 'element-ui'
import { ResType, listResource, getResource, delResource, updateResource, createResource } from '@/api/cluster/resource'
import { projectLabels } from '@/api/project/project'
import yaml from 'js-yaml'

export default {
  name: 'ConfigMap',
  components: {
    Clusterbar,
  },
  data() {
    return {
      titleName: ['ConfigMap'],
      originConfigMaps: [],
      search_name: '',
      search_ns: [],
      cellStyle: { border: 0 },
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      dialogLoading: false,
      createFormVisible: false,
      updateFormVisible: false,
      configMap: {
        apiVersion: "v1",
        kind: "ConfigMap",
        metadata: {
          name: "",
        },
        data: []
      },
      rules: {},
      namespaces: [],
    }
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
  created() {
    this.fetchData()
  },
  watch: {
    cluster: function() {
      this.originConfigMaps = []
      this.fetchData()
    }
  },
  computed: {
    configMaps: function() {
      let data = []
      for (let c of this.originConfigMaps) {
        if (
          this.search_ns.length > 0 &&
          this.search_ns.indexOf(c.namespace) < 0
        )
          continue
        if (this.search_name && !c.name.includes(this.search_name)) continue
        var str = ''
        for (let s of c.keys) {
          str += s + ','
        }
        if (str.length > 0) {
          str = str.substr(0, str.length - 1)
        }
        // c['keys'] = str
        data.push(c)
      }
      return data
    },
    projectId() {
      return this.$route.params.workspaceId
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    namespace: function() {
      return this.$store.state.namespace
    }
  },
  methods: {
    nameClick: function(namespace, name) {
      this.$router.push({
        name: 'configMapDetail',
        params: { namespace: namespace, configMapName: name },
      })
    },
    nsSearch: function(vals) {
      this.search_ns = []
      for (let ns of vals) {
        this.search_ns.push(ns)
      }
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    fetchData: function() {
      this.loading = true
      const cluster = this.$store.state.cluster
      let params = {namespace: this.namespace}
      if(this.projectId) params['label_selector'] = {"matchLabels": projectLabels()}
      if (cluster) {
        listResource(cluster, ResType.ConfigMap, params, {project_id: this.projectId})
          .then((response) => {
            this.loading = false
            let originConfigMaps = response.data || []
            this.$set(this, 'originConfigMaps', originConfigMaps)
          })
          .catch(() => {
            this.loading = false
          })
      } else if(!this.projectId) {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
    handleCreateConfigMap() {
      let configMap = JSON.parse(JSON.stringify(this.configMap))
      if(!configMap.metadata.name) {
        Message.error("请输入名称")
        return
      }
      if(configMap.data.length == 0) {
        Message.error("请添加配置项")
        return
      }
      let data = {}
      for(let d of configMap.data) {
        if(!d.key) {
          Message.error("配置项Key不能为空")
          return
        }
        data[d.key] = d.value || ''
      }
      configMap.data = data
      if(this.namespace){
        configMap.metadata.namespace = this.namespace
      }
      if(!configMap.metadata.namespace) {
        Message.error("命名空间不能为空")
        return
      }
      if(this.projectId) {
        configMap.metadata.labels = projectLabels()
      }
      let yamlStr = yaml.dump(configMap)
      this.dialogLoading = true
      createResource(this.cluster, yamlStr, {project_id: this.projectId}).then((response) => {
        this.dialogLoading = false
        this.createFormVisible = false
        Message.success("创建ConfigMap成功")
        this.fetchData()
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    handleUpdateConfigMap() {
      let configMap = JSON.parse(JSON.stringify(this.configMap))
      let data = {}
      for(let d of configMap.data) {
        if(!d.key) {
          Message.error("配置项Key不能为空")
          return
        }
        data[d.key] = d.value || ''
      }
      configMap.data = data
      let yamlStr = yaml.dump(configMap)
      this.dialogLoading = true
      updateResource(this.cluster, ResType.ConfigMap, configMap.metadata.namespace, configMap.metadata.name, yamlStr, {project_id: this.projectId}).then((response) => {
        this.dialogLoading = false
        this.createFormVisible = false
        Message.success("编辑ConfigMap成功")
        this.fetchData()
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    handleDeleteConfigMap(cms) {
      let cs = ''
      for(let c of cms) {
        cs += `${c.namespace}/${c.name},`
      }
      cs = cs.substr(0, cs.length - 1)
      this.$confirm(`请确认是否删除「${cs}」ConfigMap?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        delResource(this.cluster, ResType.ConfigMap, {resources: cms}, {project_id: this.projectId}).then(() => {
          Message.success("删除ConfigMap成功")
          this.loading = false
          this.fetchData()
        }).catch((err) => {
          this.loading = false
        });
      }).catch(() => {       
      });
    },
    getConfigMap: function(namespace, name) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error('获取集群参数异常，请刷新重试')
        return
      }
      if (!namespace) {
        Message.error('获取命名空间参数异常，请刷新重试')
        return
      }
      if (!name) {
        Message.error('获取ConfigMap名称参数异常，请刷新重试')
        return
      }
      this.dialogLoading = true
      getResource(cluster, ResType.ConfigMap, namespace, name, '', {project_id:this.projectId}).then((response) => {
        this.dialogLoading = false
        this.configMap = response.data
        let data = []
        for(let k in this.configMap.data) {
          data.push({key: k, value: this.configMap.data[k]})
        }
        this.configMap.data = data
      }).catch(() => {
        this.dialogLoading = false
      })
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
      if(this.namespaces.length == 0 && !this.namespace) {
        this.fetchNamespace()
      }
      this.configMap = {
        apiVersion: "v1",
        kind: "ConfigMap",
        metadata: {
          name: "",
        },
        data: []
      }
      this.createFormVisible = true
    },
    openUpdateFormDialog(namespace, name) {
      this.createFormVisible = true
      this.updateFormVisible = true
      if(this.namespaces.length == 0 && !this.namespace) {
        this.fetchNamespace()
      }
      this.getConfigMap(namespace, name)
    },
    closeFormDialog() {
      this.createFormVisible = false
      this.updateFormVisible = false
    }
  },
}
</script>
<style lang="scss" scoped>
  
  .name-class {
    cursor: pointer;
  }
  .name-class:hover {
    color: #409EFF;
  }
  </style>