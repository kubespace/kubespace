<template>
  <div>
    <workloadCreate :workType="workType"></workloadCreate>
  </div>
</template>

<script>
import { WorkloadCreate } from '@/views/components'

export default {
  name: 'Deployment',
  components: {
    WorkloadCreate
  },
  data() {
    return {
      workType: "Deployments"
    }
  }
}
</script>
