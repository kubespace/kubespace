<template>
  <div class="template-class">
    <el-form label-position="left" :model="template.metadata" :rules="rules" style="padding: 10px 20px 10px; color: #909399" label-width="100px">
      <el-form-item label="名称" style="width: 500px" prop="name">
        <el-input v-model="template.metadata.name" placeholder="只能包含小写字母数字以及-和.,数字或者字母开头或结尾" size="small"></el-input>
      </el-form-item>
      <el-form-item label="配置项" required >
        <el-row style="margin-bottom: 5px; margin-top: 8px;">
          <el-col :span="7" style="background-color: #F5F7FA; padding-left: 10px;">
            <div class="border-span-header">
              <span  class="border-span-content">*</span>Key
            </div>
          </el-col>
          <el-col :span="10" style="background-color: #F5F7FA">
            <div class="border-span-header">
              Value
            </div>
          </el-col>
          <!-- <el-col :span="5"><div style="width: 100px;"></div></el-col> -->
        </el-row>
        <el-row style="padding-bottom: 5px;" v-for="(d, i) in template.data" :key="i">
          <el-col :span="7">
            <div class="border-span-header">
              <el-input v-model="d.key" size="small" style="padding-right: 10px" placeholder="配置项Key"></el-input>
            </div>
          </el-col>
          <el-col :span="10">
            <div class="border-span-header">
              <el-input type="textarea" style="border-radius: 0px;" v-model="d.value" size="small" placeholder="配置项Value"></el-input>
            </div>
          </el-col>
          <el-col :span="2" style="padding-left: 10px">
            <el-button circle size="mini" style="padding: 5px;" 
              @click="template.data.splice(i, 1)" icon="el-icon-close"></el-button>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="17">
          <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px;
            border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
            @click="template.data.push({})" icon="el-icon-plus">添加配置项</el-button>
          </el-col>
        </el-row>
      </el-form-item>
    </el-form>
    
  </div>
</template>

<script>

export default {
  name: 'ConfigMap',
  components: {
  },
  data() {
    return {
      rules: {
        'name': [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
      },
    }
  },
  props: ['template',],
  computed: {
    
  },
  methods: {
    
  }
}
</script>

<style scoped lang="scss">

.border-span-content {
  color: #F56C6C; 
  margin-right: 4px;
}
</style>

<style lang="scss">
</style>