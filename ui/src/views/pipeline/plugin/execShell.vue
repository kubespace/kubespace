<template>
  <div v-loading="loading">
    <el-form :model="params" ref="job" label-position="left" label-width="105px">
      <el-form-item label="目标资源" prop="" :required="true">
        <!-- <el-input style="width: 250px;" v-model="params.resource" autocomplete="off" size="small"
          placeholder=""></el-input> -->
        <el-select v-model="params.resource" placeholder="请选择要执行的目标资源" size="small" style="width: 320px" 
          >
          <el-option :key="0" label="Spacelet节点" value=""></el-option>
          <el-option
            v-for="res in resources"
            :key="res.id"
            :label="res.name"
            :value="res.id">
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="脚本类型" prop="">
        <el-radio-group v-model="params.shell">
          <el-radio label="bash">bash</el-radio>
          <el-radio label="sh">sh</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="执行脚本" prop="">
        <el-input type="textarea" :rows="6" v-model="params.script" autocomplete="off" size="small"></el-input>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import { listResources } from "@/api/pipeline/resource";

export default {
  name: 'ExecuteShell',
  data() {
    return {
      loading: false,
      resources: [],
      resource: this.params.resource ? this.params.resource : 0,
    }
  },
  props: ['params'],
  computed: {
    workspaceId() {
      return this.$route.params.workspaceId
    },
  },
  beforeMount() {
    if(!this.params.shell) {
      this.$set(this.params, 'shell', 'bash')
    }
    this.fetchResources()
  },
  methods: {
    fetchResources() {
      this.loading = true
      listResources(this.workspaceId).then((resp) => {
        this.resources = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        this.loading = false
      })
    },
    // resourceSelectChanged(res) {
    //   console.log(res)
    //   if(res == 0) {
    //     this.pa
    //   }
    // },
  }
}
</script>

<style scoped>

</style>