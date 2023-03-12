<template>
  <i v-if="status != 'pause'" :class="statusIconMap[status]"></i>
  <svg-icon v-else icon-class="pause" />
</template>

<script>

export default {
  name: 'StatusIcon',
  data() {
    return {
      statusIconMap: {
        ok: "el-icon-success",
        error: "el-icon-error",
        wait: "el-icon-remove",
        doing: "el-icon-refresh refresh-rotate",
        cancel: "el-icon-remove",
        canceled: "el-icon-remove",
      },
    }
  },
  props: ['status'],
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