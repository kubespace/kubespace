<template>
  <div>
    <scope-role :scope="scope" :scopeId="scopeId"></scope-role>
  </div>
</template>
<script>
import { Clusterbar, ScopeRole } from "@/views/components";

export default {
  name: "pipelineRole",
  components: {
    Clusterbar,
    ScopeRole
  },
  data() {
    let workspaceId = parseInt(this.$route.params.workspaceId)
    return {
      scope: "pipeline",
      scopeId: workspaceId,
    };
  },
  computed: {
  },
  methods: {
    
  },
};
</script>


<style lang="scss" scoped>
@import "~@/styles/variables.scss";

.table-fix {
  height: calc(100% - 100px);
}

</style>
