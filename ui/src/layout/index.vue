<template>
  <div class="app-wrapper">
    <div class="fixed-header">
      <navbar :nav="nav" />
    </div>
    <template v-if="hasSideBar">
      <sidebar class="sidebar-container" />
    </template>
    <div class="main-container" :style='hasSideBar ? "" : "width: 100%; margin-left:0px"'>
      <template v-if="$viewerRole()">
        <app-main />
      </template>
      <template v-else>
        <div style="min-height: calc(100vh - 50px); width: 100%; position: relative;
           overflow: hidden; padding-top: 110px; text-align: center;">
          您没有权限访问该页面，请联系管理员进行授权！
        </div>
      </template>
      <!-- main container -->
    </div>

    <el-dialog title="导入YAML" :visible.sync="nav.dialog" :close-on-click-modal="false" width="60%" top="55px">
      <yaml v-if="nav.dialog" v-model="nav.yamlValue" ></yaml>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="nav.dialog = false" size="small">取 消</el-button>
        <el-button plain @click="applyYaml()" size="small">确 定</el-button>
      </span>
    </el-dialog>
    <el-dialog title="修改密码" :visible.sync="nav.changePwdDialog" :close-on-click-modal="false" width="50%" top="55px">
      <div>
        <el-form :model="form" ref="form" label-position="left" label-width="90px">
          <el-form-item label="原密码" prop="name" autofocus>
            <el-input v-model="form.originPassword" type="password" autocomplete="off" placeholder="请输入原密码" size="small"></el-input>
          </el-form-item>
          <el-form-item label="新密码" prop="description">
            <el-input v-model="form.newPassword" type="password" autocomplete="off" placeholder="请输入新密码" size="small"></el-input>
          </el-form-item>
          <el-form-item label="确认新密码" prop="description">
            <el-input v-model="form.newConfirmPassword" type="password" autocomplete="off" placeholder="请输入新密码确认" size="small"></el-input>
          </el-form-item>
        </el-form>
      </div>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="nav.changePwdDialog = false; form={}" size="small">取 消</el-button>
        <el-button type="primary" @click="changePwd()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Navbar, Sidebar, AppMain} from './components'
import { Yaml } from '@/views/components'
import { Message } from 'element-ui'
//import { applyYaml } from '@/api/cluster'
import { applyResource } from '@/api/cluster/resource'
import { updatePassword } from '@/api/user'

export default {
  name: 'Layout',
  
  components: {
    Navbar,
    Sidebar,
    AppMain,
    Yaml
  },
  data() {
    return {
      yamlValue: "",
      nav: {dialog: false, yamlValue: '', changePwdDialog: false},
      form: {}
    }
  },
  computed: {
    hasSideBar() {
      const route = this.$route
      const { meta } = route
      if (meta.noSidebar) {
        this.$store.dispatch('watchCluster', '')
        this.$store.dispatch('watchNamespace', '')
        return false
      }
      return true
    },
  },
  methods: {
    applyYaml: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!this.nav.yamlValue) {
        Message.error("请输入要导入的YAML")
        return
      }
      applyResource(cluster, this.nav.yamlValue).then((resp) => {
        Message.success(resp.msg)
        this.nav.dialog = false
      }).catch(() => {
        // console.log(e) 
      })
    },
    changePwd() {
      if(!this.form.originPassword) {
        Message.error("请输入原密码")
        return
      }
      if(!this.form.newPassword) {
        Message.error("请输入新密码")
        return
      }
      if(!this.form.newConfirmPassword) {
        Message.error("请输入新密码确认")
        return
      }
      if(this.form.newConfirmPassword != this.form.newPassword) {
        Message.error("新密码两次输入不同，请重新输入")
        return
      }
      updatePassword({origin_password: this.form.originPassword, new_password: this.form.newPassword}).then((resp) => {
        Message.success("修改密码成功")
        this.nav.changePwdDialog = false
      }).catch(() => {
        // console.log(e) 
      })
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "~@/styles/mixin.scss";
  @import "~@/styles/variables.scss";

  .app-wrapper {
    @include clearfix;
    position: relative;
    height: 100%;
    width: 100%;
  }

  .fixed-header {
    position: fixed;
    top: 0;
    right: 0;
    z-index: 9;
    width: 100%;
    transition: width 0.28s;
  }

  .fixed-sidebar {
    position: fixed;
    top: 50px;
    right: 0;
    z-index: 9;
    width: $sideBarWidth;
    transition: width 0.28s;
  }

</style>
