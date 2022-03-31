<template>
  <div class="template-class">
    <el-form label-position="left" :model="template.metadata" :rules="rules" style="padding: 10px 20px" label-width="100px">
      <el-form-item label="负载名称" style="width: 400px" prop="name">
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
      <el-form-item label="副本数" style="width: 400px" required v-if="template.kind == 'Deployment' || template.kind == 'StatefulSet'">
        <el-input-number v-model="template.spec.replicas" style="width: 200px;" size="small" :min="1" label="描述文字"></el-input-number>
      </el-form-item>
      <el-form-item label="更新策略" style="width: 100%" required v-if="template.kind == 'Deployment'">
        <span slot="label">更新策略
          <el-link icon="el-icon-question" :underline="false" target="_blank"
            href="https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/#strategy">
          </el-link>
        </span>
        <el-radio-group v-model="template.spec.strategy.type" size="small" name="strategy">
          <el-radio-button label="RollingUpdate"></el-radio-button>
          <el-radio-button label="Recreate"></el-radio-button>
        </el-radio-group>
        <div v-if="template.spec.strategy.type == 'RollingUpdate'" style="margin-top: 10px; color: #99a9bf; line-height: 1">
          <!-- <span style="margin-right: 10px;">最大不可用</span> -->
          <el-input v-model="template.spec.strategy.maxUnavailable" style="width: 200px;" size="small" :min="1" >
            <template slot="prepend">最大不可用</template>
          </el-input>
          <!-- <span style="margin-left: 25px;margin-right: 10px;">最大峰值</span> -->
          <el-input v-model="template.spec.strategy.maxUnavailable" style="width: 185px; margin-left: 10px;" size="small" :min="1">
            <template slot="prepend">最大峰值</template>
          </el-input>
        </div>
      </el-form-item>
      <el-form-item v-if="template.kind == 'CronJob'" label="定时配置" style="width: 400px" required>
        <span slot="label">定时配置
          <el-link icon="el-icon-question" :underline="false" target="_blank"
            href="https://kubernetes.io/zh/docs/concepts/workloads/controllers/cron-jobs/#cron-%E6%97%B6%E9%97%B4%E8%A1%A8%E8%AF%AD%E6%B3%95">
          </el-link>
        </span>
        <el-input v-model="template.spec.job.schedule" placeholder="请输入Cron定时配置" size="small"></el-input>
      </el-form-item>
      <el-form-item v-if="template.kind == 'CronJob'" label="并发策略" style="width: 400px">
        <el-radio-group v-model="template.spec.job.concurrencyPolicy"  size="small" name="concurrencyPolicy">
          <el-radio-button label="Allow"></el-radio-button>
          <el-radio-button label="Forbid"></el-radio-button>
          <el-radio-button label="Replace"></el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="Job配置" style="width: 70%" v-if="template.kind == 'CronJob' || template.kind == 'Job'">
        <!-- <el-input-number v-model="template.spec.job.backoffLimit" style="width: 200px;" size="small" :min="1" label="描述文字"></el-input-number> -->
        <span slot="label">Job配置
          <el-link icon="el-icon-question" :underline="false" target="_blank"
            href="https://kubernetes.io/zh/docs/concepts/workloads/controllers/job/#parallel-jobs">
          </el-link>
        </span>
        <el-row>
          <el-col :span="6" style="padding-right: 15px;">
            <div style="color: #8B959C; padding-bottom: 0px; line-height: 24px; margin-top: 0px;">
              重试次数
            </div>
            <el-input-number v-model="template.spec.job.backoffLimit" size="small" placeholder="默认为10">
              <template slot="suffix">次</template>
            </el-input-number>
          </el-col>
          <el-col :span="6" style="padding-right: 15px;">
            <div style="color: #8B959C; padding-bottom: 0px; line-height: 24px; margin-top: 0px;">
              并发数
            </div>
            <el-input-number v-model="template.spec.job.completions" size="small" placeholder="">
              <template slot="suffix"></template>  
            </el-input-number>
          </el-col>
          <el-col :span="6" style="padding-right: 15px;">
            <div style="color: #8B959C; padding-bottom: 0px; line-height: 24px; margin-top: 0px;">
              完成数
            </div>
            <el-input-number v-model="template.spec.job.parallelism" size="small" placeholder="">
              <template slot="suffix">秒</template>  
            </el-input-number>
          </el-col>
        </el-row>
      </el-form-item>
    </el-form>
    <el-tabs tab-position="left" style="" value="container">
      <el-tab-pane label="容器组" name="container">
        <container :template="template" :appResources="appResources"></container>
      </el-tab-pane>
      <el-tab-pane label="存储">
        <div class="border-workload-content">
          <pod-volume :template="template" :appResources="appResources"></pod-volume>
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
  props: ['template', 'noName', 'appResources'],
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