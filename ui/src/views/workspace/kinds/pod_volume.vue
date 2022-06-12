<template>
  <div>
    <template v-for="(vol, idx) in volumes" >
      <el-card class="box-card" style="margin-bottom: 20px;" :key="idx" :body-style="{padding: '0px'}"
         v-if="!(vol.type == 'volumeClaimTemplates' && template.kind != 'StatefulSet')">
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
                  <el-option v-if="template.kind == 'StatefulSet'" label="VolumeClaimTemplates" value="volumeClaimTemplates"></el-option>
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
          <div v-if="vol.type == 'configMap'">
            <el-row>
              <el-col :span="12">
                <div class="border-span-header">
                  <span  class="border-span-content">*</span>{{ vol.type === 'configMap' ? '配置映射' : '密钥' }}
                </div>
                <el-select v-model="vol[vol.type].name" value-key="name" size="small" :placeholder="`请选择${vol.type}`" class="input-class">
                  
                  <el-option :key="k" :label="s.metadata.name" :value="s.metadata.name" v-for="(s, k) in nsConfigmaps"></el-option>
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
            <el-row style="padding-top: 0px;" v-for="(item, idx) in vol[vol.type].items" :key="idx">
              <el-col :span="7" style="padding-right: 10px;">
                <div class="border-span-header">
                  <el-select v-model="item.key" size="small" :placeholder="`请选择${vol.type}中的键`" style="width: 100%;" >
                    <el-option :key="s" :label="v.key" :value="v.key" v-for="(v, s) in nsConfigmaps[vol[vol.type].name] ? nsConfigmaps[vol[vol.type].name].data : []"></el-option>
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
                <el-button circle size="mini" style="padding: 5px; margin-top: 3px;" 
                  @click="vol[vol.type].items.splice(idx, 1)" icon="el-icon-close"></el-button>
              </el-col>
            </el-row>
            <el-button plain size="small" @click="vol[vol.type].items.push({})" style="margin-top: 10px;">
              添加指定文件项
            </el-button>
          </div>
          <div v-if="vol.type == 'secret'">
            <el-row>
              <el-col :span="12">
                <div class="border-span-header">
                  <span  class="border-span-content">*</span>{{ vol.type === 'configMap' ? '配置映射' : '密钥' }}
                </div>
                <el-select v-model="vol[vol.type].secretName" value-key="name" size="small" :placeholder="`请选择${vol.type}`" class="input-class">
                  
                  <el-option :key="k" :label="s.metadata.name" :value="s.metadata.name" v-for="(s, k) in nsSecrets"></el-option>
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
                    <el-option :key="s" :label="s" :value="s" v-for="(v, s) in nsSecrets[vol[vol.type].secretName] ? nsSecrets[vol[vol.type].secretName].data : {}"></el-option>
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
                <el-button circle size="mini" style="padding: 5px; margin-top: 3px;" 
                  @click="vol[vol.type].items.splice(idx, 1)" icon="el-icon-close"></el-button>
              </el-col>
            </el-row>
            <el-button plain size="small" @click="vol[vol.type].items.push({})" style="margin-top: 10px;">
              添加指定文件项
            </el-button>
          </div>
          <div v-if="vol.type === 'nfs'">
            <el-row>
              <el-col :span="8" style="padding-right: 10px;">
                <div class="border-span-header">
                  <span  class="border-span-content">*</span>服务器地址
                </div>
                <el-input v-model="vol['nfs'].server" size="small" class="input-class" style="width: 100%;" placeholder="NFS服务器地址"></el-input>
              </el-col>
              <el-col :span="8" style="padding-right: 10px;">
                <div class="border-span-header">
                  <span  class="border-span-content">*</span>访问路径
                </div>
                <el-input v-model="vol['nfs'].path" size="small" class="input-class" style="width: 100%;" placeholder="NFS服务访问路径"></el-input>
              </el-col>
              <el-col :span="8" style="padding-left: 10px;">
                <div class="border-span-header" style="margin-right: 25px;">
                  只读
                </div>
                <el-switch style="margin-top: 5px;" v-model="vol['nfs'].readOnly"></el-switch>
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
          <div v-if="vol.type === 'volumeClaimTemplates'">
            <el-row>
              <el-col :span="8" style="padding-right: 10px;">
                <div class="border-span-header">
                  <span  class="border-span-content">*</span>访问模式
                </div>
                <el-select v-model="vol['volumeClaimTemplates'].accessModes" placeholder="访问模式" multiple size="small" style="width: 100%;" >
                  <el-option label="ReadWriteOnce" value="ReadWriteOnce"></el-option>
                  <el-option label="ReadWriteMany" value="ReadWriteMany"></el-option>
                  <el-option label="ReadOnlyMany" value="ReadOnlyMany"></el-option>
                </el-select>
              </el-col>
              <el-col :span="8" style="padding-right: 10px;">
                <div class="border-span-header">
                  存储类
                </div>
                <el-input v-model="vol['volumeClaimTemplates'].storageClassName" size="small" class="input-class" 
                  style="width: 100%;" placeholder="存储类"></el-input>
              </el-col>
              <el-col :span="8">
                <div class="border-span-header">
                  <span  class="border-span-content">*</span>申请大小
                </div>
                <el-input v-model="vol['volumeClaimTemplates'].requests" size="small" class="input-class" 
                  placeholder="存储卷申请大小" style="width: 140px;">
                  <span slot="suffix" style="padding-right: 5px; margin-top: 5px;" >Gi </span>
                </el-input>
              </el-col>
            </el-row>
          </div>
        </div>
      </el-card>
    </template>
    <el-button plain size="small" @click="volumes.push(newPodVolume())">添加存储卷</el-button>
  </div>
</template>

<script>
import { newPodVolume, resolveConfigMap } from '@/views/workspace/kinds'
import { transferSecret } from '@/api/secret'

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
      // nsPvcs: [],
      // nsSecrets: [],
      // nsConfigmaps: [],
      rules: {
        'name': [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
      },
    }
  },
  props: ['template', 'appResources', 'projectResources'],
  computed: {
    nsPvcs() {
      let c = []
      for(let r of this.appResources) {
        if(r.kind == 'PersistentVolumeClaim' && r.metadata.name) {
          c.push({
            name: r.metadata.name
          })
        }
      }
      for(let r in this.projectResources) {
        if(r == 'PersistentVolumeClaim') {
          for(let p of this.projectResources[r]) {
            c.push({name: p.metadata.name})
          }
        }
      }
      return c
    },
    nsConfigmaps() {
      let c = {}
      for(let r of this.appResources) {
        if(r.kind == 'ConfigMap' && r.metadata.name) {
          c[r.metadata.name] = r
        }
      }
      for(let r in this.projectResources) {
        if(r == 'ConfigMap') {
          for(let cm of this.projectResources[r]) {
            let configmap = JSON.parse(JSON.stringify(cm))
            resolveConfigMap(configmap)
            c[cm.metadata.name] = configmap
          }
        }
      }
      return c
    },
    nsSecrets() {
      let c = {}
      for(let r of this.appResources) {
        if(r.kind == 'Secret' && r.metadata.name) {
          // c[r.metadata.name] = r
          let secret = JSON.parse(JSON.stringify(r))
          let err = transferSecret(secret)
          if(!err) {
            c[r.metadata.name] = secret
          }
        }
      }
      for(let r in this.projectResources) {
        if(r == 'Secret') {
          for(let s of this.projectResources[r]) {
            let secret = JSON.parse(JSON.stringify(s))
            c[s.metadata.name] = secret
          }
        }
      }
      return c
    }
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