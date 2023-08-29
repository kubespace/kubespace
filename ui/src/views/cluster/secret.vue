<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="projectId ? undefined : nsSearch" 
      :nameFunc="nameSearch" :createFunc="openCreateFormDialog"/>
    <div class="dashboard-container">
      <el-table
        ref="multipleTable"
        :data="secrets"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort="{ prop: 'name' }"
        row-key="uid"
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
          prop="type"
          label="类型"
          min-width="50"
          show-overflow-tooltip
        >
        </el-table-column>
        <el-table-column
          prop="keys"
          label="配置项"
          min-width="80"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            <template v-for="(v, k) in scope.row.data">
              <el-tooltip :key="k" class="item" effect="light" placement="right-end">
                <div slot="content" style="max-width: 400px;white-space: pre-wrap;">
                  {{ decodeBase(v) }}
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
              <el-link :disabled="!$editorRole()" :underline="false" type="danger" @click="handleDeleteSecret([{namespace: scope.row.namespace, name: scope.row.name}])">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog :title="updateFormVisible ? '编辑Secret' : '创建Secret'" :visible.sync="createFormVisible"
      @close="closeFormDialog" :destroy-on-close="true" width="70%" :close-on-click-modal="false">
      <div  v-loading="dialogLoading">
      <div class="dialogContent" style="">
        <el-form :model="secret.metadata" :rules="rules" ref="form" label-position="left" label-width="105px">
          <el-form-item label="名称" prop="name" autofocus required>
            <el-input v-model="secret.metadata.name" style="width: 50%;" autocomplete="off" 
              placeholder="只能包含小写字母数字以及-和.,数字或者字母开头或结尾" size="small" :disabled="updateFormVisible"></el-input>
          </el-form-item>
          <el-form-item label="命名空间" prop="description">
            <span v-if="namespace">{{ namespace }}</span>
            <!-- <el-input v-else :disabled="updateFormVisible" v-model="secret.metadata.namespace" style="width: 50%;"  autocomplete="off" placeholder="请输入空间描述" size="small"></el-input> -->
            <el-select v-else :disabled="updateFormVisible" v-model="secret.metadata.namespace" placeholder="请选择命名空间"
              size="small" style="width: 50%;" >
              <el-option
                v-for="item in namespaces"
                :key="item.name"
                :label="item.name"
                :value="item.name">
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="类型" style="width: 100%" prop="name">
            <el-radio-group v-model="secret.type"  size="small">
              <el-radio-button v-if="updateFormVisible" :label="secret.type">{{ secret.type }}</el-radio-button>
              <template v-else>
                <el-radio-button label="Opaque">Opaque</el-radio-button>
                <el-radio-button label="kubernetes.io/tls">TLS</el-radio-button>
                <el-radio-button label="kubernetes.io/basic-auth">用户名/密码</el-radio-button>
                <el-radio-button label="kubernetes.io/dockerconfigjson">镜像服务</el-radio-button>
              </template>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="配置项" prop="" :required="true" v-if="secret.type=='Opaque' || secretTypes.indexOf(secret.type) == -1">
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
            <el-row style="padding-top: 0px;" v-for="(d, i) in secret.data" :key="i">
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
                  @click="secret.data.splice(i, 1)" icon="el-icon-close"></el-button>
              </el-col>
            </el-row>
            <el-row>
              <el-col :span="23">
              <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px; border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
                @click="secret.data.push({})" icon="el-icon-plus">添加配置项</el-button>
              </el-col>
            </el-row>
          </el-form-item>
          <el-form-item v-if="secret.type == 'kubernetes.io/tls'" label="证书" style="width: 60%" required>
            <el-input type="textarea" v-model="secret.tls.crt" placeholder="" size="small"></el-input>
          </el-form-item>
          <el-form-item v-if="secret.type == 'kubernetes.io/tls'" label="密钥" style="width: 60%" required>
            <el-input type="textarea" v-model="secret.tls.key" placeholder="" size="small"></el-input>
          </el-form-item>
          <el-form-item v-if="secret.type == 'kubernetes.io/basic-auth'" label="用户" style="width: 500px" required>
            <el-input v-model="secret.userPass.username" placeholder="" size="small"></el-input>
          </el-form-item>
          <el-form-item v-if="secret.type == 'kubernetes.io/basic-auth'" label="密码" style="width: 500px;" required>
            <el-input type="password" autocomplete="new-password" v-model="secret.userPass.password" placeholder="" size="small"></el-input>
          </el-form-item>
          <el-form-item v-if="secret.type == 'kubernetes.io/dockerconfigjson'" label="仓库地址" style="width: 500px" required>
            <el-input v-model="secret.imagePass.url" placeholder="镜像仓库地址" size="small"></el-input>
          </el-form-item>
          <el-form-item v-if="secret.type == 'kubernetes.io/dockerconfigjson'" label="用户" style="width: 500px;" required>
            <el-input v-model="secret.imagePass.username" placeholder="镜像仓库认证用户" size="small"></el-input>
          </el-form-item>
          <el-form-item v-if="secret.type == 'kubernetes.io/dockerconfigjson'" label="密码" style="width: 500px" required>
            <el-input type="password" autocomplete="new-password" v-model="secret.imagePass.password" placeholder="镜像仓库认证密码" size="small"></el-input>
          </el-form-item>
          <el-form-item v-if="secret.type == 'kubernetes.io/dockerconfigjson'" label="邮箱" style="width: 500px;" required>
            <el-input v-model="secret.imagePass.email" placeholder="用户邮箱" size="small"></el-input>
          </el-form-item>
        </el-form>
      </div>
      <div slot="footer" class="dialogFooter" style="margin-top: 25px;">
        <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
        <el-button type="primary" @click="updateFormVisible ? handleUpdateSecret() : handleCreateSecret()" >确 定</el-button>
      </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { Message } from 'element-ui'
import { ResType, listResource, getResource, delResource, updateResource, createResource,
         resolveSecret, transferSecret } from '@/api/cluster/resource'
import { projectLabels } from '@/api/project/project'
import yaml from 'js-yaml'
import { Base64 } from 'js-base64';

export default {
  name: 'Secrets',
  components: {
    Clusterbar,
  },
  data() {
    return {
      titleName: ["Secrets"],
      originSecrets: [],
      search_name: '',
      search_ns: [],
      cellStyle: { border: 0 },
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      dialogLoading: false,
      createFormVisible: false,
      updateFormVisible: false,
      secretTypes: ["Opaque", "kubernetes.io/tls", "kubernetes.io/basic-auth", "kubernetes.io/dockerconfigjson"],
      secret: {
        apiVersion: "v1",
        kind: "Secret",
        metadata: {
          name: "",
        },
        data: [],
        type: 'Opaque',
        tls: {},
        userPass: {},
        imagePass: {},
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
    secrets: function() {
      let dlist = []
      for (let p of this.originSecrets) {
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
      return this.$store.state.namespace
    }
  },
  methods: {
    decodeBase(val) {
      try{
        return Base64.decode(val)
      } catch(e) {
        return val
      }
    },
    nameClick: function(namespace, name) {
      this.$router.push({
        name: 'secretDetail',
        params: { namespace: namespace, secretName: name },
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
      // this.originSecrets = []
      const cluster = this.$store.state.cluster
      let params = {namespace: this.namespace}
      if(this.projectId) params['label_selector'] = {"matchLabels": projectLabels()}
      if (cluster) {
        listResource(cluster, ResType.Secret, params, {project_id: this.projectId}).then(response => {
          this.loading = false
          let originSecrets = response.data || []
          this.$set(this, 'originSecrets', originSecrets)
        }).catch(() => {
          this.loading = false
        })
      } else if(!this.projectId) {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
    handleCreateSecret() {
      let secret = JSON.parse(JSON.stringify(this.secret))
      if(!secret.metadata.name) {
        Message.error("请输入名称")
        return
      }
      let err = transferSecret(secret)
      if(err) {
        Message.error(err)
        return
      }
      if(this.namespace){
        secret.metadata.namespace = this.namespace
      }
      if(!secret.metadata.namespace) {
        Message.error("命名空间不能为空")
        return
      }
      if(this.projectId) {
        secret.metadata.labels = projectLabels()
      }
      let yamlStr = yaml.dump(secret)
      this.dialogLoading = true
      createResource(this.cluster, yamlStr, {project_id: this.projectId}).then((response) => {
        this.dialogLoading = false
        this.createFormVisible = false
        Message.success("创建Secret成功")
        this.fetchData()
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    handleUpdateSecret() {
      let secret = JSON.parse(JSON.stringify(this.secret))
      let err = transferSecret(secret)
      if(err) {
        Message.error(err)
        return
      }
      let yamlStr = yaml.dump(secret)
      this.dialogLoading = true
      updateResource(this.cluster, ResType.Secret, this.secret.metadata.namespace, this.secret.metadata.name, yamlStr, {project_id: this.projectId}).then((response) => {
        this.dialogLoading = false
        this.createFormVisible = false
        Message.success("编辑Secret成功")
        this.fetchData()
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    handleDeleteSecret(cms) {
      let cs = ''
      for(let c of cms) {
        cs += `${c.namespace}/${c.name},`
      }
      cs = cs.substr(0, cs.length - 1)
      this.$confirm(`请确认是否删除「${cs}」Secret?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        delResource(this.cluster, ResType.Secret, {resources: cms}, {project_id: this.projectId}).then(() => {
          Message.success("删除Secret成功")
          this.loading = false
          this.fetchData()
        }).catch((err) => {
          this.loading = false
        });
      }).catch(() => {       
      });
    },
    getSecret: function(namespace, name) {
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
        Message.error('获取Secret名称参数异常，请刷新重试')
        return
      }
      this.dialogLoading = true
      getResource(cluster, ResType.Secret, namespace, name, '', {project_id: this.projectId}).then((response) => {
        this.dialogLoading = false
        let secret = response.data
        resolveSecret(secret)
        this.secret = secret
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
      this.getSecret(namespace, name)
    },
    closeFormDialog() {
      this.createFormVisible = false
      this.updateFormVisible = false
      this.secret = {
        apiVersion: "v1",
        kind: "Secret",
        metadata: {
          name: "",
        },
        data: [],
        type: 'Opaque',
        tls: {},
        userPass: {},
        imagePass: {},
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
  </style>