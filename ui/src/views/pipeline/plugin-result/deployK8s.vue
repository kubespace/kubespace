<template>
  <div v-if="result && result.data && result.data.cluster">
    <div>
      <span>集群：</span>
      <span>{{ cluster }}</span>
    </div>
    <div>
      <span>命名空间：</span>
      <span>{{ namespace }}</span>
    </div>
    <div>
      <span>部署资源：</span>
      <span>{{ resources.join(", ") }}</span>
    </div>
    <div v-if="imageRes && imageRes.length > 0">
      <span>部署镜像：</span>
      <div v-for="(res, i) of imageRes" :key="i">
        {{ res.kind }}/{{ res.name }} ( {{ res.images.join(", ") }} )
      </div>
    </div>
  </div>
</template>

<script>

export default {
  name: 'DeployK8sResult',
  data() {
    return {
    }
  },
  props: ['result', 'params'],
  computed: {
    cluster() {
      if(!this.result.data) return ''
      if(this.result.data.cluster) return this.result.data.cluster
      return ''
    },
    namespace() {
      if(!this.params) return ''
      return this.params.namespace
    },
    resources() {
      if(!this.result.data) return []
      if(!this.result.data.resources) return []
      let res = []
      for(let r of this.result.data.resources) {
        if(!r.images || r.images.length == 0) res.push(r.kind + "/" + r.name)
      }
      return res
    },
    imageRes() {
      if(!this.result.data) return []
      if(!this.result.data.resources) return []
      let res = []
      for(let r of this.result.data.resources) {
        if(r.images && r.images.length > 0) res.push(r)
      }
      return res
    },
  },
  beforeMount() {
  },
  methods: {
  }
}
</script>

<style scoped>

</style>