<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="createFormDialog" createDisplay="创建资源"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="originResources"
        class="table-fix"
        :cell-style="cellStyle"
        v-loading="loading"
        :default-sort = "{prop: 'name'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column prop="name" label="名称" show-overflow-tooltip min-width="13">
        </el-table-column>
        <el-table-column prop="global" label="全局共享" show-overflow-tooltip min-width="8">
          <template slot-scope="scope">
            {{ scope.row.global ? "是" : "否" }}
          </template>
        </el-table-column>
        <el-table-column prop="roles" label="类型" show-overflow-tooltip min-width="8">
          <template slot-scope="scope">
            {{ typeMap[scope.row.type] }}
          </template>
        </el-table-column>
        <el-table-column prop="value" label="资源地址" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip min-width="18">
        </el-table-column>
        <el-table-column prop="update_user" label="操作人" show-overflow-tooltip min-width="10">
          <template slot-scope="scope">
            {{ scope.row.update_user }}
          </template>
        </el-table-column>
        <el-table-column prop="update_time" label="更新时间" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.update_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template slot-scope="scope">
            <div class="tableOperate" v-if="scope.row.workspace_id == workspaceId">
              <el-link :underline="false" style="margin-right: 15px; color:#409EFF" @click="updateFormDialog(scope.row)">编辑</el-link>
              <el-link :underline="false" style="color: #F56C6C" @click="handleDeleteResource(scope.row.id, scope.row.name)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="updateVisible ? '修改资源' : '创建流水线资源'" :visible.sync="createFormVisible"
      @close="closeDialog" :destroy-on-close="true">
        <div v-loading="dialogLoading">
          <div class="dialogContent" style="">
            <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
              <el-form-item label="资源名称" prop="name">
                <el-input :disabled="updateVisible" v-model="form.name" autocomplete="off" placeholder="请输入资源名称" size="small"></el-input>
              </el-form-item>
              <el-form-item label="资源类型" prop="type" size="small" >
                <el-radio-group v-model="form.type" :disabled="updateVisible">
                  <el-radio-button label="image">容器镜像</el-radio-button>
                  <el-radio-button label="host">主 机</el-radio-button>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="全局共享" required>
                <el-switch v-model="form.global"></el-switch>
              </el-form-item>
              <el-form-item label="镜像地址" v-if="form.type == 'image'" required>
                <el-input v-model="form.value" autocomplete="off" clearable placeholder="请输入容器镜像地址" size="small"></el-input>
              </el-form-item>
              <el-form-item label="主机地址" v-if="form.type == 'host'" required>
                <el-input v-model="form.value" autocomplete="off" clearable placeholder="请输入主机地址" size="small"></el-input>
              </el-form-item>
              <el-form-item label="认证密钥">
                <el-select v-model="form.secret_id" placeholder="请选择访问认证密钥" size="small" style="width: 100%">
                  <el-option label="" value="">不选择</el-option>
                  <el-option
                    v-for="secret in form.type == 'image' ? imageSecrets : hostSecrets"
                    :key="secret.id"
                    :label="secret.name"
                    :value="secret.id">
                  </el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="描述" prop="description">
                <el-input v-model="form.description" type="textarea" autocomplete="off" placeholder="请输入资源描述" size="small"></el-input>
              </el-form-item>
            </el-form>
          </div>
          <div slot="footer" class="dialogFooter" style="margin-top: 15px;">
            <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
            <el-button type="primary" @click="updateVisible ? handleUpdateResource() : handleCreateResource()" >确 定</el-button>
          </div>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { listSecret } from "@/api/settings/secret";
import { createResource, listResources, updateResource, deleteResource } from "@/api/pipeline/resource";
import { Message } from "element-ui";

export default {
  name: "resource",
  components: {
    Clusterbar,
  },
  mounted: function () {
    const that = this;
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 150;
        that.maxHeight = heightStyle;
      })();
    };
  },
  data() {
    return {
      maxHeight: window.innerHeight - 150,
      cellStyle: { border: 0 },
      titleName: ["资源管理"],
      loading: true,
      createFormVisible: false,
      updateVisible: false,
      dialogLoading: false,
      form: {
        id: "",
        name: "",
        description: "",
        value: "",
        type: "image",
        secret_id: "",
        global: false
      },
      typeMap: {
        image: "镜像",
        host: "主机"
      },
      rules: {
        name: [{ required: true, message: ' ', trigger: 'blur' },],
        type: [{ required: true, message: ' ', trigger: 'blur' },],
        value: [{ required: true, message: ' ', trigger: 'blur' },],
      },
      originResources: [],
      search_name: "",
      secrets: []
    };
  },
  created() {
    this.fetchResources();
  },
  computed: {
    workspaceId() {
      return this.$route.params.workspaceId
    },
    imageSecrets() {
      let res = []
      for(let s of this.secrets) {
        if(s.type == 'password') res.push(s)
      }
      return res
    },
    hostSecrets() {
      let res = []
      for(let s of this.secrets) {
        if(s.type == 'password' || s.type == 'key') res.push(s)
      }
      return res
    }
  },
  methods: {
    handleEdit(index, row) {
      console.log(index, row);
    },
    fetchResources() {
      this.loading = true
      listResources(this.workspaceId).then((resp) => {
        this.originResources = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        this.loading = false
      })
    },
    fetchSecrets() {
      listSecret().then((resp) => {
        this.secrets = resp.data ? resp.data : []
      }).catch((err) => {
        this.loading = false
      })
    },
    handleCreateResource() {
      if(!this.form.name) {
        Message.error("资源名称不能为空");
        return
      }
      if(!this.form.type) {
        Message.error("资源类型不能为空，请重新选择");
        return
      }
      if(!this.form.value) {
        Message.error("资源地址不能为空");
        return
      }
      let resource = {
        workspace_id: parseInt(this.workspaceId),
        name: this.form.name, 
        type: this.form.type, 
        value: this.form.value,
        description: this.form.description,
        global: this.form.global
      }
      if(this.form.secret_id) resource['secret_id'] = this.form.secret_id
      this.dialogLoading = true
      createResource(resource).then(() => {
        this.dialogLoading = false
        this.createFormVisible = false;
        Message.success("创建资源成功")
        this.fetchResources()
      }).catch((err) => {
        this.dialogLoading = false
      });
    },
    handleUpdateResource() {
      if(!this.form.id) {
        Message.error("获取资源id参数异常，请刷新重试");
        return
      }
      if(!this.form.value) {
        Message.error("资源地址不能为空")
        return
      }
      let resource = {
        value: this.form.value,
        description: this.form.description,
        global: thif.form.global
      }
      if(this.form.secret_id) resource['secret_id'] = this.form.secret_id
      this.dialogLoading = true
      updateResource(this.form.id, resource).then(() => {
        this.dialogLoading = false
        this.createFormVisible = false;
        Message.success("更新资源成功")
        this.fetchResources()
      }).catch((err) => {
        this.dialogLoading = false
      });
    },
    handleDeleteResource(id, name) {
      if(!id) {
        Message.error("获取资源id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${name}」此资源?`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          deleteResource(id).then(() => {
            Message.success("删除资源成功")
            this.fetchResources()
          }).catch((err) => {
            console.log(err)
          });
        }).catch(() => {       
        });
    },
    nameSearch(val) {
      this.search_name = val;
    },
    
    createFormDialog() {
      this.fetchSecrets()
      this.form = {
        id: "",
        name: "",
        description: "",
        value: "",
        type: "image",
        secret_id: "",
        global: false,
      }
      this.createFormVisible = true;
    },
    updateFormDialog(resource) {
      this.fetchSecrets()
      this.form = {
        id: resource.id,
        name: resource.name,
        description: resource.description,
        type: resource.type,
        value: resource.value,
        global: resource.global,
      }
      if(resource.secret_id) this.form.secret_id = resource.secret_id
      this.updateVisible = true;
      this.createFormVisible = true
    },
    closeDialog() {
      this.updateVisible = false; 
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
