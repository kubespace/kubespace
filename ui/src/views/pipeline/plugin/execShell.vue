<template>
  <div >
    <el-form :model="params" ref="job" label-position="left" label-width="105px">
      <el-form-item label="目标资源" prop="" :required="true">
        <el-input style="width: 250px;" v-model="params.resource" autocomplete="off" size="small"
          placeholder=""></el-input>
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

export default {
  name: 'ExecuteShell',
  data() {
    return {
    }
  },
  props: ['params'],
  beforeMount() {
    if(!this.params.shell) {
      this.$set(this.params, 'shell', 'bash')
    }
  }
}
</script>

<style scoped>

</style>