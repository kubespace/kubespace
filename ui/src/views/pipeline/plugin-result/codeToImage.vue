<template>
  <div>
    <div v-for="(img, i) in images" :key="i">
      {{ img }}
    </div>
  </div>
</template>

<script>

export default {
  name: 'CodeToImageResult',
  data() {
    return {
    }
  },
  props: ['result'],
  computed: {
    workspaceId() {
      return this.$route.params.workspaceId
    },
    images() {
      if(!this.result.data.images) {
        return []
      }
      let images = []
      try{
        let resImg = JSON.parse(this.result.data.images)
        for(let i in resImg) {
          images.push(resImg[i])
        }
      } catch(e) {
        console.log('parse release images error: ', e)
      }
      return images
    }
  },
  beforeMount() {
  },
  methods: {
  }
}
</script>

<style scoped>

</style>