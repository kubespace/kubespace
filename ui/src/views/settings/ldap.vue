<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="createLdapFormDialog"
      createDisplay="创建Ldap服务" />
    <div class="dashboard-container" ref="tableCot">
      <el-table ref="multipleTable" :data="originLdaps" class="table-fix" :cell-style="cellStyle" v-loading="loading"
        :default-sort="{ prop: 'name' }" tooltip-effect="dark" style="width: 100%">
        <el-table-column prop="name" label="名称" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="url" label="URL" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="enable" label="是否启用" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 15px;"
                @click="startProgress(scope.row.id,scope.row.name,scope.row.enable)">同步</el-link>
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 15px;"
                @click="updateLdapFormDialog(scope.row)">编辑</el-link>
              <el-link :disabled="!$editorRole()" :underline="false" type="danger"
                @click="handleDeleteLdap(scope.row.id, scope.row.name)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog title="同步进度" :visible.sync="progressVisible" @close="closeLdapProgressDialog" :destroy-on-close="true">
        <el-progress :text-inside="true" :stroke-width="26" :percentage="progressPercent" :color="progressColor" />
      </el-dialog>

      <el-dialog :title="updateLdapVisible ? '修改Ldap服务' : '创建Ldap服务'" :visible.sync="createLdapFormVisible"
        @close="closeLdapDialog" :destroy-on-close="true">
        <div v-loading="dialogLoading">
          <div class="dialogContent" style="">
            <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
              <el-form-item label="名称" prop="name">
                <el-input :disabled="updateLdapVisible" v-model="form.name" autocomplete="off" placeholder="请输入ldap名称"
                  size="small"></el-input>
              </el-form-item>
              <el-form-item label="Ldap开关" prop="enable">
                <el-radio-group v-model="form.enable">
                  <el-radio label="1">开启</el-radio>
                  <el-radio label="0">关闭</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="Ldap地址" prop="url">
                <el-input v-model="form.url" autocomplete="off" clearable placeholder="请输入URL(eg: 192.168.0.1:389)" size="small"></el-input>
              </el-form-item>
              <el-form-item label="BaseDN" prop="baseDN">
                <el-input v-model="form.baseDN" autocomplete="off" clearable placeholder="请输入baseDN(eg: dc=xxx,dc=com)"
                  size="small"></el-input>
              </el-form-item>
              <el-form-item label="Admin用户" prop="adminDN">
                <el-input v-model="form.adminDN" autocomplete="off" clearable placeholder="请输入adminDN用户"
                  size="small"></el-input>
              </el-form-item>
              <el-form-item label="Admin密码" prop="adminDNPass">
                <el-input v-model="form.adminDNPass" autocomplete="new-password" clearable placeholder="请输入adminDN密码"
                  size="small" show-password></el-input>
              </el-form-item>
            </el-form>
          </div>
          <div slot="footer" class="dialogFooter" style="margin-top: 20px;">
            <el-button @click="createLdapFormVisible = false" style="margin-right: 20px;">取 消</el-button>
            <el-button type="primary" @click="updateLdapVisible ? handleUpdateLdap() : handleCreateLdap()">确 定</el-button>
          </div>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { createLdap, listLdap, updateLdap, deleteLdap } from "@/api/settings/ldap";
import { Message } from "element-ui";
import { thistle } from "color-name";

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
      progressVisible: false,
      progressPercent: 0,
      progressColor: '#409EFF',
      progressStrokeWidth: 30,
      progressInterval: null,

      maxHeight: window.innerHeight - 150,
      cellStyle: { border: 0 },
      titleName: ["LDAP管理"],
      loading: true,
      dialogLoading: false,
      createLdapFormVisible: false,
      updateLdapVisible: false,
      form: {
        id: "",
        name: "",
        enable: "1",
        url: "",
        baseDN: "",
        adminDN: "",
        adminDNPass: "password"
      },
      rules: {
        name: [{ required: true, message: ' ', trigger: 'blur' },],
        enable: [{ required: true, message: ' ', trigger: 'blur' },],
        url: [{ required: true, message: ' ', trigger: 'blur' },],
        baseDN: [{ required: true, message: ' ', trigger: 'blur' },],
        adminDN: [{ required: true, message: ' ', trigger: 'blur' },],
        adminDNPass: [{ required: true, message: ' ', trigger: 'blur' },],
      },
      originLdaps: [],
      search_name: ""
    };
  },
  created() {
    this.fetchLdaps();
  },
  beforeDestroy() {
    if (this.source) {
      this.source.close();
    }
    this.progressVisible=false
    this.progressPercent = 0
  },
  methods: {
    closeLdapProgressDialog(){
      if (this.source) {
        this.source.close();
      }
      this.progressVisible=false
      this.progressPercent = 0
    },
    startProgress(id, name, enable) {
      
      if (enable == 0 ) {
        Message.error(name+"-Ldap服务未启用");
        return
      }

      // 显示进度条对话框
      this.progressVisible = true;

      // 设置进度条的颜色和宽度
      this.progressColor = '#409EFF';
      this.progressStrokeWidth = 6;

      const timestamp = new Date().getTime()
      const Url = `/api/v1/settings/ldap/sync_progress/${id}/${timestamp}`
      this.source = new EventSource(Url)
      this.source.addEventListener('progress', event => {
        this.progressPercent = event.data
      })
      this.source.addEventListener('error', event => {
          Message.error(event.data);
          this.closeLdapProgressDialog()
      })
      this.source.addEventListener('success', event => {
          Message.success(event.data);
          this.closeLdapProgressDialog()
      })
    },
    handleEdit(index, row) {
      console.log(index, row);
    },
    fetchLdaps() {
      this.loading = true
      listLdap().then((resp) => {
        this.originLdaps = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    handleCreateLdap() {
      if (!this.form.name) {
        Message.error("Ldap名称不能为空");
        return
      }
      if (!this.form.enable) {
        Message.error("是否开启不能为空，请选择");
        return
      }
      if (!this.form.url) {
        Message.error("ldap服务地址不能为空");
        return
      }else{
        const ipPortRegex = /^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}$/
        if (!ipPortRegex.test(this.form.url)) {
          Message.error("l请输入正确的 IP 地址和端口号（例如 192.168.1.1:389");
        } 
      }
      if (!this.form.baseDN) {
        Message.error("baseDN不能为空");
        return
      }
      if (!this.form.adminDN) {
        Message.error("adminDN不能为空");
        return
      }

      let ldap = {
        name: this.form.name,
        enable: this.form.enable,
        url: this.form.url,
        baseDN: this.form.baseDN,
        adminDN: this.form.adminDN,
        adminDNPass: this.form.adminDNPass
      }
      this.dialogLoading = true
      createLdap(ldap).then(() => {
        this.dialogLoading = false
        this.createLdapFormVisible = false;
        Message.success("创建密钥成功")
        this.fetchLdaps()
      }).catch((err) => {
        this.dialogLoading = false
        console.log(err)
      });
    },
    handleUpdateLdap() {
      if (!this.form.id) {
        Message.error("获取密钥id参数异常，请刷新重试");
        return
      }
      if (!this.form.name) {
        Message.error("Ldap名称不能为空");
        return
      }
      if (!this.form.enable) {
        Message.error("是否开启不能为空，请选择");
        return
      }
      if (!this.form.url) {
        Message.error("ldap服务地址不能为空");
        return
      }
      if (!this.form.baseDN) {
        Message.error("baseDN不能为空");
        return
      }
      if (!this.form.adminDN) {
        Message.error("adminDN不能为空");
        return
      }
      let ldap = {
        name: this.form.name,
        enable: this.form.enable,
        url: this.form.url,
        baseDN: this.form.baseDN,
        adminDN: this.form.adminDN,
        adminDNPass: this.form.adminDNPass
      }
      this.dialogLoading = true
      updateLdap(this.form.id, ldap).then(() => {
        this.dialogLoading = false
        this.createLdapFormVisible = false;
        Message.success("更新Ldap成功")
        this.fetchLdaps()
      }).catch((err) => {
        this.dialogLoading = false
        console.log(err)
      });
    },
    handleDeleteLdap(id, name) {
      if (!id) {
        Message.error("获取密钥id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${name}」此密钥?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        deleteLdap(id).then(() => {
          this.loading = false
          Message.success("删除Ldap成功")
          this.fetchLdaps()
        }).catch((err) => {
          console.log(err)
          this.loading = false
        });
      }).catch(() => {
      });
    },
    nameSearch(val) {
      this.search_name = val;
    },
    createLdapFormDialog() {
      this.createLdapFormVisible = true;
    },
    updateLdapFormDialog(ldap) {
      this.form = {
        id: ldap.id,
        name: ldap.name,
        enable: ldap.enable,
        url: ldap.url,
        baseDN: ldap.baseDN,
        adminDN: ldap.adminDN,
        adminDNPass: ldap.adminDNPass
      }
      this.updateLdapVisible = true;
      this.createLdapFormVisible = true
    },
    closeLdapDialog() {
      this.form = {
        name: "",
        url: "",
        baseDN: "",
        adminDN: "",
        adminDNPass: ""
      }
      this.updateLdapVisible = false;
      this.createLdapFormVisible = false;
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
