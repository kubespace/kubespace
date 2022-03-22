<template>
  <div class="navbar-item">
    <el-menu :default-active="activeMenu" mode="horizontal" >
        <el-menu-item index="workspace" class="submenu-class" v-on:click="workspaceClick">工作空间</el-menu-item>
        <el-menu-item index="pipeline" class="submenu-class" v-on:click="pipelineClick">流水线</el-menu-item>
        <el-menu-item index="cluster" class="submenu-class" v-on:click="clusterClick">集群管理</el-menu-item>
        <el-menu-item index="appstore" class="submenu-class" v-on:click="appStoreClick">应用商店</el-menu-item>
        <el-menu-item index="settings" class="submenu-class" v-on:click="globalClick">平台配置</el-menu-item>
    </el-menu>
  </div>
</template>

<script>

export default {
  data() {
    return {
      
    };
  },
  created() {
    
  },
  watch: {
    
  },
  computed: {
    activeMenu() {
      const route = this.$route
      const { meta } = route
      return meta.group ? meta.group : ''
    },
  },
  methods: {
    globalClick() {
      this.$router.push({name: 'settinsSecret'})
    },
    pipelineClick() {
      this.$router.push({name: 'pipelineWorkspace'})
    },
    workspaceClick() {
      this.$router.push({name: 'workspaceIndex'})
    },
    clusterClick() {
      this.$router.push({name: 'clusterIndex'})
    },
    appStoreClick() {
      this.$router.push({name: 'appstoreIndex'})
    }
  }
}
</script>

<style lang="scss" scoped>
.navbar-item {
  position: relative;
  display: inline-block;
  height: 50px;
  line-height: 50px;
  overflow: hidden;
  margin-left: 90px;

}
</style>
