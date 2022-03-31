<template>
  <div class="template-class">
    <el-form label-position="left" :model="template.metadata" :rules="rules" style="padding: 10px 20px 10px; color: #909399" label-width="100px">
      <el-form-item label="名称" style="width: 500px" prop="name">
        <el-input v-model="template.metadata.name" placeholder="只能包含小写字母数字以及-和.,数字或者字母开头或结尾" size="small"></el-input>
      </el-form-item>
      <el-form-item label="类型" style="width: 500px" prop="name">
        <el-radio-group v-model="template.type"  size="small">
          <el-radio-button label="Opaque">Opaque</el-radio-button>
          <el-radio-button label="kubernetes.io/tls">TLS</el-radio-button>
          <el-radio-button label="kubernetes.io/basic-auth">用户名/密码</el-radio-button>
          <el-radio-button label="kubernetes.io/dockerconfigjson">镜像服务</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="template.type == 'Opaque'" label="配置项" required >
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
          <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px; border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
            @click="template.data.push({})" icon="el-icon-plus">添加配置项</el-button>
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item v-if="template.type == 'kubernetes.io/tls'" label="证书" style="width: 60%" required>
        <el-input type="textarea" v-model="template.tls.crt" placeholder="" size="small"></el-input>
      </el-form-item>
      <el-form-item v-if="template.type == 'kubernetes.io/tls'" label="密钥" style="width: 60%" required>
        <el-input type="textarea" v-model="template.tls.key" placeholder="" size="small"></el-input>
      </el-form-item>
      <el-form-item v-if="template.type == 'kubernetes.io/basic-auth'" label="用户" style="width: 500px" required>
        <el-input v-model="template.userPass.username" placeholder="" size="small"></el-input>
      </el-form-item>
      <el-form-item v-if="template.type == 'kubernetes.io/basic-auth'" label="密码" style="width: 500px;" required>
        <el-input type="password" autocomplete="new-password" v-model="template.userPass.password" placeholder="" size="small"></el-input>
      </el-form-item>
      <el-form-item v-if="template.type == 'kubernetes.io/dockerconfigjson'" label="仓库地址" style="width: 500px" required>
        <el-input v-model="template.imagePass.url" placeholder="镜像仓库地址" size="small"></el-input>
      </el-form-item>
      <el-form-item v-if="template.type == 'kubernetes.io/dockerconfigjson'" label="用户" style="width: 500px;" required>
        <el-input v-model="template.imagePass.username" placeholder="镜像仓库认证用户" size="small"></el-input>
      </el-form-item>
      <el-form-item v-if="template.type == 'kubernetes.io/dockerconfigjson'" label="密码" style="width: 500px" required>
        <el-input type="password" autocomplete="new-password" v-model="template.imagePass.password" placeholder="镜像仓库认证密码" size="small"></el-input>
      </el-form-item>
      <el-form-item v-if="template.type == 'kubernetes.io/dockerconfigjson'" label="邮箱" style="width: 500px;" required>
        <el-input v-model="template.imagePass.email" placeholder="用户邮箱" size="small"></el-input>
      </el-form-item>
    </el-form>
    
  </div>
</template>

<script>

export default {
  name: 'Secret',
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