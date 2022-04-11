<template>
  <div v-loading="loading">
    <el-form :model="params" ref="job" label-position="left" label-width="105px">
      <span>发布插件：发布后会根据执行输入的版本号，给代码仓库打上版本标签，同时会给构建出来的镜像打上版本标签，并推送到镜像仓库</span>
    </el-form>
  </div>
</template>

<script>

export default {
  name: 'Release',
  data() {
    return {
      loading: false,
    }
  },
  props: ['params'],
  computed: {
  },
  beforeMount() {
  },
  mounted() {
  },
  methods: {
  }
}
</script>

<style scoped>

</style>