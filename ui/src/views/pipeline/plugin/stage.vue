<template>
  <div >
    <el-form :model="stage" ref="stage" label-position="left" label-width="105px">
      <el-form-item label="阶段名称" prop="name" :required="true">
        <el-input style="width: 250px;" v-model="stage.name" autocomplete="off" size="small"
          placeholder="未命名"></el-input>
      </el-form-item>
      <el-form-item label="触发方式" prop="type">
        <el-radio-group v-model="stage.trigger_mode">
          <el-radio label="auto">自动触发</el-radio>
          <el-radio label="manual">手动触发</el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>

export default {
  name: 'PipelineStage',
  data() {
    return {
    }
  },
  props: ['stage']
}
</script>

<style scoped>

</style>