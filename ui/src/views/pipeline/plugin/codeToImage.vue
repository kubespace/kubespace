<template>
  <div v-loading="loading">
    <el-form :model="params" ref="stage" label-position="left" label-width="105px">
      <el-form-item label="编译" prop="">
        <el-switch v-model="params.code_build"></el-switch>
      </el-form-item>
      <template v-if="params.code_build">
        <el-form-item label="编译镜像" prop="" :required="true">
          <!-- <el-input style="width: 320px;" v-model="params.code_build_image" autocomplete="off" size="small"></el-input> -->
          <el-select v-model="params.code_build_image" placeholder="请选择编译镜像" size="small" style="width: 320px">
            <el-option
              v-for="res in imageResources"
              :key="res.id"
              :label="res.name"
              :value="res.id">
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="编译方式" prop="">
          <el-radio-group v-model="params.code_build_type">
            <el-radio label="file">脚本文件</el-radio>
            <el-radio label="script">自定义脚本</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="编译文件" prop="" v-if="params.code_build_type == 'file'">
          <el-input style="width: 320px;" v-model="params.code_build_file" autocomplete="off" size="small"
            placeholder="默认为当前代码库根目录下的「build.sh」文件"></el-input>
        </el-form-item>
        <el-form-item label="编译脚本" prop="" v-if="params.code_build_type == 'script'">
          <el-input type="textarea" :rows="6" v-model="params.code_build_script" autocomplete="off" size="small"></el-input>
        </el-form-item>
      </template>
      <el-divider></el-divider>
      <el-form-item label="镜像仓库" prop="" :required="true">
        <!-- <el-input style="width: 320px" v-model="params.image_build_server" autocomplete="off" size="small"></el-input> -->
        <el-select v-model="params.image_build_registry" placeholder="请选择镜像要推送的仓库" size="small" style="width: 320px">
          <el-option
            v-for="res in registry"
            :key="res.id"
            :label="res.registry"
            :value="res.id">
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="构建镜像" prop="" :required="true">
        <el-row :gutter="20">
          <el-col :span="8">
            <span>Dockerfile</span>
          </el-col>
          <el-col :span="13">
            <span>镜像名称</span>
          </el-col>
          <el-col :span="3"></el-col>
        </el-row>
        <el-row :gutter="20" v-for="(build,i) in params.image_builds" :key="i">
          <el-col :span="8">
            <el-input v-model="build.dockerfile" placeholder="默认为Dockerfile" autocomplete="off" size="small"></el-input>
          </el-col>
          <el-col :span="13">
            <el-input v-model="build.image" autocomplete="off" size="small"></el-input>
          </el-col>
          <el-col :span="3">
            <span v-if="params.image_builds.length > 1" class="build-info-operator"
              style="margin-right: 5px;" @click="params.image_builds.splice(i, 1)">—</span>
            <span @click="addBuildInfo()" v-if="i +1 == params.image_builds.length"
              class="build-info-operator">+</span>
          </el-col>
        </el-row>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import { listResources } from "@/api/pipeline/resource";
import { listImageRegistry } from "@/api/settings/image_registry";

export default {
  name: 'CodeToImage',
  data() {
    return {
      image_builds: [],
      // params: {},
      resources: [],
      registry: [],
      loading: false,
    }
  },
  props: ['params',],
  computed: {
    workspaceId() {
      return this.$route.params.workspaceId
    },
    imageResources() {
      let res = []
      for(let r of this.resources) {
        if(r.type == 'image') {
          res.push(r)
        }
      }
      return res
    }
  },
  beforeMount() {
    // console.log(this.params)
    // let params = this.params
    if(this.params.code_build == undefined) {
      this.$set(this.params, 'code_build', true)
    }
    if(!this.params.code_build_type) {
      // this.params.code_build_type = 'file'
      this.$set(this.params, 'code_build_type', 'file')
    }
    if(!this.params.image_builds) {
      this.$set(this.params, 'image_builds', [{
        'dockerfile': '',
        'image': ''
      }])
    }
    this.fetchResources()
    this.fetchImageRegistry()
    // this.params = this.params
    // this.image_builds = this.params.image_build_infos
    // console.log(this.job)
  },
  methods: {
    fetchResources() {
      this.loading = true
      listResources(this.workspaceId).then((resp) => {
        this.resources = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        this.loading = false
      })
    },
    fetchImageRegistry() {
      this.loading = true
      listImageRegistry().then((resp) => {
        this.registry = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    addBuildInfo() {
      // let infos = this.params.image_build_infos
      this.params.image_builds.push({
        'dockerfile': '',
        'image_name': ''
      })
      // this.$set(this.params, 'image_build_infos', infos)
      // let len = this.image_builds.length
      // this.image_builds.splice(len, 0, {
      //   'dockerfile': '',
      //   'image_name': ''
      // })
      // this.image_builds.splice(len)
    }
  }
}
</script>

<style scoped>
.build-info-operator{
  border: 1px solid #DCDFE6; 
  padding: 3px 6px; 
  color: #909399;
}
</style>