<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="createRoleDialog" createDisplay="创建角色">
      <el-button slot="right-btn" size="small" plain type="" @click="permissionListVisible=true;">
        权限列表
      </el-button>
    </clusterbar>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="roles"
        class="table-fix"
        :cell-style="cellStyle"
        tooltip-effect="dark"
        :default-sort = "{prop: 'name'}"
        style="width: 100%"
      >
        <el-table-column 
          type="selection" 
          width="45">
        </el-table-column>
        <el-table-column 
          prop="name" 
          label="角色" 
          width="150"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column 
          prop="description" 
          label="描述" 
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="create_time"
          label="创建时间"
          width="300"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column label="" width="70">
          <template slot-scope="scope">
            <el-dropdown size="medium" v-if="['admin', 'edit', 'view'].indexOf(scope.row.name) < 0">
              <el-link :underline="false">
                <svg-icon style="width: 1.3em; height: 1.3em" icon-class="operate"/>
              </el-link>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item v-if="$updatePerm()" @click.native.prevent="setPermCheck(scope.row); updateRoleForm=true; createRoleFormVisible=true;">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em;"
                    icon-class="edit"/>
                  <span style="margin-left: 5px">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$deletePerm()" @click.native.prevent="deleteRoles([{ name: scope.row.name }])">
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

    <el-dialog :title="updateRoleForm ? '修改角色' : '创建角色'" :visible.sync="createRoleFormVisible" :destroy-on-close="true" top="5vh"
      @close="clearPermCheck(); updateRoleForm = false; createRoleFormVisible = false; form={name: '', description: ''};">
      <el-form :model="form" label-width="80px" label-position="left">
        <el-form-item label="角色名">
          <el-input v-model="form.name" :disabled='updateRoleForm' autocomplete="off"></el-input>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" autocomplete="off" placeholder="">
          </el-input>
        </el-form-item>
      </el-form>
      <el-table
        ref="rolePermTable"
        :data="permissions"
        class="table-fix"
        :cell-style="cellStyle"
        tooltip-effect="dark"
        :height="maxHeight - 220"
        style="width: 100%"
        v-if="permissions.length > 0"
      >
        <el-table-column 
          prop="name" 
          label="对象" 
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column label="权限" align="center">
          <el-table-column label="查询">
            <template slot-scope="scope">
              <template v-if="scope.row.operations.indexOf('get') >= 0">
                <el-checkbox
                  :checked="permCheck[scope.row.scope][scope.row.object]['get']"
                  @change="rolePermChange(scope.row, 'get')"></el-checkbox>
              </template>
            </template>
          </el-table-column>
          <el-table-column label="创建" >
            <template slot-scope="scope">
              <el-checkbox v-if="scope.row.operations.indexOf('create') >= 0"
                :checked="permCheck[scope.row.scope][scope.row.object]['create']"
                @change="rolePermChange(scope.row, 'create')"></el-checkbox>
            </template>
          </el-table-column>
          <el-table-column label="更新" >
            <template slot-scope="scope">
              <el-checkbox v-if="scope.row.operations.indexOf('update') >= 0"
                :checked="permCheck[scope.row.scope][scope.row.object]['update']"
                @change="rolePermChange(scope.row, 'update')"></el-checkbox>
            </template>
          </el-table-column>
          <el-table-column label="删除" >
            <template slot-scope="scope">
              <el-checkbox v-if="scope.row.operations.indexOf('delete') >= 0"
                :checked="permCheck[scope.row.scope][scope.row.object]['delete']"
                @change="rolePermChange(scope.row, 'delete')"></el-checkbox>
            </template>
          </el-table-column>
        </el-table-column>
      </el-table>
      <div slot="footer" class="dialog-footer">
        <el-button @click="createRoleFormVisible = false">取 消</el-button>
        <el-button type="primary" @click="updateRoleForm ? handleUpdateRole() : handleCreateRole()">确 定</el-button>
      </div>
    </el-dialog>

    <el-dialog title="权限列表" :visible.sync="permissionListVisible" top="5vh">
      <el-table
        ref="permTable"
        :data="permissions"
        class="table-fix"
        :cell-style="cellStyle"
        :height="maxHeight - 80"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column 
          prop="name" 
          label="对象" 
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column label="权限" align="center">
          <el-table-column label="查询">
            <template slot-scope="scope">
              <!-- <el-checkbox :value="scope.row.operations.indexOf('get') >= 0" disabled></el-checkbox> -->
              <!-- {{ scope.row.operations.indexOf('get') >= 0 ? '✅': '' }} -->
              <i v-if="scope.row.operations.indexOf('get') >= 0 " class="el-icon-check"></i>
            </template>
          </el-table-column>
          <el-table-column label="创建" >
            <template slot-scope="scope">
              <!-- <el-checkbox :value="scope.row.operations.indexOf('create') >= 0" disabled></el-checkbox> -->
              <i v-if="scope.row.operations.indexOf('create') >= 0 " class="el-icon-check"></i>
            </template>
          </el-table-column>
          <el-table-column label="更新" >
            <template slot-scope="scope">
              <!-- <el-checkbox :value="scope.row.operations.indexOf('update') >= 0" disabled></el-checkbox> -->
              <i v-if="scope.row.operations.indexOf('update') >= 0 " class="el-icon-check"></i>
            </template>
          </el-table-column>
          <el-table-column label="删除" >
            <template slot-scope="scope">
              <!-- <el-checkbox :value="scope.row.operations.indexOf('delete') >= 0" disabled></el-checkbox> -->
              <i v-if="scope.row.operations.indexOf('delete') >= 0 " class="el-icon-check"></i>
            </template>
          </el-table-column>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { getPermissions, getRoles, createRole, updateRole, deleteRoles } from "@/api/settings_role";
import { Message } from "element-ui";

export default {
  name: "settings_role",
  components: {
    Clusterbar,
  },
  created: function() {
    this.handleGetRoles();
    this.handleGetPermissions();
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
      test: true,
      titleName: ["角色管理"],
      createRoleFormVisible: false,
      updateRoleForm: false,
      form: {
        name: "",
        description: "",
      },
      formLabelWidth: "120px",
      roles: [],
      userData: [],
      search_name: "",
      permissionListVisible: false,
      permissions: [],
      permCheck: {},
    };
  },
  methods: {
    handleGetRoles() {
      var params = {}
      if(this.search_name) {
        params['name'] = this.search_name
      }
      getRoles(params).then((response) => {  
        this.roles = response.data;
      });
    },
    rolePermChange(obj, op) {
      this.permCheck[obj.scope][obj.object][op] = !this.permCheck[obj.scope][obj.object][op];
    },
    handleEdit(index, row) {
      console.log(index, row);
    },
    handleCreateRole() {
      if (!this.form.name) {
        Message.error("请输入角色名");
        return
      }
      if (!this.form.description) {
        Message.error("请输入角色描述");
        return
      }
      var perms = []
      for(var s in this.permCheck) {
        for(var o in this.permCheck[s]) {
          var ps = []
          for(var op in this.permCheck[s][o]){
            if(this.permCheck[s][o][op]) {
              ps.push(op)
            }
          }
          if(ps.length > 0){
            perms.push({
              scope: s,
              object: o,
              operations: ps
            })
          }
        }
      }
      var roleObj = {
        name: this.form.name,
        description: this.form.description,
        permissions: perms
      }
      createRole(roleObj).then(() => {
        this.handleGetRoles()
        this.createRoleFormVisible = false;
        this.clearPermCheck()
        Message.success("创建成功")
      }).catch((err) => {
        console.log(err)
        this.clearPermCheck()
      });
    },
    handleUpdateRole() {
      if (!this.form.name) {
        Message.error("请输入角色名");
        return
      }
      if (!this.form.description) {
        Message.error("请输入角色描述");
        return
      }
      var perms = []
      for(var s in this.permCheck) {
        for(var o in this.permCheck[s]) {
          var ps = []
          for(var op in this.permCheck[s][o]){
            if(this.permCheck[s][o][op]) {
              ps.push(op)
            }
          }
          if(ps.length > 0){
            perms.push({
              scope: s,
              object: o,
              operations: ps
            })
          }
        }
      }
      var roleObj = {
        name: this.form.name,
        description: this.form.description,
        permissions: perms
      }
      updateRole(this.form.name, roleObj).then(() => {
        this.handleGetRoles()
        this.createRoleFormVisible = false;
        this.clearPermCheck();
        Message.success("修改成功")
      }).catch((err) => {
        console.log(err)
        this.clearPermCheck()
      });
    },
    handleGetPermissions() {
      getPermissions().then((response) => {  
        this.permissions = response.data;
        this.clearPermCheck();
      });
    },
    clearPermCheck() {
      for(let p of this.permissions) {
        if(!this.permCheck[p.scope]) {
          this.permCheck[p.scope] = {}
        }
        this.permCheck[p.scope][p.object]  = {
          'get': false,
          'create': false,
          'update': false,
          'delete': false
        }
      }
    },
    setPermCheck(role) {
      console.log(role);
      this.clearPermCheck();
      if(role.permissions){
        for(let p of role.permissions) {
          for(let op of p.operations) {
            this.permCheck[p.scope][p.object][op] = true;
          }
        }
      }
      this.form = {name: role.name, description: role.description}
    },
    deleteRoles(delRoles) {
      if(delRoles.length <= 0) {
        Message.error('请选择要删除的角色')
        return
      }
      deleteRoles(delRoles).then((response) => {
          this.handleGetRoles()
      }).catch((e) => {
        console.log(e)
      })
    },
    nameSearch: function (val) {
      this.search_name = val;
    },
    createRoleDialog() {
      this.createRoleFormVisible = true;
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