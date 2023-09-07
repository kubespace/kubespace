<template>
  <div v-if="result && result.data">
    <div v-for="(v, k) of result.data" :key="k">
      <span>{{ k }}={{ v }}</span>
    </div>
  </div>
</template>

<script>

export default {
  name: 'ExecShellResult',
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