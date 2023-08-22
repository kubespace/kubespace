<template>
  <div>
    <scope-role :scope="scope" :scopeId="scopeId"></scope-role>
  </div>
</template>
<script>
import { Clusterbar, ScopeRole } from "@/views/components";

export default {
  name: "WorkspaceRole",
  components: {
    Clusterbar,
    ScopeRole
  },
  data() {
    let workspaceId = parseInt(this.$route.params.workspaceId)
    return {
      scope: "project",
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
