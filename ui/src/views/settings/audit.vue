<template>
  <div>
    <audit :scope="scope" :scopeId="scopeId"></audit>
  </div>
</template>
<script>
import { Audit } from "@/views/audit";

export default {
  name: "platformAudit",
  components: {
    Audit
  },
  data() {
    return {
      scope: "",
      scopeId: 0
    };
  },
};
</script>
