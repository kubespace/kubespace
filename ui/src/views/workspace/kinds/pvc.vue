<template>
  <div class="template-class">
    <el-form label-position="left" :model="template.metadata" :rules="rules" style="padding: 10px 20px 10px; color: #909399" label-width="100px">
      <el-form-item label="名称" style="width: 500px" prop="name">
        <el-input v-model="template.metadata.name" placeholder="只能包含小写字母数字以及-和.,数字或者字母开头或结尾" size="small"></el-input>
      </el-form-item>
      <el-form-item label="申请大小" style="width: 500px" required>
        <el-input v-model="template.spec.resources.requests.storage" style="width: 139px;"
          placeholder="存储卷申请大小" size="small">
          <span slot="suffix" style="padding-right: 5px;" >Gi </span>
        </el-input>
      </el-form-item>
      <el-form-item label="访问模式" style="width: 500px;">
        <el-select v-model="template.spec.accessModes" placeholder="访问模式" style="width: 100%;" multiple size="small" >
          <el-option label="ReadWriteOnce" value="ReadWriteOnce"></el-option>
          <el-option label="ReadWriteMany" value="ReadWriteMany"></el-option>
          <el-option label="ReadOnlyMany" value="ReadOnlyMany"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="存储类" style="width: 500px;">
        <el-input v-model="template.spec.storageClassName" placeholder="存储类" size="small"></el-input>
      </el-form-item>
    </el-form>
    
  </div>
</template>

<script>

export default {
  name: 'pvc',
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