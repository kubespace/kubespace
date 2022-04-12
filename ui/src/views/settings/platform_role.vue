<template>
  <div>
    <scope-role :scope="scope" :scopeId="scopeId"></scope-role>
  </div>
</template>
<script>
import { Clusterbar, ScopeRole } from "@/views/components";

export default {
  name: "PlatformRole",
  components: {
    Clusterbar,
    ScopeRole
  },
  data() {
    return {
      scope: "platform",
      scopeId: 0,
      maxHeight: window.innerHeight - 135,

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
