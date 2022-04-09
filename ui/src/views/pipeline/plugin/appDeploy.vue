<template>
  <div v-loading="loading">
    <el-form :model="params" ref="job" label-position="left" label-width="105px">
      <el-form-item label="工作空间" prop="" :required="true">
        <!-- <el-input style="width: 250px;" v-model="params.resource" autocomplete="off" size="small"
          placeholder=""></el-input> -->
        <el-select v-model="params.project" placeholder="请选择工作空间" size="small" style="width: 320px"
          @change="fetchApps">
          <el-option
            v-for="res in projects"
            :key="res.id"
            :label="res.name"
            :value="res.id">
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="应用" prop="">
        <el-select v-model="params.apps" placeholder="请选择要更新的应用" size="small" style="width: 320px" multiple>
          <el-option
            v-for="res in apps"
            :key="res.id"
            :label="res.name"
            :value="res.id">
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="是否部署" prop="">
        <el-switch v-model="params.with_install"></el-switch>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import { listProjects } from "@/api/project/project";
import { listApps } from "@/api/project/apps";

export default {
  name: 'ExecuteShell',
  data() {
    return {
      loading: false,
      resources: [],
      projects: [],
      apps: [],
    }
  },
  props: ['params'],
  computed: {
    workspaceId() {
      return this.$route.params.workspaceId
    },
  },
  beforeMount() {
    if(this.params.with_install == undefined) {
      this.$set(this.params, 'with_install', true)
    }
    if(this.params.apps == undefined) {
      this.$set(this.params, 'apps', [])
    }
  },
  mounted() {
    this.fetchProjects()
  },
  methods: {
    fetchProjects() {
      listProjects().then((resp) => {
        this.projects = resp.data ? resp.data : []
        this.projects.sort((a, b) => {return a.name > b.name ? 1 : -1})
      }).catch((err) => {
        console.log(err)
      })
    },
    fetchApps() {
      this.apps = []
      this.$set(this.params, 'apps', [])
      listApps({scope_id: this.params.project, scope: "project_app"}).then((resp) => {
        let originApps = resp.data ? resp.data : []
        this.apps = originApps
      }).catch((err) => {
      })
    },
  }
}
</script>

<style scoped>

</style>