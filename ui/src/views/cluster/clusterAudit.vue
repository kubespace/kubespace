<template>
  <div>
    <audit :scope="scope" :scopeId="scopeId"></audit>
  </div>
</template>
<script>
import { Audit } from "@/views/audit";

export default {
  name: "clusterAudit",
  components: {
    Audit
  },
  data() {
    return {
      scope: "cluster",
    };
  },
  watch: {
    scopeId: function() {
      // this.fetchData()
    }
  },
  computed: {
    scopeId() {
      return parseInt(this.$store.state.cluster)
    }
  },
};
</script>
