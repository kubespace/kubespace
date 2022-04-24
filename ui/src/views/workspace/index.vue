<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="openCreateFormDialog" createDisplay="创建空间"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="originWorkspaces"
        class="table-fix"
        :cell-style="cellStyle"
        v-loading="loading"
        :default-sort = "{prop: 'name'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column prop="name" label="名称" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.id)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="cluster_id" label="绑定集群" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ scope.row.cluster ? scope.row.cluster.name1 : scope.row.cluster_id }}
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="owner" label="负责人" show-overflow-tooltip min-width="10">
          <template slot-scope="scope">
            {{ scope.row.owner }}
          </template>
        </el-table-column>
        <el-table-column prop="update_time" label="更新时间" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.update_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :underline="false" style="margin-right: 13px; color:#409EFF" @click="nameClick(scope.row.id)">详情</el-link>
              <el-link :disabled="!$editorRole(scope.row.id)" :underline="false" type="primary" style="margin-right: 13px" @click="openUpdateFormDialog(scope.row)">编辑</el-link>
              <el-link :disabled="!$adminRole(scope.row.id)" :underline="false" type="danger" @click="handleDeleteWorkspace(scope.row.id, scope.row.name)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="updateFormVisible ? '修改空间' : '创建空间'" :visible.sync="createFormVisible"
      @close="closeFormDialog" :destroy-on-close="true">
        <div v-loading="dialogLoading">
          <div class="dialogContent" style="">
            <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
              <el-form-item label="名称" prop="name" autofocus>
                <el-input v-model="form.name" autocomplete="off" placeholder="请输入空间名称" size="small"></el-input>
              </el-form-item>
              <el-form-item label="描述" prop="description">
                <el-input v-model="form.description" type="textarea" autocomplete="off" placeholder="请输入空间描述" size="small"></el-input>
              </el-form-item>
              <el-form-item label="集群" prop="" :required="true">
                <el-select v-model="form.cluster_id" placeholder="请选择要绑定的集群" size="small" style="width: 100%;"
                  @change="fetchNamespace" :disabled="updateFormVisible">
                  <el-option
                    v-for="item in clusters"
                    :key="item.name"
                    :label="item.name1"
                    :value="item.name"
                    :disabled="item.status != 'Connect'">
                  </el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="命名空间" prop="" :required="true">
                <el-select v-model="form.namespace" placeholder="请选择要绑定的命名空间" size="small" style="width: 100%;"
                  :disabled="updateFormVisible">
                  <el-option
                    v-for="item in namespaces"
                    :key="item.name"
                    :label="item.name"
                    :value="item.name">
                  </el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="负责人" prop="" :required="true">
                <el-input v-model="form.owner" autocomplete="off" placeholder="请输入该项目空间负责人" size="small"></el-input>
              </el-form-item>
            </el-form>
          </div>
          <div slot="footer" class="dialogFooter" style="margin-top: 20px;">
            <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
            <el-button type="primary" @click="updateFormVisible ? handleUpdateWorkspace() : handleCreateWorkspace()" >确 定</el-button>
          </div>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { createProject, listProjects, updateProject, deleteProject } from "@/api/project/project";
import { listCluster } from '@/api/cluster'
import { listNamespace } from '@/api/namespace'
import { Message } from "element-ui";

export default {
  name: "workspace",
  components: {
    Clusterbar,
  },
  mounted: function () {
    const that = this;
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 135;
        that.maxHeight = heightStyle;
      })();
    };
  },
  data() {
    return {
      maxHeight: window.innerHeight - 135,
      cellStyle: { border: 0 },
      titleName: ["工作空间"],
      loading: true,
      dialogLoading: false,
      createFormVisible: false,
      updateFormVisible: false,
      clusters: [],
      namespaces: [],
      form: {
        id: "",
        name: "",
        description: "",
        cluster_id: "",
        namespace: "",
        owner: this.$store.getters.username,
      },
      rules: {
        name: [{ required: true, message: '请输入空间名称', trigger: 'blur' },],
      },
      originWorkspaces: [],
      search_name: "",
      secretTypeMap: {
        'password': '密码',
        'key': '密钥',
        'token': 'AccessToken'
      }
    };
  },
  created() {
    this.fetchWorkspaces();
  },
  computed: {
    secrets() {

    }
  },
  methods: {
    nameClick: function(id) {
      this.$router.push({name: 'workspaceOverview', params: {'workspaceId': id}})
    },
    handleEdit(index, row) {
      console.log(index, row);
    },
    fetchWorkspaces() {
      this.loading = true
      listProjects().then((resp) => {
        this.originWorkspaces = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    handleCreateWorkspace() {
      if(!this.form.name) {
        Message.error("空间名称不能为空");
        return
      }
      if(!this.form.cluster_id) {
        Message.error("请选择要绑定的集群");
        return
      }
      if(!this.form.namespace) {
        Message.error("请选择要绑定的集群命名空间");
        return
      }
      if(!this.form.owner) {
        Message.error("空间负责人不能为空");
        return
      }
      let project = {
        name: this.form.name, 
        description: this.form.description, 
        cluster_id: this.form.cluster_id,
        namespace: this.form.namespace,
        owner: this.form.owner
      }
      this.dialogLoading = true
      createProject(project).then(() => {
        this.dialogLoading = false
        this.createFormVisible = false;
        Message.success("创建项目空间成功")
        this.fetchWorkspaces()
      }).catch((err) => {
        this.dialogLoading = false
        console.log(err)
      });
    },
    handleUpdateWorkspace() {
      if(!this.form.id) {
        Message.error("获取空间id参数异常，请刷新重试");
        return
      }
      let project = {
        name: this.form.name, 
        description: this.form.description, 
        owner: this.form.owner
      }
      if(!this.form.name) {
        Message.error("空间名称不能为空");
        return
      }
      if(!this.form.owner) {
        Message.error("空间负责人不能为空");
        return
      }
      this.dialogLoading = true
      updateProject(this.form.id, project).then(() => {
        this.dialogLoading = false
        this.createFormVisible = false;
        Message.success("更新项目空间成功")
        this.fetchWorkspaces()
      }).catch((err) => {
        this.dialogLoading = false
        console.log(err)
      });
    },
    handleDeleteWorkspace(id, name) {
      if(!id) {
        Message.error("获取密钥id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${name}」此项目空间?`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          this.loading = true
          deleteProject(id).then(() => {
            this.loading = false
            Message.success("删除项目空间成功")
            this.fetchWorkspaces()
          }).catch((err) => {
            this.loading = false
            console.log(err)
          });
        }).catch(() => {       
        });
    },
    nameSearch(val) {
      this.search_name = val;
    },
    fetchClusters() {
      this.namespaces = []
      listCluster()
        .then((response) => {
          this.clusters = response.data || [];
        }).catch(() => {
        })
    },
    fetchNamespace: function() {
      this.namespaces = []
      const cluster = this.form.cluster_id
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
    openCreateFormDialog() {
      this.form = {
        name: "",
        description: "",
        cluster_id: "",
        namespace: "",
        owner: this.$store.getters.username,
      }
      this.createFormVisible = true;
      this.fetchClusters()
    },
    openUpdateFormDialog(project) {
      this.form = {
        id: project.id,
        name: project.name,
        description: project.description,
        cluster_id: project.cluster_id,
        namespace: project.namespace,
        owner: project.owner,
      }
      this.fetchClusters()
      this.updateFormVisible = true;
      this.createFormVisible = true;
    },
    closeFormDialog() {
      this.updateFormVisible = false; 
      this.createFormVisible = false;
    }
  },
};
</script>


<style lang="scss" scoped>
@import "~@/styles/variables.scss";

.table-fix {
  height: calc(100% - 100px);
}

</style>
