<template>
  <component :is="type" v-bind="linkProps(to)">
    <slot />
  </component>
</template>

<script>

export default {
  props: {
    to: {
      type: String,
      required: true
    }
  },
  computed: {
    type() {
      return 'router-link'
    }
  },
  methods: {
    linkProps(to) {
      return {
        'to': to
      }
    }
  }
}
</script>
