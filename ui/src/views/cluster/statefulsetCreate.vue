<template>
  <div>
    <workloadCreate :workType="workType"></workloadCreate>
  </div>
</template>

<script>
import { WorkloadCreate } from '@/views/components'

export default {
  name: 'Statefulset',
  components: {
    WorkloadCreate
  },
  data() {
    return {
      workType: "StatefulSets"
    }
  }
}
</script>
