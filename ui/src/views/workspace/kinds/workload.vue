<template>
  <div class="template-class">
    <el-form label-position="left" :model="template.metadata" :rules="rules" style="padding: 10px 20px" label-width="100px">
      <el-form-item v-if="!noName" label="负载名称" style="width: 400px" prop="name">
        <el-input v-model="template.metadata.name" placeholder="请输入工作负载名称" size="small"></el-input>
      </el-form-item>
      <el-form-item label="负载类型" style="width: 700px" prop="" required>
        <el-radio-group v-model="template.kind"  size="small">
          <el-radio-button label="Deployment"></el-radio-button>
          <el-radio-button label="StatefulSet"></el-radio-button>
          <el-radio-button label="DaemonSet"></el-radio-button>
          <el-radio-button label="CronJob"></el-radio-button>
          <el-radio-button label="Job"></el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="副本数" style="width: 400px" required>
        <el-input-number v-model="template.spec.replicas" style="width: 200px;" size="small" :min="1" label="描述文字"></el-input-number>
      </el-form-item>
    </el-form>
    <el-tabs tab-position="left" style="" value="container">
      <el-tab-pane label="容器组" name="container">
        <container :template="template"></container>
      </el-tab-pane>
      <el-tab-pane label="存储">
        <div class="border-workload-content">
          <pod-volume :template="template"></pod-volume>
        </div>
      </el-tab-pane>
      <el-tab-pane label="网络">
        <div class="border-workload-content">
          <pod-network :template="template"></pod-network>
        </div>
      </el-tab-pane>
      <el-tab-pane label="调度">
        <div class="border-workload-content">
          <pod-affinity :template="template"></pod-affinity>
        </div>
      </el-tab-pane>
      <el-tab-pane label="安全">
        <div class="border-workload-content">
          <pod-security :template="template"></pod-security>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import { Container, PodVolume, PodNetwork, PodAffinity, PodSecurity } from '@/views/workspace/kinds'
import Pod_affinity from './pod_affinity.vue'
import Pod_security from './pod_security.vue'

export default {
  name: 'Workload',
  components: {
    Container,
    PodVolume,
    PodNetwork,
    PodAffinity,
    PodSecurity
  },
  data() {
    return {
      rules: {
        'name': [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
      },
    }
  },
  props: ['template', 'noName'],
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

.border-workload-content {
  margin: 0px 20px 0px 10px;
  padding: 15px;
  border: 1px solid #DCDFE6;
  box-shadow: 0 2px 4px 0 rgb(0 0 0 / 12%), 0 0 6px 0 rgb(0 0 0 / 4%);
}
</style>

<style lang="scss">
.template-class {
  .el-tabs--border-card>.el-tabs__content {
    padding: 15px;
  }
}
</style>