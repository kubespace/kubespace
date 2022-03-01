<template>
  <div class="card-div">
    <el-form :model="podSpec" :rules="rules" label-width="120px"
      label-position="left" size="small">
      <el-form-item label="DNS策略" prop="dnsPolicy">
        <el-select v-model="podSpec.dnsPolicy" size="small" class="input-class">
          <el-option label="ClusterFirst" value="ClusterFirst"></el-option>
          <el-option label="Default" value="Default"></el-option>
          <el-option label="ClusterFirstWithHostNet" value="ClusterFirstWithHostNet"></el-option>
          <el-option label="None" value="None"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="宿主机资源" prop="hostResource">
        <el-checkbox v-model="podSpec.hostNetwork">HostNetwork</el-checkbox>
        <el-checkbox v-model="podSpec.hostPID">HostPID</el-checkbox>
        <el-checkbox v-model="podSpec.hostIPC">HostIPC</el-checkbox>
      </el-form-item>
      <el-form-item label="主机名" prop="hostname">
        <el-input v-model="podSpec.hostname" size="small" class="input-class" placeholder=""></el-input>
      </el-form-item>
      <el-form-item label="子域名" prop="subdomain">
        <el-input v-model="podSpec.subdomain" size="small" class="input-class" placeholder=""></el-input>
      </el-form-item>
      <el-form-item label="主机别名" prop="hostAliases" style="font-size: 14px; color: #99a9bf">
        <el-row style="margin-bottom: -15px;" v-if="podSpec.hostAliases.length > 0">
          <el-col :span="7">
            <div class="border-span-header">
              <span  class="border-span-content">*</span>主机名
            </div>
          </el-col>
          <el-col :span="17">
            <div class="border-span-header">
              <span  class="border-span-content">*</span>IP地址
            </div>
          </el-col>
        </el-row>
        <el-row style="padding-top: 10px;" v-for="(item, idx) in podSpec.hostAliases" :key="idx">
          <el-col :span="7" style="padding-right: 10px;">
            <div class="border-span-header">
              <el-input v-model="item.hostnames" size="small" placeholder="如：foo.com"></el-input>
            </div>
          </el-col>
          <el-col :span="7" style="padding-right: 10px;">
            <div class="border-span-header">
              <el-input v-model="item.ip" size="small" placeholder="如：1.1.1.1"></el-input>
            </div>
          </el-col>
          <el-col :span="3">
            <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
              @click="podSpec.hostAliases.splice(idx, 1)" icon="el-icon-minus"></el-button>
          </el-col>
        </el-row>
        <el-button plain size="mini" @click="podSpec.hostAliases.push({})" icon="el-icon-plus"></el-button>
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