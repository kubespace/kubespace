<template>
  <div class="app-wrapper">
    <div class="fixed-header">
      <navbar :nav="nav" />
    </div>
    <template v-if="hasSideBar">
      <sidebar class="sidebar-container" />
    </template>
    <div class="main-container" :style='hasSideBar ? "" : "width: 100%; margin-left:0px"'>
      <app-main />
      <!-- main container -->
    </div>

    <el-dialog title="导入YAML" :visible.sync="nav.dialog" :close-on-click-modal="false" width="60%" top="55px">
      <yaml v-if="nav.dialog" v-model="nav.yamlValue" ></yaml>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="nav.dialog = false" size="small">取 消</el-button>
        <el-button plain @click="applyYaml()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Navbar, Sidebar, AppMain} from './components'
import { Yaml } from '@/views/components'
import { Message } from 'element-ui'
import { applyYaml } from '@/api/cluster'

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
      nav: {dialog: false, yamlValue: ''}
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
      console.log(cluster, this.nav.yamlValue)
      applyYaml(cluster, this.nav.yamlValue).then((resp) => {
        console.log(resp.msg)
        Message.success(resp.msg)
      }).catch(() => {
        // console.log(e) 
      })
    },
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
