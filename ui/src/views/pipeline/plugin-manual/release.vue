<template>
  <div v-loading="loading">
    <el-form :model="params" ref="job" label-position="left" label-width="105px">
      <el-form-item label="发布版本" prop="" :required="true">
        <el-input style="width: 100%;" placeholder="请输入发布版本" v-model="params.version" autocomplete="off" size="small"></el-input>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import { getLatestRelease } from "@/api/pipeline/workspace";

export default {
  name: 'ManualRelease',
  data() {
    return {
      loading: false,
      latestRelease: {}
    }
  },
  props: ['pipeline', 'params'],
  computed: {
    workspaceId() {
      return this.$route.params.workspaceId
    },
  },
  beforeMount() {
    if(!this.params.version) {
      this.fetchLatestRelease()
    }
  },
  methods: {
    fetchLatestRelease() {
      // this.loading = true
      getLatestRelease({workspace_id: this.workspaceId}).then((resp) => {
        this.latestRelease = resp.data ? resp.data : {}
        let version = '1.0.0'
        if(this.latestRelease.release_version) {
          let latestVersion = this.latestRelease.release_version
          if(latestVersion.indexOf(".") > -1) {
            let splitVersion = latestVersion.split(".")
            let lastVersion = splitVersion[splitVersion.length - 1]
            splitVersion.splice(splitVersion.length - 1, 1)
            if(!isNaN(lastVersion)) {
              lastVersion = parseInt(lastVersion) + 1
            }
            splitVersion.push(lastVersion)
            version = splitVersion.join(".")
          }
        }
        this.$set(this.params, 'version', version)
        // this.loading = false
      }).catch((err) => {
        // this.loading = false
      })
    },
  }
}
</script>

<style scoped>

</style>