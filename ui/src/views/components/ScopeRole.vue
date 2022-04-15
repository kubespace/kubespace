<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch">
      <div slot="right-btn" style="display: inline-block">
        <el-button size="small" type="primary" @click="createUserRoleFormDialog" icon="el-icon-plus" :disabled="!$adminRole()">
          添加成员
        </el-button>
      </div>
    </clusterbar>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="originUserRoles"
        class="table-fix"
        :cell-style="cellStyle"
        v-loading="loading"
        :default-sort = "{prop: 'username'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column prop="username" label="用户" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="roles" label="权限" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ userRoleMap[scope.row.role] }}
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
              <el-link :disabled="!$adminRole()" :underline="false" type="primary" style="margin-right: 15px;" @click="updateUserRoleFormDialog(scope.row)">编辑</el-link>
              <el-link :disabled="!$adminRole()" :underline="false" type="danger" @click="handleDeleteUserRole(scope.row.id, scope.row.username)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="updateFormVisible ? '修改用户授权' : '用户授权'" :visible.sync="createFormVisible"
      @close="closeUserRoleDialog" :destroy-on-close="true">
        <div class="dialogContent" style="">
          <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
            <el-form-item label="用户" required>
              <el-select v-model="form.userIds" placeholder="请选择授权用户" size="small" style="width: 100%;" multiple
                :disabled="updateFormVisible">
                <el-option
                  v-for="item in users"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id">
                </el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="权限" required>
              <el-select v-model="form.role" placeholder="请选择平台权限" size="small" style="width: 100%;">
                <el-option
                  v-for="item in userRoles"
                  :key="item.type"
                  :label="item.name"
                  :value="item.type">
                </el-option>
              </el-select>
            </el-form-item>
          </el-form>
          <div style="background-color: #f4f4f5; margin-top: 10px;
    color: #909399; border: 1px solid #e9e9eb;border-radius: 4px;padding: 0 10px;font-size: 12px; line-height: 16px;" >
            <div style="margin: 10px 0px;">
              <span style="font-weight: 550">{{ roleInfo.viewer }}：</span>
              <span>{{ roleInfo.viewerDesc }}</span>
            </div>
            <div style="margin-bottom: 10px;">
              <span style="font-weight: 550">{{ roleInfo.editor }}：</span>
              <span>{{ roleInfo.editorDesc }}</span>
            </div>
            <div style="margin-bottom: 10px;">
              <span style="font-weight: 550">{{ roleInfo.admin }}：</span>
              <span>{{ roleInfo.adminDesc }}</span>
            </div>
          </div>
        </div>
        <div slot="footer" class="dialogFooter">
          <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="updateFormVisible ? handleUpdateUserRole() : handleCreateUserRole()" >确 定</el-button>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { listUserRole, updateUserRole, deleteUserRole } from "@/api/settings/user_role";
import { getUser } from "@/api/user";
import { Message } from "element-ui";

export default {
  name: "ScopeRole",
  components: {
    Clusterbar,
  },
  props: ['scope', 'scopeId'],
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
    let roleInfo = {
      project: {
        title: "成员管理",
        viewer: "空间观察员",
        editor: "空间编辑员",
        admin: "空间管理员",
        viewerDesc: "对当前工作空间中所有资源有查看权限",
        editorDesc: "对当前工作空间中所有资源有查看、增加、修改、删除权限，但不包括成员管理权限",
        adminDesc: "对当前工作空间中所有资源有查看、增加、修改、删除权限，包括成员管理权限"
      },
      cluster: {
        title: "成员管理",
        viewer: "集群观察员",
        editor: "集群编辑员",
        admin: "集群管理员",
        viewerDesc: "对当前集群中所有资源有查看权限",
        editorDesc: "对当前集群中所有资源有查看、增加、修改、删除权限，但不包括成员权限管理",
        adminDesc: "对当前集群中所有资源有查看、增加、修改、删除权限，包括成员权限管理"
      },
      pipeline: {
        title: "权限配置",
        viewer: "空间观察员",
        editor: "空间编辑员",
        admin: "空间管理员",
        viewerDesc: "对当前流水线空间中所有资源有查看权限",
        editorDesc: "对当前流水线空间中所有资源有查看、增加、修改、删除权限，但不包括成员权限管理",
        adminDesc: "对当前流水线空间中所有资源有查看、增加、修改、删除权限，包括成员权限管理"
      },
      platform: {
        title: "平台权限",
        viewer: "平台观察员",
        editor: "平台编辑员",
        admin: "平台管理员",
        viewerDesc: "对平台所有资源，包括工作空间、流水线、集群管理、应用商店等有查看权限",
        editorDesc: "对平台所有资源，包括工作空间、流水线、集群管理、应用商店等有查看、修改、增加、删除权限，但不包括各模块的成员权限管理",
        adminDesc: "对平台所有资源，包括工作空间、流水线、集群管理、应用商店等有查看、修改、增加、删除权限，包括各模块的成员权限管理"
      },
    }[this.scope]
    return {
      maxHeight: window.innerHeight - 135,
      cellStyle: { border: 0 },
      titleName: [roleInfo.title],
      loading: true,
      createFormVisible: false,
      updateFormVisible: false,
      form: {
        userIds: [],
        role: "",
      },
      rules: {
      },
      originUserRoles: [],
      search_name: "",
      users: [],
      roleInfo: roleInfo,
      userRoles: [{
        "type": "viewer",
        "name": roleInfo.viewer,
        "desc": roleInfo.viewerDesc
      },{
        "type": "editor",
        "name": roleInfo.editor,
        "desc": roleInfo.editorDesc
      },{
        "type": "admin",
        "name": roleInfo.admin,
        "desc": roleInfo.adminDesc
      }],

    };
  },
  created() {
    this.fetchPlatformUserRoles();
    this.fetchUsers()
  },
  computed: {
    secrets() {

    },
    userRoleMap() {
      let m = {}
      for(let r of this.userRoles) {
        m[r.type] = r.name
      }
      return m
    }
  },
  methods: {
    handleEdit(index, row) {
      console.log(index, row);
    },
    fetchPlatformUserRoles() {
      this.loading = true
      listUserRole({scope: this.scope, scope_id: this.scopeId}).then((resp) => {
        this.originUserRoles = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    fetchUsers() {
      getUser().then((response) => {
        this.users = response.data || [];
      });
    },
    handleCreateUserRole() {
      if(!this.form.role) {
        Message.error("用户权限不能为空");
        return
      }
      if(this.form.userIds.length == 0) {
        Message.error("请选择要授权的用户");
        return
      }
      let userRole = {
        user_ids: this.form.userIds, 
        role: this.form.role, 
        scope: this.scope,
        scope_id: this.scopeId,
      }
      updateUserRole(userRole).then(() => {
        this.createFormVisible = false;
        Message.success("添加用户权限成功")
        this.fetchPlatformUserRoles()
      }).catch((err) => {
        console.log(err)
      });
    },
    handleUpdateUserRole() {
      if(!this.form.role) {
        Message.error("用户权限为空");
        return
      }
      if(this.form.userIds.length == 0) {
        Message.error("请选择要授权的用户");
        return
      }
      let userRole = {
        user_ids: this.form.userIds, 
        role: this.form.role, 
        scope: this.scope,
        scope_id: this.scopeId,
      }
      updateUserRole(userRole).then(() => {
        this.createFormVisible = false;
        Message.success("用户权限修改成功")
        this.fetchPlatformUserRoles()
      }).catch((err) => {
        console.log(err)
      });
    },
    handleDeleteUserRole(id, name) {
      if(!id) {
        Message.error("获取权限id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${name}」此成员?`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          deleteUserRole(id).then(() => {
            Message.success("删除权限成功")
            this.fetchPlatformUserRoles()
          }).catch((err) => {
            console.log(err)
          });
        }).catch(() => {       
        });
    },
    nameSearch(val) {
      this.search_name = val;
    },
    createUserRoleFormDialog() {
      this.createFormVisible = true;
    },
    updateUserRoleFormDialog(userRole) {
      this.form = {
        userIds: [userRole.user_id],
        role: userRole.role
      }
      this.updateFormVisible = true;
      this.createFormVisible = true
    },
    closeUserRoleDialog() {
      this.form = {
        userIds: [],
        role: ""
      }
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
