<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="createUserDialog" createDisplay="创建用户"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="userData"
        class="table-fix"
        :cell-style="cellStyle"
        :default-sort = "{prop: 'name'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column 
          type="selection" 
          width="45"> 
        </el-table-column>
        <el-table-column 
          prop="name" 
          label="用户名" 
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column 
          prop="email" 
          label="邮箱" 
          show-overflow-tooltip>
          <template slot-scope="scope">
            {{ scope.row.email ? scope.row.email : "—" }}
          </template>
        </el-table-column>
        <el-table-column 
          prop="roles" 
          label="角色" 
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span v-for="val in scope.row.roles" :key="val" class="back-class">
                {{ val }} 
            </span>
          </template>
        </el-table-column>
        <el-table-column 
          prop="status" 
          label="状态" 
          show-overflow-tooltip>
          <template slot-scope="scope">
            {{ scope.row.status | filterStatus }}
          </template>
        </el-table-column>
        <el-table-column
          prop="last_login"
          label="上次登录时间"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column label="" width="80">
          <template slot-scope="scope">
            <el-dropdown size="medium" v-if="scope.row.name != 'admin'">
              <el-link :underline="false">
                <svg-icon style="width: 1.3em; height: 1.3em" icon-class="operate"/>
              </el-link>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item v-if="$updatePerm()" 
                  @click.native.prevent="createUserFormVisible=true; updateUserVisible=true;
                                         form={name: scope.row.name, password: '', email: scope.row.email, roles: scope.row.roles}">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em;"
                    icon-class="edit"/>
                  <span style="margin-left: 5px">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$deletePerm()" @click.native.prevent="deleteUsers([{ name: scope.row.name }])">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em;"
                    icon-class="delete"/>
                  <span style="margin-left: 5px">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog :title="updateUserVisible ? '用户修改' : '创建用户'" :visible.sync="createUserFormVisible"
      @close="form={'name': '', 'password': '', 'email': '', 'roles': []}; updateUserVisible = false; createUserFormVisible = false;">
      <div style="padding: 10px 20px;">
        <el-form :model="form" label-position="left" label-width="80px">
          <el-form-item label="用户名">
            <el-input :disabled="updateUserVisible" v-model="form.name" autocomplete="off" placeholder="请输入用户名"></el-input>
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.password" autocomplete="off" placeholder="请输入密码" show-password>
            </el-input>
          </el-form-item>
          <el-form-item label="邮箱">
            <el-input v-model="form.email" autocomplete="off" placeholder="请输入邮箱"></el-input>
          </el-form-item>
          <el-form-item label="角色">
            <el-select v-model="form.roles" multiple style="width: 100%;">
              <el-option
                v-for="item in roles"
                :key="item.name"
                :label="item.name"
                :value="item.name">
              </el-option>
            </el-select>
          </el-form-item>
        </el-form>
      </div>
        <div slot="footer" class="dialog-footer">
          <el-button @click="createUserFormVisible = false">取 消</el-button>
          <el-button type="primary" @click="updateUserVisible ? handleUpdateUser() : handleCreateUser()">确 定</el-button>
        </div>
    </el-dialog>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { createUser, getUser, updateUser, deleteUser } from "@/api/user";
import { getRoles } from "@/api/settings_role";
import { Message } from "element-ui";

export default {
  name: "member",
  components: {
    Clusterbar,
  },
  mounted: function () {
    const that = this;
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 150;
        console.log(heightStyle);
        that.maxHeight = heightStyle;
      })();
    };
    this.handleGetUser();
    this.handleGetRoles();
  },
  data() {
    return {
      maxHeight: window.innerHeight - 150,
      cellStyle: { border: 0 },
      titleName: ["用户管理"],
      createUserFormVisible: false,
      updateUserVisible: false,
      form: {
        name: "",
        email: "",
        password: "",
        roles: [],
      },
      formLabelWidth: "120px",
      userData: [],
      search_name: "",
      roles: [],
    };
  },
  filters: {
    filterStatus(val) {
      switch (val) {
        case "normal":
          val = "正常";
          break;
        case "disable":
          val = "禁用";
          break;
      }
      return val;
    },
    filterEnable(val) {
      switch (val) {
        case "normal":
          val = "禁用";
          break;
        case "disable":
          val = "启用";
          break;
      }
      return val;
    },
  },
  methods: {
    handleEdit(index, row) {
      console.log(index, row);
    },
    handleCreateUser() {
      if (!this.form.name) {
        Message.error("用户名称不能为空！");
        return
      }
      if (!this.form.password) {
        Message.error("密码不能为空！");
        return
      }
      if (!this.form.email) {
        Message.error("邮箱不能为空！");
        return
      }
      createUser(this.form).then(() => {
        this.createUserFormVisible = false;
        Message.success("创建成功")
        this.handleGetUser()
      }).catch((err) => {
        console.log(err)
      });
    },
    handleUpdateUser() {
      if (!this.form.email) {
        Message.error("邮箱不能为空！");
        return
      }
      updateUser(this.form.name, this.form).then(() => {
        this.createUserFormVisible = false;
        Message.success("修改成功")
        this.handleGetUser()
      }).catch((err) => {
        console.log(err)
      });
    },
    handleGetUser(name) {
      getUser(name).then((response) => {
        this.userData = response.data;
      });
    },
    handleGetRoles() {
      getRoles().then((response) => {
        this.roles = response.data;
        this.roles.sort((a, b) => {return a.name > b.name ? 1 : -1});
      });
    },
    handleEnableUser(name, currentStatus) {
      console.log(name, status);
      this.$confirm("此操作将禁用该用户, 是否继续?", "提示", {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
      }).then(() => {
        updateUser(name, {
          status: currentStatus == "normal" ? "disable" : "normal",
        }).then((response) => {
          console.log(response);
          this.$message({
            type: "success",
            message: "修改成功!",
          });
          this.handleGetUser();
        });
      });
    },
    deleteUsers(delUsers) {
      if(delUsers.length <= 0) {
        Message.error('请选择要删除的用户')
        return
      }
      deleteUser(delUsers).then(() => {
          this.handleGetUser()
      }).catch((e) => {
        console.log(e)
      })
    },
    nameSearch: function (val) {
      this.search_name = val;
    },
    createUserDialog() {
      this.createUserFormVisible = true;
    },
  },
};
</script>


<style lang="scss" scoped>
@import "~@/styles/variables.scss";

.table-fix {
  height: calc(100% - 100px);
}

.member-bar {
  transition: width 0.28s;
  height: 55px;
  overflow: hidden;
  box-shadow: inset 0 0 4px rgba(0, 21, 41, 0.1);
  margin: 20px 20px 0px;

  .app-breadcrumb.el-breadcrumb {
    display: inline-block;
    font-size: 20px;
    line-height: 55px;
    margin-left: 8px;

    .no-redirect {
      // color: #97a8be;
      cursor: text;
      margin-left: 15px;
      font-size: 23px;
      font-family: Avenir, Helvetica Neue, Arial, Helvetica, sans-serif;
    }
  }

  .icon-create {
    display: inline-block;
    line-height: 55px;
    margin-left: 20px;
    width: 1.8em;
    height: 1.8em;
    vertical-align: 0.8em;
    color: #bfbfbf;
  }

  .right {
    float: right;
    height: 100%;
    line-height: 55px;
    margin-right: 25px;

    .el-input {
      width: 195px;
      margin-left: 15px;
    }

    .el-select {
      .el-select__tags {
        white-space: nowrap;
        overflow: hidden;
      }
    }
  }
}
</style>