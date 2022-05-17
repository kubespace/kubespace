<template>
  <div>
    <workload-detail :kind='kind' :name='name'></workload-detail>
  </div>
</template>

<script>
import { WorkloadDetail } from '@/views/components'

export default {
  name: 'DeploymentDetail',
  components: {
    WorkloadDetail
  },
  data() {
    return {
      kind: 'deployment',
    }
  },
  computed: {
    name: function() {
      return this.$route.params ? this.$route.params.deploymentName : ''
    },
  }
}
</script>

<style lang="scss" scoped>

</style>
