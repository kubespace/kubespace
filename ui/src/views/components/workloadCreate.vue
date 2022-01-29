<template>
  <div>
    <clusterbar :titleName="titleName" />
    <div class="dashboard-container" v-if="!editYaml">
      <el-form :model="dsInfo" :rules="dsRules" ref="dsInfoForm" label-width="120px"
        label-position="left" size="small">
        <el-form-item label="命名空间" prop="namespace">
          <el-select v-model="dsInfo.namespace" placeholder="请选择命名空间" class="input-class">
            <el-option
              v-for="item in namespaces"
              :key="item.name"
              :label="item.name"
              :value="item.name">
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="dsInfo.name" class="input-class" placeholder=""></el-input>
        </el-form-item>
        <el-form-item label="副本数" prop="replicas" v-if="['Deployment', 'StatefulSet'].indexOf(this.createKind) >= 0">
          <el-input-number v-model="dsInfo.replicas" :min="1" class="input-class"></el-input-number>
        </el-form-item>
        <el-form-item label="标签" prop="lables">
          <div style="margin-bottom: 10px;" v-for="(l, i) in dsInfo.labels" :key="i">
            <el-input v-model="l.key" style="width: 150px;" placeholder="key"></el-input> = 
            <el-input v-model="l.value" class="input-class" placeholder="value"></el-input>
            <el-button plain size="mini" class="minus-btn-class" @click="dsInfo.labels.splice(i, 1)" 
              icon="el-icon-minus"></el-button>
          </div>
          <el-button plain size="mini" @click="dsInfo.labels.push({key: '', value: ''})" icon="el-icon-plus"></el-button>
        </el-form-item>
        <el-form-item label="注解" prop="annotations">
          <div style="margin-bottom: 10px;" v-for="(l, i) in dsInfo.annotations" :key="i">
            <el-input v-model="l.key" style="width: 150px;" placeholder="key"></el-input> = 
            <el-input v-model="l.value" class="input-class" placeholder="value"></el-input>
            <el-button plain size="mini" class="minus-btn-class" @click="dsInfo.annotations.splice(i, 1)" 
              icon="el-icon-minus"></el-button>
          </div>
          <el-button plain size="mini" @click="dsInfo.annotations.push({key: '', value: ''})" icon="el-icon-plus"></el-button>
        </el-form-item>
      </el-form>

      <div class="border-div-class">
        <div class="border-header-class" @click="show3 = !show3">
          <div class="border-icon-class">
            <i :class="show3 ? 'el-icon-arrow-down' : 'el-icon-arrow-right'"></i>
          </div>
          <div class="border-content-class">
            <span>存储</span> <br/>
            <span style="font-size: .8em; color: #8B959C;">挂载外部存储到容器中以持久化数据</span>
          </div>
        </div>
        <el-collapse-transition>
          <div v-show="show3">
            <div class="border-transition-class">
              <el-card class="box-card" style="margin-bottom: 20px;" v-for="(vol, idx) in dsInfo.volumes" 
                :key="idx" :body-style="{padding: '0px'}">
                <div slot="header" class="clearfix">
                  <el-row>
                    <el-col :span="12">
                      <div class="border-span-header">
                        <span class="border-span-content">*</span>卷名称
                      </div>
                      <el-input v-model="vol.name" size="small" class="input-class" placeholder="卷名称，如：vol1"></el-input>
                    </el-col>
                    <el-col :span="10">
                      <div class="border-span-header">
                        <span class="border-span-content">*</span>卷类型
                      </div>
                      <el-select v-model="vol.type" size="small" placeholder="卷类型" class="input-class">
                        <el-option label="persistentVolumeClaim" value="persistentVolumeClaim"></el-option>
                        <el-option label="hostPath" value="hostPath"></el-option>
                        <el-option label="emptyDir" value="emptyDir"></el-option>
                        <el-option label="configMap" value="configMap"></el-option>
                        <el-option label="secret" value="secret"></el-option>
                        <el-option label="nfs" value="nfs"></el-option>
                        <el-option label="glusterfs" value="glusterfs"></el-option>
                      </el-select>
                    </el-col>
                    <el-col :span="2">
                      <el-button style="float: right; padding: 3px 0" type="text"
                        @click="dsInfo.volumes.splice(idx, 1)">删除</el-button>
                    </el-col>
                  </el-row>
                </div>
                <div style="padding: 20px;" v-if="vol.type !== 'emptyDir'">
                  <el-row v-if="vol.type === 'persistentVolumeClaim'">
                    <el-col :span="12">
                      <div class="border-span-header">
                        <span class="border-span-content">*</span>存储声明
                      </div>
                      <el-select v-model="vol['persistentVolumeClaim'].name" size="small" placeholder="请选择存储声明" class="input-class">
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
                      <el-col :span="7">
                        <div class="border-span-header">
                          <el-select v-model="item.key" size="small" :placeholder="`请选择${vol.type}中的键`" class="input-class">
                            <el-option :key="s" :label="s" :value="s" v-for="(v, s) in vol[vol.type].obj ? vol[vol.type].obj.data : {}"></el-option>
                          </el-select>
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="border-span-header">
                          <el-input v-model="item.path" size="small" class="input-class" placeholder="文件映射相对路径"></el-input>
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="border-span-header">
                          <el-input v-model="item.mode" size="small" class="input-class" placeholder="文件模式"></el-input>
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
              <el-button plain size="small" @click="dsInfo.volumes.push(newVolume())">添加Volume</el-button>
            </div>
          </div>
        </el-collapse-transition>
      </div>

      <el-tabs value="containers" style="padding: 0px 0px;">
        <el-tab-pane label="容器组" name="containers">
          <div style="padding: 5px 10px 10px;">
            <el-tabs type="border-card" v-model="containerTabVal" :before-leave="containerTabChange" 
              @tab-remove="containerTabRemove" @tab-click="containerTabClick">
              <el-tab-pane label="用户管理" v-for="(c, i) in containers" :key="i + ''" :name="i + ''" :closable="containers.length > 1">
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
                  <el-form-item label="Commands" prop="commands">
                    <el-input v-model="c.commands" class="input-class" style="width: 400px;"
                      placeholder='容器启动执行的命令，如：["/bin/bash"]'></el-input>
                  </el-form-item>
                  <el-form-item label="Args" prop="args">
                    <el-input v-model="c.args" class="input-class" style="width: 400px;"
                      placeholder='执行命令所需的参数，如：["-c", "echo hello"]'></el-input>
                  </el-form-item>
                  <el-form-item label="工作目录" prop="workDir">
                    <el-input v-model="c.workDir" class="input-class" placeholder='容器启动后的工作目录，如：/app'></el-input>
                  </el-form-item>
                  <el-divider></el-divider>
                  <div style="display: table;">
                    <div style="display: table-cell; width: 120px;vertical-align: middle;">
                      <label style="font-size: 14px; color: #606266; font-weight: 700;">资源</label>
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

                  <el-form-item label="端口映射" prop="port">
                    <el-row style="margin-bottom: -15px;" v-if="c.ports.length > 0">
                      <el-col :span="7">
                        <div class="border-span-header">
                          名称
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="border-span-header">
                          <span  class="border-span-content">*</span>容器端口
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="border-span-header">
                          协议
                        </div>
                      </el-col>
                    </el-row>
                    <el-row style="padding-top: 10px;" v-for="(item, idx) in c.ports" :key="idx">
                      <el-col :span="7">
                        <div class="border-span-header">
                          <el-input v-model="item.name" size="small" class="input-class" placeholder="端口名称"></el-input>
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="border-span-header">
                          <el-input v-model="item.containerPort" size="small" class="input-class" placeholder="容器暴露端口，如：80"></el-input>
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="border-span-header">
                          <el-select v-model="item.protocal" placeholder="端口所属协议" class="input-class">
                            <el-option label="TCP" value="TCP"></el-option>
                            <el-option label="UDP" value="UDP"></el-option>
                            <el-option label="SCTP" value="SCTP"></el-option>
                          </el-select>
                        </div>
                      </el-col>
                      <el-col :span="3">
                        <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                          @click="c.ports.splice(idx, 1)" icon="el-icon-minus"></el-button>
                      </el-col>
                    </el-row>
                    <el-button plain size="mini" @click="c.ports.push({protocal: 'TCP'})" icon="el-icon-plus"></el-button>
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
                          <span  class="border-span-content">*</span>值
                        </div>
                      </el-col>
                    </el-row>
                    <el-row style="padding-top: 10px;" v-for="(item, idx) in c.env" :key="idx">
                      <el-col :span="5">
                        <div class="border-span-header">
                          <el-input v-model="item.name" size="small" class="input-class" style="width: 150px;" placeholder="名称"></el-input>
                        </div>
                      </el-col>
                      <el-col :span="5">
                        <div class="border-span-header">
                          <el-select v-model="item.type" placeholder="类型" class="input-class" style="width: 150px;" @change="delete item.value; delete item.key">
                            <el-option label="value" value="value"></el-option>
                            <el-option label="configMap" value="configMap"></el-option>
                            <el-option label="secret" value="secret"></el-option>
                            <el-option label="field" value="field"></el-option>
                            <el-option label="resource" value="resource"></el-option>
                          </el-select>
                        </div>
                      </el-col>
                      <el-col :span="8">
                        <div class="border-span-header">
                          <el-input v-if="item.type === 'value'" 
                            v-model="item.value" size="small" class="input-class" placeholder="值"></el-input>   
                          <el-input v-if="item.type === 'field'" 
                            v-model="item.value" size="small" class="input-class" placeholder="如：metadata.name, status.podIP"></el-input> 
                          <div v-if="item.type === 'configMap' || item.type === 'secret'">
                            <el-select v-model="item.value" value-key="name" :placeholder="item.type" class="input-class" style="width: 150px;">
                              <el-option :label="c.name" :value="c" :key="i" v-for="(c, i) in item.type === 'configMap' ? nsConfigmaps : nsSecrets"></el-option>
                            </el-select>
                            <span style="margin: 5px; font-wight: 800">.</span>
                            <el-select v-model="item.key" placeholder="Key" class="input-class" style="width: 150px">
                              <el-option :label="k" :value="k" :key="k" v-for="(v, k) in item.value ? item.value.data : {}"></el-option>
                            </el-select>
                          </div>
                          <el-select v-if="item.type === 'resource'" v-model="item.value" placeholder="容器资源" class="input-class">
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
                        <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                          @click="c.env.splice(idx, 1)" icon="el-icon-minus"></el-button>
                      </el-col>
                    </el-row>
                    <el-button plain size="mini" @click="c.env.push({type: 'value'})" icon="el-icon-plus"></el-button>
                  </el-form-item>

                  <el-divider></el-divider>

                  <el-form-item label="存储挂载" prop="volumeMounts">
                    <el-row style="margin-bottom: -15px;" v-if="c.volumeMounts.length > 0">
                      <el-col :span="7">
                        <div class="border-span-header">
                          <span  class="border-span-content">*</span>卷名称
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="border-span-header">
                          <span  class="border-span-content">*</span>容器挂载路径
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="border-span-header">
                          卷内部子路径
                        </div>
                      </el-col>
                    </el-row>
                    <el-row style="padding-top: 10px;" v-for="(item, idx) in c.volumeMounts" :key="idx">
                      <el-col :span="7">
                        <div class="border-span-header">
                          <el-select v-model="item.name" placeholder="请选择存储卷名称" class="input-class">
                            <el-option :label="v.name" :value="v.name" v-for="v in dsInfo.volumes" :key="v.name"></el-option>
                          </el-select>
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="border-span-header">
                          <el-input v-model="item.mountPath" size="small" class="input-class" placeholder="挂载路径，如：/data"></el-input>
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="border-span-header">
                          <el-input v-model="item.subPath" size="small" class="input-class" placeholder="存储卷内部子路径"></el-input>
                        </div>
                      </el-col>
                      <el-col :span="3">
                        <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                          @click="c.volumeMounts.splice(idx, 1)" icon="el-icon-minus"></el-button>
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
                              <el-select v-model="c.securityContext.addCapabilities" multiple placeholder="" size="small">
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
                              <el-select v-model="c.securityContext.delCapabilities" multiple placeholder="" size="small">
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
                                <div class="item-header">
                                  User
                                </div>
                              </el-col>
                              <el-col :span="4">
                                <div class="item-header">
                                  Role
                                </div>
                              </el-col>
                              <el-col :span="4">
                                <div class="item-header">
                                  Type
                                </div>
                              </el-col>
                              <el-col :span="10">
                                <div class="item-header">
                                  Level
                                </div>
                              </el-col>
                            </el-row>
                            <el-row>
                              <el-col :span="4">
                                <div class="item-content">
                                  <el-input v-model="c.securityContext.seLinuxOptions.user" size="small" placeholder=""></el-input>
                                </div>
                              </el-col>
                              <el-col :span="4">
                                <div class="item-content">
                                  <el-input v-model="c.securityContext.seLinuxOptions.role" size="small" placeholder=""></el-input>
                                </div>
                              </el-col>
                              <el-col :span="4">
                                <div class="item-content">
                                  <el-input v-model="c.securityContext.seLinuxOptions.type" size="small" placeholder=""></el-input>
                                </div>
                              </el-col>
                              <el-col :span="4">
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
        </el-tab-pane>

        <el-tab-pane label="网络" name="network">
          <div style="padding: 5px 10px 10px;">
            <div class="card-div">
              <el-form :model="dsInfo" :rules="dsRules" label-width="120px"
                label-position="left" size="small">
                <el-form-item label="DNS策略" prop="dnsPolicy">
                  <el-select v-model="dsInfo.dnsPolicy" size="small" class="input-class">
                    <el-option label="ClusterFirst" value="ClusterFirst"></el-option>
                    <el-option label="Default" value="Default"></el-option>
                    <el-option label="ClusterFirstWithHostNet" value="ClusterFirstWithHostNet"></el-option>
                    <el-option label="None" value="None"></el-option>
                  </el-select>
                </el-form-item>
                <el-form-item label="宿主机资源" prop="hostResource">
                  <el-checkbox v-model="dsInfo.hostNetwork">HostNetwork</el-checkbox>
                  <el-checkbox v-model="dsInfo.hostPID">HostPID</el-checkbox>
                  <el-checkbox v-model="dsInfo.hostIPC">HostIPC</el-checkbox>
                </el-form-item>
                <el-form-item label="主机名" prop="hostname">
                  <el-input v-model="dsInfo.hostname" size="small" class="input-class" placeholder=""></el-input>
                </el-form-item>
                <el-form-item label="子域名" prop="subdomain">
                  <el-input v-model="dsInfo.subdomain" size="small" class="input-class" placeholder=""></el-input>
                </el-form-item>
                <el-form-item label="主机别名" prop="hostAliases">
                  <el-row style="margin-bottom: -15px;" v-if="dsInfo.hostAliases.length > 0">
                    <el-col :span="7">
                      <div class="border-span-header">
                        <span  class="border-span-content">*</span>主机名
                      </div>
                    </el-col>
                    <el-col :span="17">
                      <div class="border-span-header">
                        <span  class="border-span-content">*</span>IP地址
                      </div>
                    </el-col>
                  </el-row>
                  <el-row style="padding-top: 10px;" v-for="(item, idx) in dsInfo.hostAliases" :key="idx">
                    <el-col :span="7">
                      <div class="border-span-header">
                        <el-input v-model="item.hostnames" size="small" class="input-class" placeholder="如：foo.com"></el-input>
                      </div>
                    </el-col>
                    <el-col :span="7">
                      <div class="border-span-header">
                        <el-input v-model="item.ip" size="small" class="input-class" placeholder="如：1.1.1.1"></el-input>
                      </div>
                    </el-col>
                    <el-col :span="3">
                      <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                        @click="dsInfo.hostAliases.splice(idx, 1)" icon="el-icon-minus"></el-button>
                    </el-col>
                  </el-row>
                  <el-button plain size="mini" @click="dsInfo.hostAliases.push({})" icon="el-icon-plus"></el-button>
                </el-form-item>
              </el-form>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="调度" name="schedule">
          <div style="padding: 5px 10px 10px;">
            <div class="card-div">
              <el-form :model="dsInfo" :rules="dsRules" label-width="120px"
                label-position="left" size="small">
                <el-form-item label="指定节点" prop="nodeName">
                  <el-select v-model="dsInfo.nodeName" size="small" class="input-class">
                    <el-option label="不指定" value=""></el-option>
                    <el-option :label="n.name" :value="n.name" :key="i" v-for="(n, i) in nodes"></el-option>
                  </el-select>
                </el-form-item>
                <el-form-item label="指定调度器" prop="schedulerName">
                  <el-input v-model="dsInfo.schedulerName" size="small" class="input-class" placeholder="指定调度器名称"></el-input>
                </el-form-item>
                <el-form-item label="选择节点标签" prop="nodeSelector">
                  <el-row style="margin-bottom: -15px;" v-if="dsInfo.nodeSelector.length > 0">
                    <el-col :span="7">
                      <div class="border-span-header">
                        <span  class="border-span-content">*</span>标签键
                      </div>
                    </el-col>
                    <el-col :span="17">
                      <div class="border-span-header">
                        <span  class="border-span-content">*</span>标签值
                      </div>
                    </el-col>
                  </el-row>
                  <el-row style="padding-top: 10px;" v-for="(item, idx) in dsInfo.nodeSelector" :key="idx">
                    <el-col :span="7">
                      <div class="border-span-header">
                        <el-input v-model="item.key" size="small" class="input-class" placeholder="节点标签键"></el-input>
                      </div>
                    </el-col>
                    <el-col :span="7">
                      <div class="border-span-header">
                        <el-input v-model="item.value" size="small" class="input-class" placeholder="节点标签值"></el-input>
                      </div>
                    </el-col>
                    <el-col :span="3">
                      <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                        @click="dsInfo.nodeSelector.splice(idx, 1)" icon="el-icon-minus"></el-button>
                    </el-col>
                  </el-row>
                  <el-button plain size="mini" @click="dsInfo.nodeSelector.push({})" icon="el-icon-plus"></el-button>
                </el-form-item>
                <el-form-item label="污点容忍" prop="toleration">
                  <el-row style="margin-bottom: -15px;" v-if="dsInfo.tolerations.length > 0">
                    <el-col :span="4">
                      <div class="border-span-header">
                        标签键
                      </div>
                    </el-col>
                    <el-col :span="3">
                      <div class="border-span-header">
                        运算符
                      </div>
                    </el-col>
                    <el-col :span="4">
                      <div class="border-span-header">
                        标签值
                      </div>
                    </el-col>
                    <el-col :span="4">
                      <div class="border-span-header">
                        影响
                      </div>
                    </el-col>
                    <el-col :span="8">
                      <div class="border-span-header">
                        时间
                      </div>
                    </el-col>
                  </el-row>
                  <el-row style="padding-top: 10px;" v-for="(item, idx) in dsInfo.tolerations" :key="idx">
                    <el-col :span="4">
                      <div class="item-content">
                        <el-input v-model="item.key" size="small" placeholder=""></el-input>
                      </div>
                    </el-col>
                    <el-col :span="3">
                      <div class="item-content">
                        <el-select v-model="item.operator" size="small">
                          <el-option label="等于" value="Equal"></el-option>
                          <el-option label="存在" value="Exists"></el-option>
                        </el-select>
                      </div>
                    </el-col>
                    <el-col :span="4">
                      <div class="item-content">
                        <el-input v-model="item.value" size="small" placeholder=""></el-input>
                      </div>
                    </el-col>
                    <el-col :span="4">
                      <div class="item-content">
                        <el-select v-model="item.effect" size="small" >
                          <el-option label="全部" value=""></el-option>
                          <el-option label="不调度" value="NoSchedule"></el-option>
                          <el-option label="倾向于不调度" value="PreferNoSchedule"></el-option>
                          <el-option label="不执行" value="NoExecute"></el-option>
                        </el-select>
                      </div>
                    </el-col>
                    <el-col :span="2">
                      <div class="item-content">
                        <el-input v-model="item.tolerationSeconds" size="small"
                         placeholder="" controls-position="right">
                          <template slot="suffix">
                            秒
                          </template> 
                        </el-input>
                      </div>
                    </el-col>
                    <el-col :span="3">
                      <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                        @click="dsInfo.tolerations.splice(idx, 1)" icon="el-icon-minus"></el-button>
                    </el-col>
                  </el-row>
                  <el-button plain size="mini" @click="dsInfo.tolerations.push({})" icon="el-icon-plus"></el-button>
                </el-form-item>
                <el-form-item label="节点亲和性" prop="nodeAffinity">
                  <el-card class="box-card" style="margin-bottom: 20px;" v-for="(aff, idx) in dsInfo.affinity.nodeAffinity" 
                    :key="idx" :body-style="{padding: '20px'}">
                    <el-row style="margin-bottom: 10px;">
                      <el-col :span="aff.type === 'preferred' ? 9 : 22">
                        <div class="item-header">
                          优先级
                        </div>
                        <el-select v-model="aff.type" size="small" class="input-class">
                          <el-option label="必须" value="required"></el-option>
                          <el-option label="最好" value="preferred"></el-option>
                        </el-select>
                      </el-col>
                      <el-col :span="13" v-if="aff.type === 'preferred'">
                        <div class="item-header">
                          <span  class="border-span-content">*</span>权重
                        </div>
                        <el-input-number controls-position="right" v-model="aff.weigth" size="small" placeholder="" :max="100" :min="1"
                        ></el-input-number>
                      </el-col>
                      <el-col :span="2">
                        <el-button style="float: right; padding: 3px 0" type="text"
                          @click="dsInfo.affinity.nodeAffinity.splice(idx, 1)">删除</el-button>
                      </el-col>
                    </el-row>
                    <el-row style="margin-bottom: -15px;" v-if="aff.nodeSelectorTerms.length > 0">
                      <el-col :span="3">
                        <div class="border-span-header">
                          属性
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="border-span-header">
                          键
                        </div>
                      </el-col>
                      <el-col :span="3">
                        <div class="border-span-header">
                          运算符
                        </div>
                      </el-col>
                      <el-col :span="10">
                        <div class="border-span-header">
                          值
                        </div>
                      </el-col>
                    </el-row>
                    <el-row style="padding-top: 5;" v-for="(item, t_idx) in aff.nodeSelectorTerms" :key="t_idx">
                      <el-col :span="3">
                        <div class="item-content">
                          <el-select v-model="item.type" size="small" >
                            <el-option label="标签" value="label"></el-option>
                            <el-option label="字段" value="field"></el-option>
                          </el-select>
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="item-content">
                          <el-input v-model="item.key" size="small" placeholder=""></el-input>
                        </div>
                      </el-col>
                      <el-col :span="3">
                        <div class="item-content">
                          <el-select v-model="item.operator" size="small">
                            <el-option label="In" value="In"></el-option>
                            <el-option label="NotIn" value="NotIn"></el-option>
                            <el-option label="Exists" value="Exists"></el-option>
                            <el-option label="DoesNotExist" value="DoesNotExist"></el-option>
                            <el-option label="Gt(>)" value="Gt"></el-option>
                            <el-option label="Lt(<)" value="Lt"></el-option>
                          </el-select>
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="item-content">
                          <el-input v-model="item.values" size="small" placeholder="" :disabled="['Exists', 'DoesNotExist'].indexOf(item.operator) >= 0"></el-input>
                        </div>
                      </el-col>
                      <el-col :span="3">
                        <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                          @click="aff.nodeSelectorTerms.splice(t_idx, 1)" icon="el-icon-minus"></el-button>
                      </el-col>
                    </el-row>
                    <el-button style="padding: 3px 0" type="text"
                      @click="aff.nodeSelectorTerms.push({type: 'label', operator: 'In'})">添加规则</el-button>
                  </el-card>
                  <el-button plain size="mini" @click="dsInfo.affinity.nodeAffinity.push({type: 'required', weight: 1, nodeSelectorTerms: []})" icon="el-icon-plus"></el-button>
                </el-form-item>
                <el-form-item label="Pod亲和性" prop="podAffinity">
                  <el-card class="box-card" style="margin-bottom: 20px;" v-for="(aff, idx) in dsInfo.affinity.podAffinity" 
                    :key="idx" :body-style="{padding: '20px'}">
                    <el-row style="margin-bottom: 10px;">
                      <el-col :span="9">
                        <div class="item-header">
                          优先级
                        </div>
                        <div style="padding-right: 18px;" class="pnClass">
                          <el-select v-model="aff.type" size="small">
                            <el-option label="必须" value="required"></el-option>
                            <el-option label="最好" value="preferred"></el-option>
                          </el-select>
                        </div>
                      </el-col>
                      <el-col :span="13" v-if="aff.type === 'preferred'">
                        <div class="item-header">
                          <span  class="border-span-content">*</span>权重
                        </div>
                        <el-input-number controls-position="right" v-model="aff.weigth" size="small" placeholder="" max="100" min="1"
                        ></el-input-number>
                      </el-col>
                      <el-col :span="aff.type === 'preferred' ? 2 : 15">
                        <el-button style="float: right; padding: 3px 0" type="text"
                          @click="dsInfo.affinity.podAffinity.splice(idx, 1)">删除</el-button>
                      </el-col>
                    </el-row>
                    <el-row style="margin-bottom: 10px;">
                      <el-col :span="9">
                        <div class="item-header">
                          <span  class="border-span-content">*</span>拓扑键
                        </div>
                        <div style="padding-right: 18px;">
                          <el-input v-model="aff.podAffinityTerm.topologyKey" size="small" placeholder=""></el-input>
                        </div>
                      </el-col>
                      <el-col :span="9">
                        <div class="item-header">
                          命名空间
                        </div>
                        <div class="pnClass" style="padding-right: 18px;">
                          <el-select v-model="aff.podAffinityTerm.namespaces" multiple placeholder="默认为空，表示当前命名空间" size="small">
                            <el-option
                              v-for="n in namespaces"
                              :key="n.name"
                              :label="n.name"
                              :value="n.name">
                            </el-option>
                          </el-select>
                        </div>
                      </el-col>
                    </el-row>
                    <el-row style="margin-bottom: -15px;" v-if="aff.podAffinityTerm.labelSelector.length > 0">
                      <el-col :span="7">
                        <div class="border-span-header">
                          标签键
                        </div>
                      </el-col>
                      <el-col :span="3">
                        <div class="border-span-header">
                          运算符
                        </div>
                      </el-col>
                      <el-col :span="14">
                        <div class="border-span-header">
                          标签值
                        </div>
                      </el-col>
                    </el-row>
                    <el-row style="padding-top: 5;" v-for="(item, t_idx) in aff.podAffinityTerm.labelSelector" :key="t_idx">
                      <el-col :span="7">
                        <div class="item-content">
                          <el-input v-model="item.key" size="small" placeholder=""></el-input>
                        </div>
                      </el-col>
                      <el-col :span="3">
                        <div class="item-content">
                          <el-select v-model="item.operator" size="small">
                            <el-option label="Equal" value="Equal"></el-option>
                            <el-option label="In" value="In"></el-option>
                            <el-option label="NotIn" value="NotIn"></el-option>
                            <el-option label="Exists" value="Exists"></el-option>
                            <el-option label="DoesNotExist" value="DoesNotExist"></el-option>
                          </el-select>
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="item-content">
                          <el-input v-model="item.values" size="small" placeholder=""></el-input>
                        </div>
                      </el-col>
                      <el-col :span="3">
                        <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                          @click="aff.podAffinityTerm.labelSelector.splice(t_idx, 1)" icon="el-icon-minus"></el-button>
                      </el-col>
                    </el-row>
                    <el-button style="padding: 3px 0" type="text"
                      @click="aff.podAffinityTerm.labelSelector.push({operator: 'Equal'})">添加规则</el-button>
                  </el-card>
                  <el-button plain size="mini" @click="dsInfo.affinity.podAffinity.push({type: 'required', weight: 1, podAffinityTerm: {labelSelector: [], namespaces: []}})" icon="el-icon-plus"></el-button>
                </el-form-item>
                <el-form-item label="Pod反亲和性" prop="podAntiAffinity">
                  <el-card class="box-card" style="margin-bottom: 20px;" v-for="(aff, idx) in dsInfo.affinity.podAntiAffinity" 
                    :key="idx" :body-style="{padding: '20px'}">
                    <el-row style="margin-bottom: 10px;">
                      <el-col :span="9">
                        <div class="item-header">
                          优先级
                        </div>
                        <div style="padding-right: 18px;" class="pnClass">
                          <el-select v-model="aff.type" size="small">
                            <el-option label="必须" value="required"></el-option>
                            <el-option label="最好" value="preferred"></el-option>
                          </el-select>
                        </div>
                      </el-col>
                      <el-col :span="13" v-if="aff.type === 'preferred'">
                        <div class="item-header">
                          <span  class="border-span-content">*</span>权重
                        </div>
                        <el-input-number controls-position="right" v-model="aff.weigth" size="small" placeholder="" max="100" min="1"
                        ></el-input-number>
                      </el-col>
                      <el-col :span="aff.type === 'preferred' ? 2 : 15">
                        <el-button style="float: right; padding: 3px 0" type="text"
                          @click="dsInfo.affinity.podAntiAffinity.splice(idx, 1)">删除</el-button>
                      </el-col>
                    </el-row>
                    <el-row style="margin-bottom: 10px;">
                      <el-col :span="9">
                        <div class="item-header">
                          <span  class="border-span-content">*</span>拓扑键
                        </div>
                        <div style="padding-right: 18px;">
                          <el-input v-model="aff.podAffinityTerm.topologyKey" size="small" placeholder=""></el-input>
                        </div>
                      </el-col>
                      <el-col :span="9">
                        <div class="item-header">
                          命名空间
                        </div>
                        <div class="pnClass" style="padding-right: 18px;">
                          <el-select v-model="aff.podAffinityTerm.namespaces" multiple placeholder="默认为空，表示当前命名空间" size="small">
                            <el-option
                              v-for="n in namespaces"
                              :key="n.name"
                              :label="n.name"
                              :value="n.name">
                            </el-option>
                          </el-select>
                        </div>
                      </el-col>
                    </el-row>
                    <el-row style="margin-bottom: -15px;" v-if="aff.podAffinityTerm.labelSelector.length > 0">
                      <el-col :span="7">
                        <div class="border-span-header">
                          标签键
                        </div>
                      </el-col>
                      <el-col :span="3">
                        <div class="border-span-header">
                          运算符
                        </div>
                      </el-col>
                      <el-col :span="14">
                        <div class="border-span-header">
                          标签值
                        </div>
                      </el-col>
                    </el-row>
                    <el-row style="padding-top: 5;" v-for="(item, t_idx) in aff.podAffinityTerm.labelSelector" :key="t_idx">
                      <el-col :span="7">
                        <div class="item-content">
                          <el-input v-model="item.key" size="small" placeholder=""></el-input>
                        </div>
                      </el-col>
                      <el-col :span="3">
                        <div class="item-content">
                          <el-select v-model="item.operator" size="small">
                            <el-option label="Equal" value="Equal"></el-option>
                            <el-option label="In" value="In"></el-option>
                            <el-option label="NotIn" value="NotIn"></el-option>
                            <el-option label="Exists" value="Exists"></el-option>
                            <el-option label="DoesNotExist" value="DoesNotExist"></el-option>
                          </el-select>
                        </div>
                      </el-col>
                      <el-col :span="7">
                        <div class="item-content">
                          <el-input v-model="item.values" size="small" placeholder="" :disabled="['Exists', 'DoseNotExist'].indexOf(item.operator)"></el-input>
                        </div>
                      </el-col>
                      <el-col :span="3">
                        <el-button plain size="mini" style="padding-left: 10px; padding-right: 10px;" 
                          @click="aff.podAffinityTerm.labelSelector.splice(t_idx, 1)" icon="el-icon-minus"></el-button>
                      </el-col>
                    </el-row>
                    <el-button style="padding: 3px 0" type="text"
                      @click="aff.podAffinityTerm.labelSelector.push({operator: 'Equal'})">添加规则</el-button>
                  </el-card>
                  <el-button plain size="mini" @click="dsInfo.affinity.podAntiAffinity.push({type: 'required', weight: 1, podAffinityTerm: {labelSelector: [], namespaces: []}})" icon="el-icon-plus"></el-button>
                </el-form-item>
              </el-form>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="安全" name="security">
          <div style="padding: 5px 10px 10px;">
            <div class="card-div">
              <el-form :model="dsInfo" :rules="dsRules" label-width="120px"
                  label-position="left" size="small">
                <el-form-item label="以用户ID运行" prop="runAsUser">
                  <el-input v-model="dsInfo.securityContext.runAsUser" size="small" class="input-class" placeholder="如：1001"></el-input>
                </el-form-item>
                <el-form-item label="以用户组ID运行" prop="runAsUser">
                  <el-input v-model="dsInfo.securityContext.runAsGroup" size="small" class="input-class" placeholder="如：1001"></el-input>
                </el-form-item>
                <el-form-item label="以非root运行" prop="runAsNonRoot">
                  <el-switch v-model="dsInfo.securityContext.runAsNonRoot"></el-switch>
                </el-form-item>
                <el-form-item label="sysctls" prop="sysctls">
                  <div style="margin-bottom: 10px;" v-for="(l, i) in dsInfo.securityContext.sysctls" :key="i">
                    <el-input v-model="l.name" style="width: 150px;" placeholder="name"></el-input> = 
                    <el-input v-model="l.value" class="input-class" placeholder="value"></el-input>
                    <!-- <el-link :underline="false" style="margin-left: 10px;" @click="dsInfo.labels.splice(i, 1)">删除</el-link> -->
                    <el-button plain size="mini" style="margin-left: 10px; padding-left: 10px; padding-right: 10px;" 
                      @click="dsInfo.securityContext.sysctls.splice(i, 1)" icon="el-icon-minus"></el-button>
                  </div>
                  <el-button plain size="mini" @click="dsInfo.securityContext.sysctls.push({})" icon="el-icon-plus"></el-button>
                </el-form-item>
                <el-form-item label="SELinux参数" prop="seLinuxOptions">
                  <el-row>
                    <el-col :span="4">
                      <div class="item-header">
                        User
                      </div>
                    </el-col>
                    <el-col :span="4">
                      <div class="item-header">
                        Role
                      </div>
                    </el-col>
                    <el-col :span="4">
                      <div class="item-header">
                        Type
                      </div>
                    </el-col>
                    <el-col :span="10">
                      <div class="item-header">
                        Level
                      </div>
                    </el-col>
                  </el-row>
                  <el-row>
                    <el-col :span="4">
                      <div class="item-content">
                        <el-input v-model="dsInfo.securityContext.seLinuxOptions.user" size="small" placeholder=""></el-input>
                      </div>
                    </el-col>
                    <el-col :span="4">
                      <div class="item-content">
                        <el-input v-model="dsInfo.securityContext.seLinuxOptions.role" size="small" placeholder=""></el-input>
                      </div>
                    </el-col>
                    <el-col :span="4">
                      <div class="item-content">
                        <el-input v-model="dsInfo.securityContext.seLinuxOptions.type" size="small" placeholder=""></el-input>
                      </div>
                    </el-col>
                    <el-col :span="4">
                      <div class="item-content">
                        <el-input v-model="dsInfo.securityContext.seLinuxOptions.level" size="small" placeholder=""></el-input>
                      </div>
                    </el-col>
                  </el-row>
                </el-form-item>
              </el-form>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
    <div class="dashboard-container" v-else>
      <yaml v-model="yamlValue"></yaml>
    </div>
    <div class="dashboard-container">
      <el-row style="margin-bottom: 20px;">
        <el-col :span="24">
          <el-button type="primary" style="float: right; margin-left: 12px;padding-left: 45px; padding-right: 45px;" 
            @click="create" :loading="createBtnLoad">创建</el-button>
          <el-button plain type="primary" style="float: right; margin-right: 12px; padding-left: 30px; padding-right: 30px;"
            @click="toEditYaml">{{ editYaml ? '编辑表单' : '编辑YAML' }}</el-button>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script>
import { Clusterbar, HealthProbe } from '@/views/components'
import { listDeployments } from '@/api/deployment'
import { listNamespace } from '@/api/namespace'
import { listConfigMaps } from '@/api/config_map'
import { listSecrets } from '@/api/secret'
import { listNodes } from '@/api/nodes'
import { createYaml } from '@/api/cluster'
import { listPersistentVolumeClaim } from '@/api/persistent_volume_claim'
import { Message } from 'element-ui'
import yaml from 'js-yaml'
import { Yaml } from '@/views/components'

export default {
  name: 'Deployment',
  components: {
    Clusterbar,
    HealthProbe,
    Yaml,
  },
  props: {
    workType: {
      type: String,
      required: true,
      default: 'Deployments'
    },
  },
  data() {
    var containers = [this.newContainerObj()]
    var checkDsName = (rule, value, callback) => {
      if (!value) {
        return callback(new Error('名称不能为空'));
      } else {
        callback();
      }
    };
    return {
      createBtnLoad: false,
      show3: false,
      scShow: false,
      editYaml: false,
      yamlValue: '',
      titleName: [this.workType, "创建"],
      namespaces: [],
      configmaps: [],
      pvcs: [],
      secrets: [],
      nodes: [],
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
      dsInfo: {
        name: '',
        namespace: 'default',
        replicas: 1,
        labels: [],
        annotations: [],
        volumes: [],
        dnsPolicy: 'ClusterFirst',
        hostNetwork: false,
        hostPID: false,
        hostIPC: false,
        hostname: '',
        subdomain: '',
        hostAliases: [],
        nodeName: '',
        schedulerName: '',
        affinity: {
          nodeAffinity: [],
          podAffinity: [],
          podAntiAffinity: []
        },
        tolerations: [],
        nodeSelector: [],
        securityContext: {seLinuxOptions: {}, sysctls: []},
      },
      containers: containers,
      containerTabVal: "0",
      dsRules: {
        name: [
          { required: true, validator: checkDsName, message: '', trigger: 'blur'}
        ],
        namespace: [
          {required: true, message: '', trigger: 'change'}
        ],
        replicas: [
          {required: true, message: '', trigger: 'change'}
        ]
      },
      containerRules: {
        name: [
          {required: true, message: '', trigger: 'blur'}
        ],
        image: [
          {required: true, message: '', trigger: 'blur'}
        ],
      }
    }
  },
  created() {
    // this.fetchData()
    this.fetchNamespace()
    this.fetchConfigmap()
    this.fetchSecret()
    this.fetchPvc()
    this.fetchNode()
  },
  computed: {
    cluster: function() {
      return this.$store.state.cluster
    },
    nsSecrets: function() {
      var s = this.secrets.filter(({namespace})=> namespace === this.dsInfo.namespace)
      s.sort((a,b) => {return a.name > b.name ? 1 : -1})
      return s
    },
    nsConfigmaps: function() {
      var s = this.configmaps.filter(({namespace})=> namespace === this.dsInfo.namespace)
      s.sort((a,b) => {return a.name > b.name ? 1 : -1})
      return s
    },
    nsPvcs: function() {
      var s = this.pvcs.filter(({namespace})=> namespace === this.dsInfo.namespace)
      s.sort((a,b) => {return a.name > b.name ? 1 : -1})
      return s
    },
    createKind: function() {
      if(this.workType === 'Deployments') return 'Deployment'
      if(this.workType === 'StatefulSets') return 'StatefulSet'
      if(this.workType === 'DaemonSets') return 'DaemonSet'
      if(this.workType === 'Jobs') return 'Job'
    },
    createApiVersion: function() {
      if(this.workType === 'Deployments') return 'apps/v1'
      if(this.workType === 'StatefulSets') return 'apps/v1'
      if(this.workType === 'DaemonSets') return 'apps/v1'
      if(this.workType === 'Jobs') return 'batch/v1'
    },
    workloadDetail: function() {
      if(this.workType === 'Deployments') return 'deploymentDetail'
      if(this.workType === 'StatefulSets') return 'statefulsetDetail'
      if(this.workType === 'DaemonSets') return 'daemonsetDetail'
      if(this.workType === 'Jobs') return 'jobDetail'
    },
  },
  methods: {
    fetchNamespace: function() {
      this.namespaces = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listNamespace(cluster).then(response => {
          this.namespaces = response.data
          this.namespaces.sort((a, b) => {return a.name > b.name ? 1 : -1})
        }).catch((err) => {
          console.log(err)
        })
      } else {
        Message.error("获取集群异常，请刷新重试")
      }
    },
    fetchConfigmap: function() {
      this.namespaces = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listConfigMaps(cluster).then(response => {
          this.configmaps = response.data
          this.configmaps.sort((a, b) => {return a.name > b.name ? 1 : -1})
        }).catch((err) => {
          console.log(err)
        })
      } else {
        Message.error("获取集群异常，请刷新重试")
      }
    },
    fetchSecret: function() {
      this.namespaces = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listSecrets(cluster).then(response => {
          this.secrets = response.data
          this.secrets.sort((a, b) => {return a.name > b.name ? 1 : -1})
        }).catch((err) => {
          console.log(err)
        })
      } else {
        Message.error("获取集群异常，请刷新重试")
      }
    },
    fetchNode: function() {
      this.nodes = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listNodes(cluster).then(response => {
          this.nodes = response.data
          this.nodes.sort((a, b) => {return a.name > b.name ? 1 : -1})
        }).catch((err) => {
          console.log(err)
        })
      } else {
        Message.error("获取集群异常，请刷新重试")
      }
    },
    fetchPvc: function() {
      this.namespaces = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listPersistentVolumeClaim(cluster).then(response => {
          this.pvcs = response.data || []
          this.pvcs.sort((a, b) => {return a.name > b.name ? 1 : -1})
        }).catch((err) => {
          console.log(err)
        })
      } else {
        Message.error("获取集群异常，请刷新重试")
      }
    },
    newVolume() {
      return {
        name: '',
        type: 'persistentVolumeClaim',
        persistentVolumeClaim: {},
        glusterfs: {},
        nfs: {},
        secret: {items: [], obj: {keys: []}},
        configMap: {items: [], obj: {keys: []}},
        emptyDir: {},
        hostPath: {}
      }
    },
    newContainerObj() {
      return {
        init: false,
        name: '',
        image: '',
        commands: '',
        args: '',
        workDir: '',
        ports: [],
        env: [],
        resources: {limits: {}, requests: {}},
        livenessProbe: {probe: false, type: 'http', handle: {}, successThreshold: 1, failureThreshold: 3,
                        initialDelaySeconds: 0, timeoutSeconds: 1, periodSeconds: 10},
        readinessProbe: {probe: false, type: 'http', handle: {}, successThreshold: 1, failureThreshold: 3,
                        initialDelaySeconds: 0, timeoutSeconds: 1, periodSeconds: 10},
        imagePullPolicy: '',
        volumeMounts: [],
        stdin: false,
        tty: false,
        securityContext: {seLinuxOptions: {}},
      }
    },
    addContainerTab() {
      var c = this.newContainerObj();
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
    create() {
      var yamlStr = ''
      if(this.editYaml) {
        yamlStr = this.yamlValue
      } else {
        var {obj, success} = this.getCreateObj()
        if (!success) return
        yamlStr = yaml.dump(obj, {indent: 0})
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!yamlStr) {
        Message.error("创建YAML不能为空")
        return
      }
      console.log(cluster, yamlStr)
      this.createBtnLoad = true
      createYaml(cluster, yamlStr).then((resp) => {
        console.log(resp.msg)
        Message.success(resp.msg)
        if(this.workType === 'Deployments') {
          this.$router.push({name: this.workloadDetail, params: {namespace: this.dsInfo.namespace, deploymentName: this.dsInfo.name}})
        }
        if(this.workType === 'StatefulSets') {
          this.$router.push({name: this.workloadDetail, params: {namespace: this.dsInfo.namespace, statefulsetName: this.dsInfo.name}})
        }
        if(this.workType === 'DaemonSets') {
          this.$router.push({name: this.workloadDetail, params: {namespace: this.dsInfo.namespace, daemonsetName: this.dsInfo.name}})
        }
        if(this.workType === 'Jobs') {
          this.$router.push({name: this.workloadDetail, params: {namespace: this.dsInfo.namespace, jobName: this.dsInfo.name}})
        }
        return
      }).catch(() => {
        // console.log(e) 
        this.createBtnLoad = false
      })
    },
    toEditYaml() {
      if(this.editYaml) {
        this.$confirm('返回表单编辑会丢失修改内容, 是否继续?', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          this.editYaml = false
        }).catch(() => {
        });
        return
      }
      var {obj, success} = this.getCreateObj()
      // console.log(obj, success)
      if (!success) return
      var yamlStr = yaml.dump(obj, {indent: 0})
      this.yamlValue = yamlStr
      this.editYaml = true
    },
    getCreateObj() {
      console.log(this.dsInfo)
      console.log(this.containers)
      var dsInfo = this.dsInfo
      var obj = {
        kind: this.createKind,
        apiVersion: this.createApiVersion,
        metadata: {
          name: dsInfo.name,
          namespace: dsInfo.namespace,
        },
        spec: {
          replicas: dsInfo.replicas,
          selector: {
            matchLabels: {
              "osp-app": dsInfo.name
            }
          },
          template: {
            metadata: {
              labels: {
                "osp-app": dsInfo.name
              }
            },
            spec: {}
          }
        }
      }
      if (dsInfo.labels && dsInfo.labels.length > 0) {
        var labels = {}
        for (var l of dsInfo.labels) {
          labels[l.key] = l.value
        }
        obj.metadata.labels = labels
      }
      if (dsInfo.annotations && dsInfo.annotations.length > 0) {
        var ann = {}
        for (let a of dsInfo.annotations) {
          ann[a.key] = a.value
        }
        obj.metadata.annotations = ann
      }
      var {spec, success} = this.getSpec()
      // console.log(spec, success)
      if (!success) return {obj: null, success: false}
      obj.spec.template.spec = spec
      console.log(obj)
      return {obj, success: true}
    },
    getSpec() {
      var spec = {}
      if (this.dsInfo.dnsPolicy) {
        spec.dnsPolicy = this.dsInfo.dnsPolicy
      }
      if (this.dsInfo.nodeSelector && this.dsInfo.nodeSelector.length > 0) {
        var ns = {}
        for(let n of this.dsInfo.nodeSelector) {
          ns[n.key] = n.value
        }
        spec.nodeSelector = ns
      }
      if(this.dsInfo.nodeName) spec.nodeName = this.dsInfo.nodeName
      if(this.dsInfo.hostNetwork) spec.hostNetwork = this.dsInfo.hostNetwork
      if(this.dsInfo.hostPID) spec.hostPID = this.dsInfo.hostPID
      if(this.dsInfo.hostIPC) spec.hostIPC = this.dsInfo.hostIPC
      if(this.dsInfo.hostname) spec.hostname = this.dsInfo.hostname
      if(this.dsInfo.subdomain) spec.subdomain = this.dsInfo.subdomain
      if(this.dsInfo.schedulerName) spec.schedulerName = this.dsInfo.schedulerName
      if(this.dsInfo.hostAliases && this.dsInfo.hostAliases.length > 0) {
        var ha = []
        for(let h of this.dsInfo.hostAliases) {
          if(!h.ip || h.ip === '') {
            Message.error("主机别名IP地址不能为空")
            return {spec:null, success:false}
          }
          if(!h.hostnames || h.hostnames === '') {
            Message.error("主机别名主机名不能为空")
            return {spec:null, success:false}
          }
          ha.push({
            ip: h.ip,
            hostnames: h.hostnames.split(',')
          })
        }
        spec.hostAliases = ha
      }
      if(this.dsInfo.tolerations && this.dsInfo.tolerations.length > 0) {
        spec.tolerations = this.dsInfo.tolerations
      }
      var securityContext = this.dsInfo.securityContext
      var specSc = {}
      var seLinuxOptions = {}
      if(securityContext.seLinuxOptions.user) {
        seLinuxOptions.user = securityContext.seLinuxOptions.user
      }
      if(securityContext.seLinuxOptions.role) {
        seLinuxOptions.role = securityContext.seLinuxOptions.role
      }
      if(securityContext.seLinuxOptions.type) {
        seLinuxOptions.type = securityContext.seLinuxOptions.type
      }
      if(securityContext.seLinuxOptions.level) {
        seLinuxOptions.levle = securityContext.seLinuxOptions.level
      }
      if(seLinuxOptions.user || seLinuxOptions.role || seLinuxOptions.type || seLinuxOptions.level) {
        specSc.seLinuxOptions = seLinuxOptions
      }
      if(securityContext.runAsUser) specSc.runAsUser = securityContext.runAsUser
      if(securityContext.runAsGroup) specSc.runAsGroup = securityContext.runAsGroup
      if(securityContext.runAsNonRoot) specSc.runAsNonRoot = securityContext.runAsNonRoot
      if(securityContext.sysctls && securityContext.sysctls.length > 0) specSc.sysctls = securityContext.sysctls
      if(JSON.stringify(specSc) !== '{}') spec.securityContext = specSc

      var affinity = {}
      if(this.dsInfo.affinity.nodeAffinity.length > 0) {
        var required = []
        var preferred = []
        for(var na of this.dsInfo.affinity.nodeAffinity) {
          var labelSelectors = []
          var fieldSelectors = []
          for(var s of na.nodeSelectorTerms) {
            var st = {
              key: s.key,
              operator: s.operator
            }
            console.log(s.operator)
            if(['In', 'NotIn'].indexOf(s.operator) >= 0) {
              var values = s.values.split(',')
              st.values = values
            }
            else if(['Gt', 'Lt'].indexOf(s.operator) >= 0) {
              try{
                var values = parseInt(s.vlaues)
              } catch (err) {
                Message.error("节点亲和性操作为Gt/Lt时，值只能为数字")
                return {spec: null, success: false}
              }
              st.values = [vlaues]
            }
            if(s.type === 'label') labelSelectors.push(st)
            else if(s.type === 'field') fieldSelectors.push(st)
          }
          var nodeTerm = {}
          if(labelSelectors.length > 0) nodeTerm.matchExpressions = labelSelectors
          if(fieldSelectors.length > 0) nodeTerm.matchFields = fieldSelectors
          if(na.type === 'required') required.push(nodeTerm)
          else if(na.type === 'preferred') {
            if(!na.weight || na.weight < 1 || na.weight > 100) {
              Message.error("节点亲和性权重值范围为1-100")
              return {spec: null, success: false}
            }
            preferred.push({
              weight: parseInt(na.weight),
              preference: nodeTerm
            })
          }
        }
        var nodeAff = {}
        if(required.length > 0) {
          nodeAff.requiredDuringSchedulingIgnoredDuringExecution = {
            nodeSelectorTerms: required
          }
        }
        if(preferred.length > 0) {
          nodeAff.preferredDuringSchedulingIgnoredDuringExecution = preferred
        }
        if(nodeAff.requiredDuringSchedulingIgnoredDuringExecution || nodeAff.preferredDuringSchedulingIgnoredDuringExecution) {
          affinity.nodeAffinity = nodeAff
        }
      }
      var {podAffinity, success} = this.getPodAffinity(this.dsInfo.affinity.podAffinity)
      if(success) {
        if(podAffinity) affinity.podAffinity = podAffinity
      } else {
        return {spec: null, success: false}
      }
      var {podAffinity, success} = this.getPodAffinity(this.dsInfo.affinity.podAntiAffinity)
      if(success) {
        if(podAffinity) affinity.podAntiAffinity = podAffinity
      } else {
        return {spec: null, success: false}
      }

      if(JSON.stringify(affinity) !== '{}') {
        spec.affinity = affinity
      }

      var {volumes, success} = this.getSpecVolumes()
      if(!success) {
        return {spec: null, success: false}
      }
      if(volumes && volumes.length > 0) spec.volumes = volumes

      var {initContainers, containers, success} = this.getSpecContainers()
      // console.log(initContainers, containers, success)
      if(!success) return {spec: null, success: false}
      if(initContainers && initContainers.length > 0) spec.initContainers = initContainers
      if(containers && containers.length > 0) {
        spec.containers = containers
      } else {
        Message.error("必须要有一个运行容器")
        return {spec: null, success: false}
      }
      
      return {spec, success:true}
    },
    getSpecContainers() {
      var initContainers = []
      var containers = []
      if(this.containers && this.containers.length > 0) {
        console.log(this.containers)
        for(var c of this.containers) {
          console.log(c)
          if(!c.name) {
            Message.error("容器名称不能为空")
            return {initContainers, containers, success:false}
          }
          var container = {
            name: c.name,
            image: c.image,
          }
          if(c.imagePullPolicy) container.imagePullPolicy = c.imagePullPolicy
          if(c.commands) {
            try{
              container.command = JSON.parse(c.commands)
            } catch(err) {
              Message.error("容器命令格式错误，如：[\"bash\", \"-c\"]")
              return {initContainers, containers, success:false}
            }
          }
          if(c.args) {
            try{
              container.args = JSON.parse(c.args)
            } catch(err) {
              Message.error("容器命令参数格式错误，如：[\"-c\", \"echo hello\"]")
              return {initContainers, containers, success:false}
            }
          }
          if(c.workDir) container.workDir = c.workDir
          if(c.ports && c.ports.length > 0) {
            for(var p of c.ports) {
              p.containerPort = parseInt(p.containerPort)
            }
            container.ports = c.ports
          }
          if(c.env && c.env.length > 0) {
            var envs = []
            for(var e of c.env) {
              var env = {
                name: e.name
              }
              if(e.type === 'value') {
                env.value = e.value
              } else if(e.type === 'field') {
                var valueFrom = {
                  fieldRef: {
                    fieldPath: e.value
                  }
                }
                env.valueFrom = valueFrom
              } else if(e.type === 'resource') {
                var valueFrom = {
                  resourceFieldRef: {
                    resource: e.value
                  }
                }
                env.valueFrom = valueFrom
              } else if(e.type === 'configMap') {
                var valueFrom = {
                  configMapKeyRef: {
                    name: e.value.name,
                    key: e.key
                  }
                }
                env.valueFrom = valueFrom
              } else if(e.type === 'secret') {
                var valueFrom = {
                  secretKeyRef: {
                    name: e.value.name,
                    key: e.key
                  }
                }
                env.valueFrom = valueFrom
              }
              envs.push(env)
            }
            container.env = envs
          }
          var limits = {}
          if(c.resources.limits.cpu) limits.cpu = parseFloat(c.resources.limits.cpu)
          if(c.resources.limits.memory) limits.memory = `${c.resources.limits.memory}Mi`
          var requests = {}
          if(c.resources.requests.cpu) requests.cpu = parseFloat(c.resources.requests.cpu)
          if(c.resources.requests.memory) requests.memory = `${c.resources.requests.memory}Mi`
          var resources = {}
          if(limits.cpu || limits.memory) {
            resources.limits = limits
          }
          if(requests.cpu || requests.memory) {
            resources.requests = requests
          }
          if(resources.limits || resources.requests) container.resources = resources
          if(c.volumeMounts.length > 0) container.volumeMounts = c.volumeMounts
          if(c.livenessProbe.probe) {
            var {probe, success} = this.getProbe(c.livenessProbe)
            if(!success) return {initContainers, containers, success:false}
            container.livenessProbe = probe
          }
          if(c.readinessProbe.probe) {
            var {probe, success} = this.getProbe(c.readinessProbe)
            if(!success) return {initContainers, containers, success:false}
            container.readinessProbe = probe
          }

          var {securityContext, success} = this.getContainerSC(c.securityContext)
          if(!success) {
            return {initContainers, containers, success:false}
          }
          if(JSON.stringify(securityContext) !== '{}') container.securityContext = securityContext

          if(c.init) {
            initContainers.push(container)
          } else {
            containers.push(container)
          }
        }
      }
      return {initContainers, containers, success:true}
    },
    getSpecVolumes() {
      var vs = []
      if(this.dsInfo.volumes && this.dsInfo.volumes.length > 0) {
        for(var v of this.dsInfo.volumes) {
          var vol = {
            name: v.name,
          }
          if(v.type === 'persistentVolumeClaim') {
            vol.persistentVolumeClaim = {
              claimName: v.persistentVolumeClaim.name
            }
          } else if(v.type === 'hostPath') {
            vol.hostPath = {
              path: v.hostPath.path
            }
          } else if(v.type === 'emptyDir') {
            vol.emptyDir = {}
          } else if(v.type === 'secret') {
            vol.secret = {
              secretName: v.secret.obj.name
            }
            if(v.secret.defaultMode) vol.secret.defaultMode = v.secret.defaultMode
            if(v.secret.items && v.secret.items.length > 0) {
              var is = []
              for(let i of v.secret.items) {
                if(!i.key) {
                  Message.error("存储卷secret键不能为空")
                  return {volumes:null, success:false}
                }
                let it = {
                  key: i.key,
                  path: i.path,
                }
                if(i.mode) it.mode = i.mode
                is.push(it)
              }
              vol.secret.items = is
            }
          } else if(v.type === 'configMap') {
            vol.configMap = {
              name: v.configMap.obj.name
            }
            if(v.configMap.defaultMode) vol.configMap.defaultMode = v.configMap.defaultMode
            if(v.configMap.items && v.configMap.items.length > 0) {
              var is = []
              for(let i of v.configMap.items) {
                if(!i.key) {
                  Message.error("存储卷secret键不能为空")
                  return {volumes:null, success:false}
                }
                let it = {
                  key: i.key,
                  path: i.path,
                }
                if(i.mode) it.mode = i.mode
                is.push(it)
              }
              vol.configMap.items = is
            }
          } else if (v.type === 'nfs') {
            if(v.nfs.server === '') {
              Message.error("存储卷NFS服务IP不能为空")
              return {volumes:null, success:false}
            }
            if(v.nfs.path === '') {
              Message.error("存储卷NFS路径不能为空")
              return {volumes:null, success:false}
            }
            vol.nfs = {
              server: v.nfs.server,
              path: v.nfs.path
            }
            if(v.nfs.readOnly) vol.nfs.readOnly = v.nfs.readOnly
          } else if (v.type === 'glusterfs') {
            if(v.glusterfs.endpoints === '') {
              Message.error("存储卷GlusterFS端点不能为空")
              return {volumes:null, success:false}
            }
            if(v.glusterfs.path === '') {
              Message.error("存储卷GlusterFS路径不能为空")
              return {volumes:null, success:false}
            }
            vol.glusterfs = {
              endpoints: v.glusterfs.endpoints,
              path: v.glusterfs.path
            }
            if(v.glusterfs.readOnly) vol.glusterfs.readOnly = v.glusterfs.readOnly
          }
          vs.push(vol)
        }
      }
      return {volumes:vs, success:true}
    },
    getPodAffinity(podAffinity) {
      console.log(podAffinity)
      var required = []
      var preferred = []
      for(var pa of podAffinity) {
        if(!pa.podAffinityTerm.topologyKey) {
          Message.error("Pod亲和性中拓扑键不能为空")
          return {podAffinity: null, success: false}
        }
        var expressions = []
        var labels = {}
        for(var s of pa.podAffinityTerm.labelSelector) {
          if(s.operator === 'Equal') labels[s.key] = s.values
          else if(['In', 'NotIn'].indexOf(s.operator) >= 0) {
            var values = s.values.split(',')
            expressions.push({
              key: s.key,
              operator: s.operator,
              values: values
            })
          } else if (['Exists', 'DoseNotExist'].indexOf(s.operator) >= 0) {
            expressions.push({
              key: s.key,
              operator: s.operator
            })
          }
        }
        var selector = {}
        if(expressions.length > 0) {
          selector.matchExpressions = expressions
        }
        if(JSON.stringify(labels) !== '{}') {
          selector.matchLabels = labels
        }
        var term = {
          namespaces: pa.podAffinityTerm.namespaces,
          topologyKey: pa.podAffinityTerm.topologyKey,
        }
        if(selector.matchExpressions || selector.matchLabels) term.labelSelector = selector
        if(pa.type === 'required') {
          required.push(term)
        } else if(pa.type === 'preferred') {
          if(!pa.weight || pa.weight < 1 || pa.weight > 100) {
            Message.error("Pod亲和性权重值范围为1-100")
            return {podAffinity: null, success: false}
          }
          preferred.push({
            weight: parseInt(pa.weight),
            podAffinityTerm: term
          })
        }
      }
      var podAffinity = {}
      var has = false
      if(required.length > 0) {
        podAffinity.requiredDuringSchedulingIgnoredDuringExecution = required
        has = true
      }
      if(preferred.length > 0) {
        podAffinity.preferredDuringSchedulingIgnoredDuringExecution = preferred
        has = true
      }
      if(has) return {podAffinity, success: true}
      return {podAffinity: null, success: true}
    },
    getProbe(probe) {
      var containerProbe = {
        initialDelaySeconds: probe.initialDelaySeconds,
        timeoutSeconds: probe.timeoutSeconds,
        periodSeconds: probe.periodSeconds,
        successThreshold: probe.successThreshold,
        failureThreshold: probe.failureThreshold,
      }
      if (probe.type === 'command') {
        try{
          var command = JSON.parse(probe.handle.command)
        } catch(err) {
          console.log(err)
          return {probe: null, success: false}
        }
        containerProbe.exec = {
          command: command
        }
      } else if(probe.type === 'tcp') {
        containerProbe.tcpSocket = {
          port: probe.handle.port
        }
      } else if(['http', 'https'].indexOf(probe.type) >= 0) {
        var httpGet = {
          path: probe.handle.path,
          port: probe.handle.port
        }
        if(probe.type === 'https') {
          httpGet.scheme = 'HTTPS'
        }
        containerProbe.httpGet = httpGet
      }
      return {probe: containerProbe, success: true}
    },
    getContainerSC(securityContext) {
      var sc = {}
      if(securityContext.privileged) sc.privileged = securityContext.privileged
      var seLinuxOptions = {}
      if(securityContext.seLinuxOptions.user) {
        seLinuxOptions.user = securityContext.seLinuxOptions.user
      }
      if(securityContext.seLinuxOptions.role) {
        seLinuxOptions.role = securityContext.seLinuxOptions.role
      }
      if(securityContext.seLinuxOptions.type) {
        seLinuxOptions.type = securityContext.seLinuxOptions.type
      }
      if(securityContext.seLinuxOptions.level) {
        seLinuxOptions.levle = securityContext.seLinuxOptions.level
      }
      if(seLinuxOptions.user || seLinuxOptions.role || seLinuxOptions.type || seLinuxOptions.level) {
        sc.seLinuxOptions = seLinuxOptions
      }
      if(securityContext.runAsUser) sc.runAsUser = securityContext.runAsUser
      if(securityContext.runAsGroup) sc.runAsGroup = securityContext.runAsGroup
      if(securityContext.runAsNonRoot) sc.runAsNonRoot = securityContext.runAsNonRoot
      var capabilities = {}
      if(securityContext.addCapabilities && securityContext.addCapabilities.length > 0) {
        capabilities.add = securityContext.addCapabilities
      }
      if(securityContext.delCapabilities && securityContext.delCapabilities.length > 0) {
        capabilities.drop = securityContext.delCapabilities
      }
      if(capabilities.add || capabilities.drop) sc.capabilities = capabilities
      return {securityContext: sc, success: true}
    }
  }
}
</script>

<style lang="scss" scoped>
.dashboard {
  &-container {
    margin: 25px 30px;
  }
  &-text {
    font-size: 30px;
    line-height: 46px;
  }

  .table-fix {
    height: calc(100% - 100px);
  }
}

.name-class {
  cursor: pointer;
}
.name-class:hover {
  color: #409EFF;
}

.scrollbar-wrapper {
  overflow-x: hidden !important;
}
.el-scrollbar__bar.is-vertical {
  right: 0px;
}

.el-scrollbar {
  height: 100%;
}

.input-class {
  width: 250px;
}

.pnClass { 
  .el-select {
    width: 100%;
  }
}

.minus-btn-class {
  margin-left: 10px; 
  padding-left: 10px; 
  padding-right: 10px;
}

.border-div-class {
  margin: 25px 0px 10px 0px;
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

.item-header {
  color: #8B959C; 
  font-size: .85em;
}

.item-content {
  color: #8B959C; 
  font-size: .85em; 
  padding-bottom: 8px; 
  margin-right: 18px;
}

.card-div {
  border: 1px solid #DCDFE6; 
  box-shadow: 0 2px 4px 0 rgba(0,0,0,.12), 0 0 6px 0 rgba(0,0,0,.04); 
  padding: 15px;
}
</style>
