<template>
  <div v-if="result && result.data">
    <div>
      <span>工作空间：</span>
      <span>{{ project }}</span>
    </div>
    <div v-for="(app, i) of apps" :key="i">
      <span>应用：</span>
      <span>
        <el-link style="font-weight: 400; margin-top: -5px;" type="primary" 
          :underline="false" @click="appClick(app)">
          {{ app.name }}
        </el-link>，
      </span>
      <span>部署镜像：</span>
      <div style="">
        <div v-for="(img, j) of app.upgrade_images" :key="j">
          {{ img.after }} ( &lt;- {{ img.before }} )
        </div>
      </div>
    </div>
  </div>
</template>

<script>

export default {
  name: 'AppDeployResult',
  data() {
    return {
    }
  },
  props: ['result', 'params'],
  computed: {
    project() {
      if(!this.result.data) return ''
      if(this.result.data.project) return this.result.data.project
      return ''
    },
    apps() {
      if(!this.result.data) return []
      if(this.result.data.apps) return this.result.data.apps
      return []
    },
    projectId() {
      if(!this.params) return 0
      return this.params.project
    }
  },
  beforeMount() {
  },
  methods: {
    appClick(app) {
      this.$router.push({name: 'workspaceAppDetail', params: {'appId': app.id, workspaceId: this.projectId}});
    }
  }
}
</script>

<style scoped>

</style>