<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="projectId ? undefined : nsSearch" 
     :nameFunc="nameSearch" :createFunc="openCreateFormDialog" />
    <div class="dashboard-container">
      <el-table
        ref="multipleTable"
        :data="persistentVolumeClaim"
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
            <span>
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="40"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="capacity"
          label="容量"
          min-width="25"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ scope.row.capacity }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="access_modes"
          label="访问模式"
          min-width="45"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="storage_class"
          label="存储类"
          min-width="40"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="status"
          label="状态"
          min-width="35"
          show-overflow-tooltip>
        </el-table-column>          
        <el-table-column
          prop="create_time"
          label="创建时间"
          min-width="45"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ $dateFormat(scope.row.create_time) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" show-overflow-tooltip width="110px">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 10px" @click="openUpdateFormDialog(scope.row.namespace, scope.row.name)">编辑</el-link>
              <el-link :disabled="!$editorRole()" :underline="false" type="danger" @click="handleDeletePvcs([{namespace: scope.row.namespace, name: scope.row.name}])">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    
    <el-dialog :title="updateFormVisible ? '编辑PVC' : '创建PVC'" :visible.sync="createFormVisible"
      @close="closeFormDialog" :destroy-on-close="true" width="70%" :close-on-click-modal="false">
      <div v-loading="dialogLoading">
        <div class="dialogContent" style="margin: 0px;">
          <el-form :model="pvc.metadata" :rules="rules" ref="form" label-position="left" label-width="105px">
            <el-form-item label="名称" prop="name" autofocus required>
              <el-input v-model="pvc.metadata.name" style="width: 100%;" autocomplete="off" 
                placeholder="只能包含小写字母数字以及-和.,数字或者字母开头或结尾" size="small" :disabled="updateFormVisible"></el-input>
            </el-form-item>
            <el-form-item label="命名空间" required>
              <span v-if="namespace">{{ namespace }}</span>
              <!-- <el-input v-else :disabled="updateFormVisible" v-model="pvc.metadata.namespace" style="width: 50%;"  autocomplete="off" placeholder="请输入空间描述" size="small"></el-input> -->
              <el-select v-else :disabled="updateFormVisible" v-model="pvc.metadata.namespace" placeholder="请选择命名空间"
                size="small" style="width: 100%;" >
                <el-option
                  v-for="item in namespaces"
                  :key="item.name"
                  :label="item.name"
                  :value="item.name">
                </el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="申请大小" required>
              <el-input v-model="pvc.spec.resources.requests.storage" style="width: 139px;"
                placeholder="存储卷申请大小" size="small">
                <span v-if="!updateFormVisible" slot="suffix" style="padding-right: 5px;" >Gi </span>
              </el-input>
            </el-form-item>
            <el-form-item label="访问模式">
              <el-select v-model="pvc.spec.accessModes" placeholder="访问模式" multiple size="small" style="width: 100%;" >
                <el-option label="ReadWriteOnce" value="ReadWriteOnce"></el-option>
                <el-option label="ReadWriteMany" value="ReadWriteMany"></el-option>
                <el-option label="ReadOnlyMany" value="ReadOnlyMany"></el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="存储类">
              <el-input v-model="pvc.spec.storageClassName" style="width: 100%;" placeholder="存储类" size="small"></el-input>
            </el-form-item>
            <el-form-item label="存储卷">
              <el-input v-model="pvc.spec.volumeName" style="width: 100%;" placeholder="存储卷" size="small"></el-input>
            </el-form-item>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter" style="margin-top: 25px;">
          <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="updateFormVisible ? handleUpdatePvc() : handleCreatePvc()" >确 定</el-button>
        </div>
      </div>
    </el-dialog>

  </div>
</template>

<script>
import { Clusterbar, } from '@/views/components'
import { Message } from 'element-ui'
import { listPersistentVolumeClaim, getPersistentVolumeClaim, deletePersistentVolumeClaims, 
updatePersistentVolumeClaim, buildPvc } from '@/api/persistent_volume_claim'
import { createYaml } from '@/api/cluster'
import { projectLabels } from '@/api/project/project'
import { listNamespace } from '@/api/namespace'
import yaml from 'js-yaml'

export default {
  name: "PersistentVolumeClaim",
  components: {
    Clusterbar,
  },
  data() {
    return {
      titleName: ["PersistentVolumeClaims"],
      originPersistentVolumeClaims: [],
      search_name: "",
      search_ns: [],
      cellStyle: {border: 0},
      maxHeight: window.innerHeight - 150,
      loading: true,
      yamlDialog: false,
      yamlName: "",
      yamlValue: "",
      yamlLoading: true,
      delFunc: undefined,
      delPvcs: [],
      dialogLoading: false,
      createFormVisible: false,
      updateFormVisible: false,
      pvc: {
        apiVersion: "v1",
        kind: "PersistentVolumeClaim",
        metadata: {
          name: "",
        },
        spec: {
          resources: {requests: {}},
          accessModes: [],
        }
      },
      rules: {},
      namespaces: [],
    }
  },
  created() {
    this.fetchData()
  },
  watch: {
    pvcsWatch: function (newObj) {
      if (newObj) {
        let newUid = newObj.resource.metadata.uid
        let newRv = newObj.resource.metadata.resourceVersion
        if (newObj.event === 'add') {
          this.originPersistentVolumeClaims.push(buildPvc(newObj.resource))
        } else if (newObj.event === 'update') {
          for (let i in this.originPersistentVolumeClaims) {
            let d = this.originPersistentVolumeClaims[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newDp = buildPvc(newObj.resource)
                this.$set(this.originPersistentVolumeClaims, i, newDp)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originPersistentVolumeClaims = this.originPersistentVolumeClaims.filter(( { uid } ) => uid !== newUid)
        }
      }
    },
    cluster: function() {
      this.fetchData()
    }
  },
  computed: {
    persistentVolumeClaim: function() {
      let data = []
      for (let c of this.originPersistentVolumeClaims) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(c.namespace) < 0) continue
        if (this.search_name && !c.name.includes(this.search_name)) continue
        data.push(c)
      }
      return data
    },
    pvcsWatch: function() {
      return this.$store.getters["ws/pvcsWatch"]
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
      console.log(namespace, name)
      this.$router.push({
        name: 'pvcDetail',
        params: { persistentVolumeClaimName: name, namespace: namespace },
      })
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
    fetchData: function() {
      this.loading = true
      const cluster = this.$store.state.cluster
      let params = {namespace: this.namespace}
      if(this.projectId) params['labels'] = projectLabels()
      if (cluster) {
        listPersistentVolumeClaim(cluster, params).then(response => {
          this.loading = false
          this.$set(this, 'originPersistentVolumeClaims', response.data ? response.data : [])
        }).catch(() => {
          this.loading = false
        })
      } else if(!this.projectId) {
        this.loading = false
        Message.error("获取集群异常，请刷新重试.")
      }
    },
    getPersistentVolumeClaim: function(namespace, name) {
      this.dialogLoading = true
      getPersistentVolumeClaim(this.cluster, namespace, name, ).then(response => {
        this.pvc = response.data
        this.dialogLoading = false
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    handleUpdatePvc: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      let pvc = JSON.parse(JSON.stringify(this.pvc))
      let yamlStr = yaml.dump(pvc)
      this.dialogLoading = true
      updatePersistentVolumeClaim(cluster, pvc.metadata.namespace, pvc.metadata.name, yamlStr).then(() => {
        Message.success("更新PVC成功")
        this.dialogLoading = false
        this.fetchData()
      }).catch(() => {
        // console.log(e)
        this.dialogLoading = false
      })
    },
    handleDeletePvcs: function(pvcs) {
      let cs = ''
      for(let c of pvcs) {
        cs += `${c.namespace}/${c.name},`
      }
      cs = cs.substr(0, cs.length - 1)
      this.$confirm(`请确认是否删除「${cs}」PVC?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        deletePersistentVolumeClaims(this.cluster, {resources: pvcs}).then(() => {
          Message.success("删除PVC成功")
          this.loading = false
          this.fetchData()
        }).catch((err) => {
          this.loading = false
        });
      }).catch(() => {       
      });
    },
    _delPvcsFunc: function() {
      if (this.delPvcs.length > 0){
        let delPvcs = []
        for (var p of this.delPvcs) {
          delPvcs.push({namespace: p.namespace, name: p.name})
        }
        this.deletePvcs(delPvcs)
      }
    },
    handleSelectionChange(val) {
      this.delPvcs = val;
      if (val.length > 0){
        this.delFunc = this._delPvcsFunc
      } else {
        this.delFunc = undefined
      }
    },
    fetchNamespace: function() {
      this.namespaces = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listNamespace(cluster).then(response => {
          this.namespaces = response.data
          this.namespaces.sort((a, b) => {return a.name > b.name ? 1 : -1})
        }).catch((err) => {
          console.log(err)
        })
      } else {
        Message.error("获取集群异常，请刷新重试")
      }
    },
    handleCreatePvc() {
      let pvc = JSON.parse(JSON.stringify(this.pvc))
      if(!pvc.metadata.name) {
        Message.error("请输入名称")
        return
      }
      if(this.namespace){
        pvc.metadata.namespace = this.namespace
      }
      if(!pvc.metadata.namespace) {
        Message.error("命名空间不能为空")
        return
      }
      if(pvc.spec.resources.requests.storage) {
        pvc.spec.resources.requests.storage += 'Gi'
      }
      if(pvc.spec.resources.limits && pvc.spec.resources.limits.storage) {
        pvc.spec.resources.limits.storage += 'Gi'
      }
      if(this.projectId) {
        pvc.metadata.labels = projectLabels()
      }
      let yamlStr = yaml.dump(pvc)
      this.dialogLoading = true
      createYaml(this.cluster, yamlStr).then((response) => {
        this.dialogLoading = false
        this.createFormVisible = false
        Message.success("创建PVC成功")
        this.fetchData()
      }).catch(() => {
        this.dialogLoading = false
      })
    },
    openCreateFormDialog() {
      if(this.namespaces.length == 0) {
        this.fetchNamespace()
      }
      this.createFormVisible = true
    },
    openUpdateFormDialog(namespace, name) {
      this.createFormVisible = true
      this.updateFormVisible = true
      if(this.namespaces.length == 0) {
        this.fetchNamespace()
      }
      this.getPersistentVolumeClaim(namespace, name)
    },
    closeFormDialog() {
      this.createFormVisible = false
      this.updateFormVisible = false
      this.pvc = {
        apiVersion: "v1",
        kind: "PersistentVolumeClaim",
        metadata: {
          name: "",
        },
        spec: {
          resources: {requests: {}},
          accessModes: [],
        }
      }
    },
    storeageUnit(res) {
      if(!res) return 'Gi'
      if(res.indexOf('G') > -1) return 'Gi'
      if(res.indexOf('T') > -1) return 'Ti'
      if(res.indexOf('M') > -1) return 'Mi'
      return 'Gi'
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
</style>