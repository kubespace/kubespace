<template>
  <div>
    <workload-detail :kind='kind' :name='name'></workload-detail>
  </div>
</template>

<script>
import { WorkloadDetail } from '@/views/components'

export default {
  name: 'DaemonSetDetail',
  components: {
    WorkloadDetail
  },
  data() {
    return {
      kind: 'daemonset',
    }
  },
  computed: {
    name: function() {
      return this.$route.params ? this.$route.params.daemonsetName : ''
    },
  }
}
</script>

<style lang="scss" scoped>

</style>
