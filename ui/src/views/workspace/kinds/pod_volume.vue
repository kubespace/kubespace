<template>
  <div>
    <el-card class="box-card" style="margin-bottom: 20px;" v-for="(vol, idx) in volumes" 
      :key="idx" :body-style="{padding: '0px'}">
      <div slot="header" class="clearfix">
        <el-form ref="form" :model="vol" :rules="rules" label-width="0px" size="mini">
          <el-row>
            <el-col :span="12">
              <div class="border-span-header">
                <span class="border-span-content">*</span>卷名称
              </div>
              <el-form-item label="" prop="name">
                <el-input v-model="vol.name" size="small" class="input-class" placeholder="卷名称，如：vol1"></el-input>
              </el-form-item>
            </el-col>
            <el-col :span="10">
              <div class="border-span-header">
                <span class="border-span-content">*</span>卷类型
              </div>
              <el-select v-model="vol.type" size="small" placeholder="卷类型" class="input-class">
                <el-option label="PersistentVolumeClaim" value="persistentVolumeClaim"></el-option>
                <el-option label="HostPath" value="hostPath"></el-option>
                <el-option label="EmptyDir" value="emptyDir"></el-option>
                <el-option label="ConfigMap" value="configMap"></el-option>
                <el-option label="Secret" value="secret"></el-option>
                <el-option label="NFS" value="nfs"></el-option>
                <el-option label="Glusterfs" value="glusterfs"></el-option>
              </el-select>
            </el-col>
            <el-col :span="2">
              <el-button style="float: right; padding: 3px 0" type="text"
                @click="volumes.splice(idx, 1)">删除</el-button>
            </el-col>
          </el-row>
        </el-form>
      </div>
      <div style="padding: 20px;" v-if="vol.type !== 'emptyDir'">
        <el-row v-if="vol.type === 'persistentVolumeClaim'">
          <el-col :span="12">
            <div class="border-span-header">
              <span class="border-span-content">*</span>存储声明
            </div>
            <el-select v-model="vol['persistentVolumeClaim'].claimName" size="small" placeholder="请选择存储声明" class="input-class">
              <el-option :label="`${p.name}`" :value="p.name" :key="i" v-for="(p, i) in nsPvcs"></el-option>
            </el-select>
          </el-col>
        </el-row>
        <el-row v-if="vol.type === 'hostPath'">
          <el-col :span="12">
            <div class="border-span-header">
              <span  class="border-span-content">*</span>宿主机目录
            </div>
            <el-input v-model="vol['hostPath'].path" size="small" class="input-class" placeholder="宿主机目录，如：/data"></el-input>
          </el-col>
          <el-col :span="12">
            <div class="border-span-header">类型</div>
            <el-select v-model="vol['hostPath'].type" size="small" placeholder="目录类型" class="input-class">
              <el-option label="默认" value=""></el-option>
              <el-option label="DirectoryOrCreate" value="DirectoryOrCreate"></el-option>
              <el-option label="Directory" value="Directory"></el-option>
              <el-option label="FileOrCreate" value="FileOrCreate"></el-option>
              <el-option label="File" value="File"></el-option>
              <el-option label="Socket" value="Socket"></el-option>
              <el-option label="CharDevice" value="CharDevice"></el-option>
              <el-option label="BlockDevice" value="BlockDevice"></el-option>
            </el-select>
          </el-col>
        </el-row>
        <div v-if="['configMap', 'secret'].indexOf(vol.type) >= 0">
          <el-row>
            <el-col :span="12">
              <div class="border-span-header">
                <span  class="border-span-content">*</span>{{ vol.type === 'configMap' ? '配置映射' : '密钥' }}
              </div>
              <el-select v-model="vol[vol.type].obj" value-key="name" size="small" :placeholder="`请选择${vol.type}`" class="input-class">
                
                <el-option :key="i" :label="s.name" :value="s" v-for="(s, i) in vol.type === 'configMap' ? nsConfigmaps : nsSecrets"></el-option>
              </el-select>
            </el-col>
            <el-col :span="12">
              <div class="border-span-header">
                默认模式
              </div>
              <el-input v-model="vol[vol.type].defaultMode" size="small" class="input-class" placeholder="创建文件的默认模式，默认为0644"></el-input>
            </el-col>
          </el-row>
          <el-row style="margin-bottom: -15px; padding-top: 10px;" v-if="vol[vol.type].items.length > 0">
            <el-col :span="7">
              <div class="border-span-header">
                <span  class="border-span-content">*</span>键
              </div>
            </el-col>
            <el-col :span="7">
              <div class="border-span-header">
                <span  class="border-span-content">*</span>路径
              </div>
            </el-col>
            <el-col :span="7">
              <div class="border-span-header">
                模式
              </div>
            </el-col>
          </el-row>
          <el-row style="padding-top: 10px;" v-for="(item, idx) in vol[vol.type].items" :key="idx">
            <el-col :span="7" style="padding-right: 10px;">
              <div class="border-span-header">
                <el-select v-model="item.key" size="small" :placeholder="`请选择${vol.type}中的键`" style="width: 100%;" >
                  <el-option :key="s" :label="s" :value="s" v-for="(v, s) in vol[vol.type].obj ? vol[vol.type].obj.data : {}"></el-option>
                </el-select>
              </div>
            </el-col>
            <el-col :span="7" style="padding-right: 10px;">
              <div class="border-span-header">
                <el-input v-model="item.path" size="small" placeholder="文件映射相对路径"></el-input>
              </div>
            </el-col>
            <el-col :span="7" style="padding-right: 10px;">
              <div class="border-span-header">
                <el-input v-model="item.mode" size="small" placeholder="文件模式"></el-input>
              </div>
            </el-col>
            <el-col :span="3">
              <el-button plain size="small" style="padding-left: 10px; padding-right: 10px;" 
                @click="vol[vol.type].items.splice(idx, 1)" icon="el-icon-minus"></el-button>
            </el-col>
          </el-row>
          <el-button plain size="small" @click="vol[vol.type].items.push({})" style="margin-top: 20px;">
            添加指定文件项
          </el-button>
        </div>
        <div v-if="vol.type === 'nfs'">
          <el-row>
            <el-col :span="8">
              <div class="border-span-header">
                <span  class="border-span-content">*</span>服务器地址
              </div>
              <el-input v-model="vol['nfs'].server" size="small" class="input-class" placeholder="NFS服务器地址"></el-input>
            </el-col>
            <el-col :span="8">
              <div class="border-span-header">
                <span  class="border-span-content">*</span>访问路径
              </div>
              <el-input v-model="vol['nfs'].path" size="small" class="input-class" placeholder="NFS服务访问路径"></el-input>
            </el-col>
            <el-col :span="8">
              <span class="border-span-header" style="margin-right: 25px;">
                只读
              </span>
              <el-switch v-model="vol['nfs'].readOnly"></el-switch>
            </el-col>
          </el-row>
        </div>
        <div v-if="vol.type === 'glusterfs'">
          <el-row>
            <el-col :span="8">
              <div class="border-span-header">
                <span  class="border-span-content">*</span>端点名称
              </div>
              <el-input v-model="vol['glusterfs'].endpoints" size="small" class="input-class" placeholder="Glusterfs端点名称"></el-input>
            </el-col>
            <el-col :span="8">
              <div class="border-span-header">
                <span  class="border-span-content">*</span>访问路径
              </div>
              <el-input v-model="vol['glusterfs'].path" size="small" class="input-class" placeholder="Glusterfs卷路径"></el-input>
            </el-col>
            <el-col :span="8">
              <span style="color: #8B959C; font-size: .85em; padding-bottom: 8px; margin-right: 25px;">
                只读
              </span>
              <el-switch v-model="vol['glusterfs'].readOnly"></el-switch>
            </el-col>
          </el-row>
        </div>
      </div>
    </el-card>
    <el-button plain size="small" @click="volumes.push(newPodVolume())">添加存储卷</el-button>
  </div>
</template>

<script>
import { newPodVolume } from '@/views/workspace/kinds'

export default {
  name: 'PodVolume',
  components: {
  },
  data() {
    return {
      volumes: this.template.spec.template.spec.volumes,
      containerRules: {
        name: [
          {required: true, message: ' ', trigger: ['blur', 'change']}
        ],
        image: [
          {required: true, message: ' ', trigger: ['blur', 'change']}
        ],
      },
      nsPvcs: [],
      nsSecrets: [],
      nsConfigmaps: [],
      rules: {
        'name': [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
      },
    }
  },
  props: ['template'],
  computed: {
  },
  methods: {
    newPodVolume,
  }
}
</script>

<style scoped lang="scss">
.input-class {
  width: 300px;
}

.border-span-header {
  padding-bottom: 8px;
}

.border-span-content {
  color: #F56C6C; 
  margin-right: 4px;
}
.box-card {
  font-size: 14px;
  color: #99a9bf
}
</style>

<style lang="scss">

</style>