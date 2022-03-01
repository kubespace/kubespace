<template>
  <div class="card-div" style="font-size: 14px; color: #99a9bf">
    <el-form :model="podSpec" :rules="rules" label-width="120px"
        label-position="left" size="small">
      <el-form-item label="以用户ID运行" prop="runAsUser">
        <el-input v-model="podSpec.securityContext.runAsUser" size="small" class="input-class" placeholder="如：1001"></el-input>
      </el-form-item>
      <el-form-item label="以用户组ID运行" prop="runAsUser">
        <el-input v-model="podSpec.securityContext.runAsGroup" size="small" class="input-class" placeholder="如：1001"></el-input>
      </el-form-item>
      <el-form-item label="以非root运行" prop="runAsNonRoot">
        <el-switch v-model="podSpec.securityContext.runAsNonRoot"></el-switch>
      </el-form-item>
      <el-form-item label="sysctls" prop="sysctls">
        <div style="margin-bottom: 10px;" v-for="(l, i) in podSpec.securityContext.sysctls" :key="i">
          <el-input v-model="l.name" style="width: 150px;" placeholder="name"></el-input> = 
          <el-input v-model="l.value" class="input-class" placeholder="value"></el-input>
          <!-- <el-link :underline="false" style="margin-left: 10px;" @click="podSpec.labels.splice(i, 1)">删除</el-link> -->
          <el-button plain size="mini" style="margin-left: 10px; padding-left: 10px; padding-right: 10px;" 
            @click="podSpec.securityContext.sysctls.splice(i, 1)" icon="el-icon-minus"></el-button>
        </div>
        <el-button plain size="mini" @click="podSpec.securityContext.sysctls.push({})" icon="el-icon-plus"></el-button>
      </el-form-item>
      <el-form-item label="SELinux参数" prop="seLinuxOptions">
        <el-row>
          <el-col :span="4">
            <div class="item-header">
              User
            </div>
          </el-col>
          <el-col :span="4">
            <div class="item-header">
              Role
            </div>
          </el-col>
          <el-col :span="4">
            <div class="item-header">
              Type
            </div>
          </el-col>
          <el-col :span="10">
            <div class="item-header">
              Level
            </div>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="4" style="padding-right: 10px;">
            <div class="item-content">
              <el-input v-model="podSpec.securityContext.seLinuxOptions.user" size="small" placeholder=""></el-input>
            </div>
          </el-col>
          <el-col :span="4" style="padding-right: 10px;">
            <div class="item-content">
              <el-input v-model="podSpec.securityContext.seLinuxOptions.role" size="small" placeholder=""></el-input>
            </div>
          </el-col>
          <el-col :span="4" style="padding-right: 10px;">
            <div class="item-content">
              <el-input v-model="podSpec.securityContext.seLinuxOptions.type" size="small" placeholder=""></el-input>
            </div>
          </el-col>
          <el-col :span="4" style="padding-right: 10px;">
            <div class="item-content">
              <el-input v-model="podSpec.securityContext.seLinuxOptions.level" size="small" placeholder=""></el-input>
            </div>
          </el-col>
        </el-row>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>

export default {
  name: 'PodNetwork',
  components: {
  },
  data() {
    return {
      podSpec: this.template.spec.template.spec,
      rules: {
        'name': [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
      },
    }
  },
  props: ['template'],
  computed: {
  },
  methods: {
  }
}
</script>

<style scoped lang="scss">
.input-class {
  width: 300px;
}

.border-span-header {
  padding-bottom: 8px;
}

.border-span-content {
  color: #F56C6C; 
  margin-right: 4px;
}
.box-card {
  font-size: 14px;
  color: #99a9bf
}
</style>

<style lang="scss">

</style>