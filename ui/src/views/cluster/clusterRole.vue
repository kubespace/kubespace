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
    return {
      scope: "cluster",
      maxHeight: window.innerHeight - this.$contentHeight,

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
