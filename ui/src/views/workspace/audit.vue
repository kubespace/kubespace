<template>
  <div>
    <audit :scope="scope" :scopeId="scopeId"></audit>
  </div>
</template>
<script>
import { Audit } from "@/views/audit";

export default {
  name: "projectAudit",
  components: {
    Audit
  },
  data() {
    let workspaceId = parseInt(this.$route.params.workspaceId)
    return {
      scope: "project",
      scopeId: workspaceId,
    };
  },
  watch: {
    scopeId: function() {
      // this.fetchData()
    }
  },
};
</script>
