<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="createSecretFormDialog" createDisplay="创建密钥"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="originSecrets"
        class="table-fix"
        :cell-style="cellStyle"
        v-loading="loading"
        :default-sort = "{prop: 'name'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column prop="name" label="名称" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="roles" label="密钥类型" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ secretTypeMap[scope.row.type] }}
          </template>
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
            <div class="tableOperate">
              <el-link :underline="false" style="margin-right: 15px; color:#409EFF" @click="updateSecretFormDialog(scope.row)">编辑</el-link>
              <el-link :underline="false" style="color: #F56C6C" @click="handleDeleteSecret(scope.row.id, scope.row.name)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="updateSecretVisible ? '修改密钥' : '创建密钥'" :visible.sync="createSecretFormVisible"
      @close="closeSecretDialog" :destroy-on-close="true">
        <div class="dialogContent" style="">
          <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
            <el-form-item label="名称" prop="name">
              <el-input :disabled="updateSecretVisible" v-model="form.name" autocomplete="off" placeholder="请输入密钥名称" size="small"></el-input>
            </el-form-item>
            <el-form-item label="描述" prop="description">
              <el-input v-model="form.description" type="textarea" autocomplete="off" placeholder="请输入密钥描述" size="small"></el-input>
            </el-form-item>
            <el-form-item label="密钥类型" prop="secret_type">
              <el-radio-group v-model="form.secret_type">
                <el-radio label="password">用户密码</el-radio>
                <el-radio label="key">私钥</el-radio>
                <el-radio label="token">AccessToken</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="用户" prop="user" v-if="form.secret_type == 'password'">
              <el-input v-model="form.user" autocomplete="off" placeholder="请输入用户" size="small"></el-input>
            </el-form-item>
            <el-form-item label="密码" prop="password" v-if="form.secret_type == 'password'">
              <el-input v-model="form.password" autocomplete="off" placeholder="请输入密码" size="small" show-password></el-input>
            </el-form-item>
            <el-form-item label="私钥" prop="private_key" v-if="form.secret_type == 'key'">
              <el-input v-model="form.private_key" class="dialogTextarea" type="textarea" autocomplete="off" placeholder="请输入私钥" size="small"></el-input>
            </el-form-item>
            <el-form-item label="AccessToken" prop="access_token" v-if="form.secret_type == 'token'">
              <el-input v-model="form.access_token" class="dialogTextarea" type="textarea" autocomplete="off" placeholder="请输入私钥" size="small"></el-input>
            </el-form-item>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter">
          <el-button @click="createSecretFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="updateSecretVisible ? handleUpdateSecret() : handleCreateSecret()" >确 定</el-button>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { createSecret, listSecret, updateSecret, deleteSecret } from "@/api/settings/secret";
import { Message } from "element-ui";

export default {
  name: "secret",
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
      titleName: ["密钥管理"],
      loading: true,
      createSecretFormVisible: false,
      updateSecretVisible: false,
      form: {
        id: "",
        name: "",
        description: "",
        user: "",
        password: "",
        private_key: "",
        access_token: "",
        secret_type: "password"
      },
      rules: {
        name: [{ required: true, message: '请输入密钥名称', trigger: 'blur' },],
        secret_type: [{ required: true, message: '请选择密钥类型', trigger: 'blur' },],
        user: [{ required: true, message: '请输入用户', trigger: 'blur' },],
        password: [{ required: true, message: '请输入密码', trigger: 'blur' },],
        private_key: [{ required: true, message: '请输入私钥', trigger: 'blur' },],
        access_token: [{ required: true, message: '请输入AccessToken', trigger: 'blur' },],
      },
      originSecrets: [],
      search_name: "",
      secretTypeMap: {
        'password': '密码',
        'key': '密钥',
        'token': 'AccessToken'
      }
    };
  },
  created() {
    this.fetchSecrets();
  },
  computed: {
    secrets() {

    }
  },
  methods: {
    handleEdit(index, row) {
      console.log(index, row);
    },
    fetchSecrets() {
      this.loading = true
      listSecret().then((resp) => {
        this.originSecrets = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    handleCreateSecret() {
      if(!this.form.name) {
        Message.error("密钥名称不能为空");
        return
      }
      if(!this.form.secret_type) {
        Message.error("密钥类别不能为空，请重新选择");
        return
      }
      let secret = {
        name: this.form.name, 
        secret_type: this.form.secret_type, 
        description: this.form.description
      }
      if(this.form.secret_type == 'password') {
        if (!this.form.user) {
          Message.error("密钥用户不能为空");
          return
        }
        if (!this.form.password) {
          Message.error("密码不能为空");
          return
        }
        secret['user'] = this.form.user
        secret['password'] = this.form.password
      } else if (this.form.secret_type == 'key') {
        if (!this.form.private_key) {
          Message.error("私钥不能为空");
          return
        }
        secret['private_key'] = this.form.private_key
      } else if (this.form.secret_type == 'token') {
        if (!this.form.access_token) {
          Message.error("AccessToken不能为空");
          return
        }
        secret['access_token'] = this.form.access_token
      }
      createSecret(secret).then(() => {
        this.createSecretFormVisible = false;
        Message.success("创建密钥成功")
        this.fetchSecrets()
      }).catch((err) => {
        console.log(err)
      });
    },
    handleUpdateSecret() {
      if(!this.form.id) {
        Message.error("获取密钥id参数异常，请刷新重试");
        return
      }
      let secret = {
        name: this.form.name,
        secret_type: this.form.secret_type,
        description: this.form.description
      }
      if(this.form.secret_type == 'password') {
        if (!this.form.user) {
          Message.error("密钥用户不能为空");
          return
        }
        if (!this.form.password) {
          Message.error("密码不能为空");
          return
        }
        secret['user'] = this.form.user
        secret['password'] = this.form.password
      } else if (this.form.secret_type == 'key') {
        if (!this.form.private_key) {
          Message.error("私钥不能为空");
          return
        }
        secret['private_key'] = this.form.private_key
      } else if (this.form.secret_type == 'token') {
        if (!this.form.access_token) {
          Message.error("AccessToken不能为空");
          return
        }
        secret['access_token'] = this.form.access_token
      }
      updateSecret(this.form.id, secret).then(() => {
        this.createSecretFormVisible = false;
        Message.success("更新密钥成功")
        this.fetchSecrets()
      }).catch((err) => {
        console.log(err)
      });
    },
    handleDeleteSecret(id, name) {
      if(!id) {
        Message.error("获取密钥id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${name}」此密钥?`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          deleteSecret(id).then(() => {
            Message.success("删除密钥成功")
            this.fetchSecrets()
          }).catch((err) => {
            console.log(err)
          });
        }).catch(() => {       
        });
    },
    nameSearch(val) {
      this.search_name = val;
    },
    createSecretFormDialog() {
      this.createSecretFormVisible = true;
    },
    updateSecretFormDialog(secret) {
      this.form = {
        id: secret.id,
        name: secret.name,
        description: secret.description,
        secret_type: secret.type || 'password',
        user: secret.user,
        password: '',
        private_key: '',
        access_token: '',
      }
      this.updateSecretVisible = true;
      this.createSecretFormVisible = true
    },
    closeSecretDialog() {
      this.form = {
        name: "",
        description: "",
        user: "",
        password: "",
        private_key: "",
        access_token: "",
        secret_type: "password"
      }
      this.updateSecretVisible = false; 
      this.createSecretFormVisible = false;
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
