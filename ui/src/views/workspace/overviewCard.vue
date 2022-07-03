<template>
    <el-card shadow="never" :style="style" class="overview-card">
        <div class="overview-card_left">
        <svg-icon :icon-class="icon"></svg-icon>
        </div>
        <div class="overview-card_right">
        <div class="overview-card_right_number">{{ number }}</div>
        <div class="overview-card_right_tilte">{{ title }}</div>
        </div>
    </el-card>
</template>
<script>
export default {
  name: 'OverviewCard',
  props: {
    icon: {
      type: String,
      default: ''
    },
    title: {
      type: String,
      default: ''
    },
    number: {
        type: Number,
        default: 0
    },
    style: {
        type: String,
        default: ''
    }
  },
  mounted() {}   
}
</script>
