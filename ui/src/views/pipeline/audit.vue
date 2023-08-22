<template>
  <div>
    <audit :scope="scope" :scopeId="scopeId"></audit>
  </div>
</template>
<script>
import { Audit } from "@/views/audit";

export default {
  name: "pipelineAudit",
  components: {
    Audit
  },
  data() {
    let workspaceId = parseInt(this.$route.params.workspaceId)
    return {
      scope: "pipeline",
      scopeId: workspaceId,
    };
  },
};
</script>
