<template>
  <div >
    <el-form :model="params" ref="stage" label-position="left" label-width="105px">
      <el-form-item label="编译镜像" prop="" :required="true">
        <el-input style="width: 320px;" v-model="params.code_build_image" autocomplete="off" size="small"></el-input>
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
      <el-divider></el-divider>
      <el-form-item label="镜像仓库" prop="" :required="true">
        <el-input style="width: 320px" v-model="params.image_build_server" autocomplete="off" size="small"></el-input>
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
        <el-row :gutter="20" v-for="(build,i) in params.image_build_infos" :key="i">
          <el-col :span="8">
            <el-input v-model="build.docker_file_path" placeholder="默认为Dockerfile" autocomplete="off" size="small"></el-input>
          </el-col>
          <el-col :span="13">
            <el-input v-model="build.image_name" autocomplete="off" size="small"></el-input>
          </el-col>
          <el-col :span="3">
            <span v-if="params.image_build_infos.length > 1" class="build-info-operator"
              style="margin-right: 5px;" @click="params.image_build_infos.splice(i, 1)">—</span>
            <span @click="addBuildInfo()" v-if="i +1 == params.image_build_infos.length"
              class="build-info-operator">+</span>
          </el-col>
        </el-row>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>

export default {
  name: 'CodeToImage',
  data() {
    return {
      image_builds: [],
      // params: {},
    }
  },
  props: ['params'],
  beforeMount() {
    // console.log(this.params)
    // let params = this.params
    if(!this.params.code_build_type) {
      // this.params.code_build_type = 'file'
      this.$set(this.params, 'code_build_type', 'file')
    }
    if(!this.params.image_build_infos) {
      this.$set(this.params, 'image_build_infos', [{
        'dockerfile': '',
        'image_name': ''
      }])
    }
    // this.params = this.params
    // this.image_builds = this.params.image_build_infos
    // console.log(this.job)
  },
  methods: {
    addBuildInfo() {
      // let infos = this.params.image_build_infos
      this.params.image_build_infos.push({
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
      console.log(this.job)
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