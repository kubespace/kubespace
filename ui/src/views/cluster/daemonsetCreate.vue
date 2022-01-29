<template>
  <div>
    <workloadCreate :workType="workType"></workloadCreate>
  </div>
</template>

<script>
import { WorkloadCreate } from '@/views/components'

export default {
  name: 'Daemonsets',
  components: {
    WorkloadCreate
  },
  data() {
    return {
      workType: "DaemonSets"
    }
  }
}
</script>
