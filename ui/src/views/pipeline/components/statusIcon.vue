<template>
  <i v-if="statusIconMap[status]" :class="statusIconMap[status]"></i>
  <svg-icon v-else-if="status == 'pause'" icon-class="pause" />
</template>

<script>

export default {
  name: 'StatusIcon',
  data() {
    var iconMap
    if(this.type=='job') {
      iconMap = {
        ok: "el-icon-circle-check",
        error: "el-icon-circle-close",
        wait: "el-icon-remove-outline",
        doing: "el-icon-refresh refresh-rotate",
        cancel: "el-icon-remove-outline",
        canceled: "el-icon-remove-outline",
      }
    } else {
      iconMap = {
        ok: "el-icon-success",
        error: "el-icon-error",
        wait: "el-icon-remove",
        doing: "el-icon-refresh refresh-rotate",
        cancel: "el-icon-remove",
        canceled: "el-icon-remove",
      }
    }
    return {
      statusIconMap: iconMap,
    }
  },
  props: ['status', 'type'],
  computed: {
  },
  beforeMount() {
  },
  methods: {
  }
}
</script>

<style scoped>

</style>