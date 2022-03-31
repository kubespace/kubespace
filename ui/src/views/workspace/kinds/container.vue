<template>
  <div style="padding: 0px 20px 0px 10px;">
    <el-tabs type="border-card" v-model="containerTabVal" :before-leave="containerTabChange" 
      @tab-remove="containerTabRemove" @tab-click="containerTabClick">
      <el-tab-pane style="padding: 15px;" v-for="(c, i) in containers" :key="i + ''" :name="i + ''" :closable="containers.length > 1">
        <span slot="label">
          <i v-if="c.init" style="" class="el-icon-info"></i>
          {{ c.name ? c.name : '容器' + (i + 1) }}
        </span>
        <el-form :model="c" :rules="containerRules" label-width="120px"
          label-position="left" size="small">
          <el-form-item label="名称" prop="name">
            <el-input v-model="c.name" class="input-class" placeholder="容器名称"></el-input>
          </el-form-item>
          <el-form-item label="镜像" prop="image">
            <el-input v-model="c.image" class="input-class" placeholder="镜像名称，如：centos:7.6"></el-input>
          </el-form-item>
          <el-form-item label="镜像拉取" prop="imagePullPolicy">
            <el-select v-model="c.imagePullPolicy" placeholder="镜像拉取策略" class="input-class">
              <el-option label="Always" value="Always"></el-option>
              <el-option label="IfNotPresent" value="IfNotPresent"></el-option>
              <el-option label="Never" value="Never"></el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="initContainer" prop="init">
            <el-switch v-model="c.init"></el-switch>
          </el-form-item>
          <el-form-item label="Commands" prop="command">
            <el-input v-model="c.command" class="input-class" style="width: 400px;"
              placeholder='容器启动执行的命令，如：["/bin/bash"]'></el-input>
          </el-form-item>
          <el-form-item label="Args" prop="args">
            <el-input v-model="c.args" class="input-class" style="width: 400px;"
              placeholder='执行命令所需的参数，如：["-c", "echo hello"]'></el-input>
          </el-form-item>
          <el-form-item label="工作目录" prop="workingDir">
            <el-input v-model="c.workingDir" class="input-class" placeholder='容器启动后的工作目录，如：/app'></el-input>
          </el-form-item>
          <el-divider></el-divider>
          <div style="display: table;">
            <div style="display: table-cell; width: 120px;vertical-align: middle;">
              <label style="font-size: 14px; font-weight: 400; color: #99a9bf">资源</label>
            </div>
            <div style="display: table-cell;">
              <div style="margin-bottom: 10px;">
                <span style="color: #8B959C; font-size: .7em; margin-left: 50px; display: inline-block; width: 150px;">最小请求</span>
                <span style="color: #8B959C; font-size: .7em;">最大限制</span>
              </div>
              <div style="margin-bottom: 5px;">
                <span style="color: #8B959C; font-size: .7em; display: inline-block; width: 50px;">CPU</span>
                <el-input v-model="c.resources.requests.cpu" size="mini" class="input-class" 
                  placeholder="如：0.25" style="width: 120px;"></el-input>
                <span style="color: #8B959C; font-size: .85em; display: inline-block; width: 30px; text-align: center">～</span>
                <el-input v-model="c.resources.limits.cpu" size="mini" class="input-class" 
                  placeholder="如：0.5" style="width: 120px;"></el-input>
                <span style="color: #8B959C; font-size: .85em; margin-left: 5px;">Core</span>
              </div>
              <div>
                <span style="color: #8B959C; font-size: .7em; display: inline-block; width: 50px;">内存</span>
                <el-input v-model="c.resources.requests.memory" size="mini" class="input-class" 
                  placeholder="如：128" style="width: 120px;"></el-input>
                <span style="color: #8B959C; font-size: .85em; display: inline-block; width: 30px; text-align: center">～</span>
                <el-input v-model="c.resources.limits.memory" size="mini" class="input-class" 
                  placeholder="如：256" style="width: 120px;"></el-input>
                <span style="color: #8B959C; font-size: .85em; margin-left: 5px;">MiB</span>
              </div>
            </div>
          </div>
          <!-- </el-form-item> -->
          <el-divider></el-divider>

          <el-form-item label="端口映射" prop="port" >
            <el-row style="margin-bottom: -15px;" v-if="c.ports.length > 0">
              <el-col :span="5">
                <div class="border-span-header">
                  名称
                </div>
              </el-col>
              <el-col :span="5">
                <div class="border-span-header">
                  <span  class="border-span-content">*</span>容器端口
                </div>
              </el-col>
              <el-col :span="14">
                <div class="border-span-header">
                  协议
                </div>
              </el-col>
            </el-row>
            <el-row style="padding-top: 10px;" v-for="(item, idx) in c.ports" :key="idx">
              <el-col :span="5">
                <div class="border-span-header">
                  <el-input v-model="item.name" size="small" style="padding-right: 10px" placeholder="端口名称"></el-input>
                </div>
              </el-col>
              <el-col :span="5">
                <div class="border-span-header">
                  <el-input v-model="item.containerPort" size="small" style="padding-right: 10px" placeholder="容器暴露端口，如：80"></el-input>
                </div>
              </el-col>
              <el-col :span="4">
                <div class="border-span-header" style="padding-right: 10px">
                  <el-select v-model="item.protocol" placeholder="端口所属协议" >
                    <el-option label="TCP" value="TCP"></el-option>
                    <el-option label="UDP" value="UDP"></el-option>
                    <el-option label="SCTP" value="SCTP"></el-option>
                  </el-select>
                </div>
              </el-col>
              <el-col :span="5">
                <el-button circle size="mini" style="padding: 4px"
                  @click="c.ports.splice(idx, 1)" icon="el-icon-close"></el-button>
              </el-col>
            </el-row>
            <el-button plain size="mini" @click="c.ports.push({protocol: 'TCP'})" icon="el-icon-plus"></el-button>
          </el-form-item>

          <el-divider></el-divider>

          <el-form-item label="环境变量" prop="env">
            <el-row style="margin-bottom: -15px;" v-if="c.env.length > 0">
              <el-col :span="5">
                <div class="border-span-header">
                  <span  class="border-span-content">*</span>名称
                </div>
              </el-col>
              <el-col :span="5">
                <div class="border-span-header">
                  类型
                </div>
              </el-col>
              <el-col :span="10">
                <div class="border-span-header">
                  <span class="border-span-content">*</span>值
                </div>
              </el-col>
            </el-row>
            <el-row style="padding-top: 0px;" v-for="(item, idx) in c.env" :key="idx">
              <el-col :span="5">
                <div class="border-span-header">
                  <el-input v-model="item.name" size="small" style="padding-right: 10px;" placeholder="变量名称"></el-input>
                </div>
              </el-col>
              <el-col :span="5">
                <div class="border-span-header">
                  <el-select v-model="item.type" placeholder="类型" style="padding-right: 10px;" @change="delete item.value; delete item.key">
                    <el-option label="value" value="value"></el-option>
                    <el-option label="configMap" value="configMap"></el-option>
                    <el-option label="secret" value="secret"></el-option>
                    <el-option label="field" value="field"></el-option>
                    <el-option label="resource" value="resource"></el-option>
                  </el-select>
                </div>
              </el-col>
              <el-col :span="8" style="padding-right: 10px;">
                <div class="border-span-header">
                  <el-input v-if="item.type === 'value'" 
                    v-model="item.value" size="small" placeholder="变量值"></el-input>   
                  <el-input v-if="item.type === 'field'" 
                    v-model="item.value" size="small" placeholder="如：metadata.name, status.podIP"></el-input> 
                  <div v-if="item.type === 'configMap' || item.type === 'secret'">
                    <el-select v-model="item.value" value-key="name" :placeholder="item.type" style="width: 50%;">
                      <el-option :label="i" :value="i" :key="i" v-for="(c, i) in item.type === 'configMap' ? appConfigmaps : appSecrets"></el-option>
                    </el-select>
                    <!-- <span style="margin: 5px; font-wight: 800">.</span> -->
                    <el-select v-model="item.key" placeholder="Key" style="margin-left: 10px;width: 45%">
                      <template v-if="item.type == 'configMap'">
                        <el-option :label="d.key" :value="d.key" :key="d.key" v-for="d of resData(item.type, item.value)"></el-option>
                      </template>
                      <template v-else>
                        <el-option :label="k" :value="k" :key="k" v-for="(d, k) in resData(item.type, item.value)"></el-option>
                      </template>
                    </el-select>
                  </div>
                  <el-select v-if="item.type === 'resource'" v-model="item.value" placeholder="容器资源" style="width:100%;">
                    <el-option label="limits.cpu" value="limits.cpu"></el-option>
                    <el-option label="limits.memory" value="limits.memory"></el-option>
                    <el-option label="limits.ephemeral-storage" value="limits.ephemeral-storage"></el-option>
                    <el-option label="requests.cpu" value="requests.cpu"></el-option>
                    <el-option label="requests.memory" value="requests.memory"></el-option>
                    <el-option label="requests.ephemeral-storage" value="requests.ephemeral-storage"></el-option>
                  </el-select>
                </div>
              </el-col>
              <el-col :span="3">
                <el-button plain circle size="mini" style="padding: 4px"
                  @click="c.env.splice(idx, 1)" icon="el-icon-close"></el-button>
              </el-col>
            </el-row>
            <el-button plain size="mini" @click="c.env.push({type: 'value'})" icon="el-icon-plus"></el-button>
          </el-form-item>

          <el-divider></el-divider>

          <el-form-item label="存储挂载" prop="volumeMounts">
            <el-row style="margin-bottom: -15px;" v-if="c.volumeMounts.length > 0">
              <el-col :span="6">
                <div class="border-span-header">
                  <span  class="border-span-content">*</span>卷名称
                </div>
              </el-col>
              <el-col :span="6">
                <div class="border-span-header">
                  <span  class="border-span-content">*</span>容器挂载路径
                </div>
              </el-col>
              <el-col :span="12">
                <div class="border-span-header">
                  卷内部子路径
                </div>
              </el-col>
            </el-row>
            <el-row style="padding-top: 0px;" v-for="(item, idx) in c.volumeMounts" :key="idx">
              <el-col :span="6" style="padding-right: 10px;">
                <div class="border-span-header">
                  <el-select v-model="item.name" placeholder="请选择存储卷名称" style="width: 100%;">
                    <el-option :label="v.name" :value="v.name" v-for="v in volumes" :key="v.name"></el-option>
                  </el-select>
                </div>
              </el-col>
              <el-col :span="6" style="padding-right: 10px;">
                <div class="border-span-header">
                  <el-input v-model="item.mountPath" size="small" placeholder="挂载路径，如：/data"></el-input>
                </div>
              </el-col>
              <el-col :span="5" style="padding-right: 10px;">
                <div class="border-span-header">
                  <el-input v-model="item.subPath" size="small" placeholder="存储卷内部子路径"></el-input>
                </div>
              </el-col>
              <el-col :span="3">
                <el-button plain circle size="mini" style="padding: 4px"
                  @click="c.volumeMounts.splice(idx, 1)" icon="el-icon-close"></el-button>
              </el-col>
            </el-row>
            <el-button plain size="mini" @click="c.volumeMounts.push({})" icon="el-icon-plus"></el-button>
          </el-form-item>

          <el-divider></el-divider>

          <el-form-item label="健康检查" prop="livenessProbe">
            <el-switch v-model="c.livenessProbe.probe"></el-switch>
            <health-probe v-if="c.livenessProbe.probe" :probe="c.livenessProbe"></health-probe>
          </el-form-item>

          <el-form-item label="就绪检查" prop="readinessProbe">
            <el-switch v-model="c.readinessProbe.probe"></el-switch>
            <health-probe v-if="c.readinessProbe.probe" :probe="c.readinessProbe"></health-probe>
          </el-form-item>
          <el-divider></el-divider>
          
          <div class="border-header-class" 
            @click="scShow = !scShow">
            <div class="border-icon-class">
              <i :class="scShow ? 'el-icon-arrow-down' : 'el-icon-arrow-right'"></i>
            </div>
            <div class="border-content-class">
              <span>安全上下文</span> <br/>
              <span style="font-size: .8em; color: #8B959C;">对单个容器配置安全上下文</span>
            </div>
          </div>
          <el-collapse-transition>
            <div v-show="scShow">
              <div class="border-transition-class">
                <el-form :model="c" :rules="containerRules" label-width="120px"
                    label-position="left" size="small">
                  <el-form-item label="特权模式" prop="privileged">
                    <el-switch v-model="c.securityContext.privileged"></el-switch>
                  </el-form-item>
                  <el-form-item label="以用户ID运行" prop="runAsUser">
                    <el-input v-model="c.securityContext.runAsUser" size="small" class="input-class" placeholder="如：1001"></el-input>
                  </el-form-item>
                  <el-form-item label="以用户组ID运行" prop="runAsUser">
                    <el-input v-model="c.securityContext.runAsGroup" size="small" class="input-class" placeholder="如：1001"></el-input>
                  </el-form-item>
                  <el-form-item label="增加内核功能" prop="capabilities">
                    <div class="pnClass" style="width: 400px;">
                      <el-select v-model="c.securityContext.capabilities.add" multiple placeholder="" size="small">
                        <el-option
                          v-for="n in capabilities"
                          :key="n"
                          :label="n"
                          :value="n">
                        </el-option>
                      </el-select>
                    </div>
                  </el-form-item>
                  <el-form-item label="移除内核功能" prop="capabilities">
                    <div class="pnClass" style="width: 400px;">
                      <el-select v-model="c.securityContext.capabilities.drop" multiple placeholder="" size="small">
                        <el-option
                          v-for="n in capabilities"
                          :key="n"
                          :label="n"
                          :value="n">
                        </el-option>
                      </el-select>
                    </div>
                  </el-form-item>
                  <el-form-item label="以非root运行" prop="runAsNonRoot">
                    <el-switch v-model="c.securityContext.runAsNonRoot"></el-switch>
                  </el-form-item>
                  <el-form-item label="SELinux参数" prop="seLinuxOptions">
                    <el-row>
                      <el-col :span="4">
                        <div class="item-header" style="color: #8B959C">
                          User
                        </div>
                      </el-col>
                      <el-col :span="4">
                        <div class="item-header" style="color: #8B959C">
                          Role
                        </div>
                      </el-col>
                      <el-col :span="4">
                        <div class="item-header" style="color: #8B959C">
                          Type
                        </div>
                      </el-col>
                      <el-col :span="10">
                        <div class="item-header" style="color: #8B959C">
                          Level
                        </div>
                      </el-col>
                    </el-row>
                    <el-row>
                      <el-col :span="4" style="padding-right: 10px;">
                        <div class="item-content">
                          <el-input v-model="c.securityContext.seLinuxOptions.user" size="small" placeholder=""></el-input>
                        </div>
                      </el-col>
                      <el-col :span="4" style="padding-right: 10px;">
                        <div class="item-content">
                          <el-input v-model="c.securityContext.seLinuxOptions.role" size="small" placeholder=""></el-input>
                        </div>
                      </el-col>
                      <el-col :span="4" style="padding-right: 10px;">
                        <div class="item-content">
                          <el-input v-model="c.securityContext.seLinuxOptions.type" size="small" placeholder=""></el-input>
                        </div>
                      </el-col>
                      <el-col :span="4" style="padding-right: 10px;">
                        <div class="item-content">
                          <el-input v-model="c.securityContext.seLinuxOptions.level" size="small" placeholder=""></el-input>
                        </div>
                      </el-col>
                    </el-row>
                  </el-form-item>
                </el-form>
              </div>
            </div>
          </el-collapse-transition>
        </el-form>

      </el-tab-pane>
      <el-tab-pane label="+" name="plus">
        <span slot="label" style="font-size: 18;">+</span>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import { newContainer, HealthProbe } from '@/views/workspace/kinds'
import { transferSecret } from '@/api/secret'

export default {
  name: 'Container',
  components: {
    HealthProbe,
  },
  data() {
    return {
      containerTabVal: "0",
      scShow: false,
      containers: this.template.spec.template.spec.containers,
      volumes: this.template.spec.template.spec.volumes,
      containerRules: {
        name: [
          {required: true, message: ' ', trigger: ['blur', 'change']}
        ],
        image: [
          {required: true, message: ' ', trigger: ['blur', 'change']}
        ],
      },
      capabilities: [
        "ALL",
        "AUDIT_CONTROL",
        "AUDIT_WRITE",
        "BLOCK_SUSPEND",
        "CHOWN",
        "DAC_OVERRIDE",
        "DAC_READ_SEARCH",
        "FOWNER",
        "FSETID",
        "IPC_LOCK",
        "IPC_OWNER",
        "KILL",
        "LEASE",
        "LINUX_IMMUTABLE",
        "MAC_ADMIN",
        "MAC_OVERRIDE",
        "MKNOD",
        "NET_ADMIN",
        "NET_BIND_SERVICE",
        "NET_BROADCAST",
        "NET_RAW",
        "SETFCAP",
        "SETGID",
        "SETPCAP",
        "SETUID",
        "SYSLOGSYS_ADMIN",
        "SYS_BOOT",
        "SYS_CHROOT",
        "SYS_MODULE",
        "SYS_NICE",
        "SYS_PACCT",
        "SYS_PTRACE",
        "SYS_RAWIO",
        "SYS_RESOURCE",
        "SYS_TIME",
        "SYS_TTY_CONFIG",
        "WAKE_ALARM",
      ],
    }
  },
  props: ['template', 'appResources'],
  computed: {
    appConfigmaps() {
      let c = {}
      for(let r of this.appResources) {
        if(r.kind == 'ConfigMap' && r.metadata.name) {
          c[r.metadata.name] = r
        }
      }
      return c
    },
    appSecrets() {
      let c = {}
      for(let r of this.appResources) {
        if(r.kind == 'Secret' && r.metadata.name) {
          let secret = JSON.parse(JSON.stringify(r))
          let err = transferSecret(secret)
          if(!err) {
            c[r.metadata.name] = secret
          }
        }
      }
      return c
    }
  },
  methods: {
    resData(type, name) {
      if(type == 'configMap') {
        let data = []
        data = this.appConfigmaps[name] ? this.appConfigmaps[name].data : []
        return data
      } else {
        let data = {}
        data = this.appSecrets[name] ? this.appSecrets[name].data : {}
        return data
      }
    },
    addContainerTab() {
      var c = newContainer();
      this.containers.push(c)
      this.containerTabVal = (this.containers.length - 1) + ''
      return false
    },
    containerTabClick(tab) {
      if (tab.name === 'plus') this.addContainerTab();
    },
    containerTabChange(o, n) {
      if(o === 'plus') return false
      if(o >= this.containers.length + '') return false
    },
    containerTabRemove(removeIdx) {
      var i = parseInt(removeIdx)
      if(this.containers.length <= 1) return false
      var activeTab = this.containerTabVal
      if (i < parseInt(this.containerTabVal)) {
        activeTab = parseInt(this.containerTabVal) - 1
      } else if(i === parseInt(this.containerTabVal)) {
        activeTab = i - 1
        if (i <= 0) activeTab = 0;
      }
      
      this.containers.splice(i, 1)
      this.containerTabVal = activeTab + '';
    },
  }
}
</script>

<style scoped lang="scss">
.input-class {
  width: 300px;
}
.border-header-class {
  border: thin solid #EBEEF5; 
  display: table; 
  padding: 5px 10px; 
  width: 100%; cursor: pointer;
}

.border-icon-class {
  height: 100%; 
  vertical-align: middle; 
  display: table-cell; 
  width: 25px;
}

.border-content-class {
  display: tabel-cell; 
  height: 100%; 
  vertical-align: middle;
}

.border-transition-class {
  border: thin solid #EBEEF5; 
  border-top: 0px; 
  padding: 20px;
}

.border-span-header {
  color: #8B959C; 
  font-size: .85em; 
  padding-bottom: 8px;
}

.border-span-content {
  color: #F56C6C; 
  margin-right: 4px;
}
</style>

<style lang="scss">
.template-class {
  .el-tabs--border-card>.el-tabs__content {
    padding: 15px;
  }
}
</style>