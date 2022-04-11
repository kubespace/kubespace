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
      <el-form-item label="阶段参数" prop="type">
        <el-row style="margin-bottom: 5px; margin-top: 8px;">
          <el-col :span="7" style="background-color: #F5F7FA; padding-left: 10px;">
            <div class="border-span-header">
              <span  class="border-span-content">*</span>参数
            </div>
          </el-col>
          <el-col :span="10" style="background-color: #F5F7FA">
            <div class="border-span-header">
              默认值
            </div>
          </el-col>
          <!-- <el-col :span="5"><div style="width: 100px;"></div></el-col> -->
        </el-row>
        <el-row style="padding-bottom: 5px;" v-for="(d, i) in stage.custom_params" :key="i">
          <el-col :span="7">
            <div class="border-span-header">
              <el-input v-model="d.param" size="small" style="padding-right: 10px" placeholder="阶段参数"></el-input>
            </div>
          </el-col>
          <el-col :span="10">
            <div class="border-span-header">
              <el-input v-model="d.value" size="small" placeholder="请输入参数默认值"></el-input>
            </div>
          </el-col>
          <el-col :span="2" style="padding-left: 10px">
            <el-button circle size="mini" style="padding: 5px;" 
              @click="stage.custom_params.splice(i, 1)" icon="el-icon-close"></el-button>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="17">
          <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px;
            border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
            @click="stage.custom_params.push({})" icon="el-icon-plus">添加自定义参数</el-button>
          </el-col>
        </el-row>
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