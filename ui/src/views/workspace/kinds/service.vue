<template>
  <div class="template-class">
    <el-form label-position="left" :model="template.metadata" :rules="rules" style="padding: 10px 20px" label-width="100px">
      <el-form-item label="服务名称" style="width: 400px" prop="name">
        <el-input v-model="template.metadata.name" placeholder="请输入服务名称" size="small"></el-input>
      </el-form-item>
      <el-form-item label="服务类型" style="width: 700px" prop="" required>
        <el-radio-group v-model="template.spec.type"  size="small">
          <el-radio-button label="ClusterIP"></el-radio-button>
          <el-radio-button label="NodePort"></el-radio-button>
          <el-radio-button label="LoadBalancer"></el-radio-button>
          <el-radio-button label="ExternalName"></el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="副本数" style="width: 400px" required>
        <el-input-number v-model="template.spec.replicas" style="width: 200px;" size="small" :min="1" label="描述文字"></el-input-number>
      </el-form-item>
    </el-form>
    
  </div>
</template>

<script>
import { newContainer } from '@/views/workspace/kinds'

export default {
  name: 'Service',
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

</style>

<style lang="scss">
</style>