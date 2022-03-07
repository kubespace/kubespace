<template>
  <div class="template-class">
    <el-form label-position="left" :model="template.metadata" :rules="rules" style="padding: 10px 20px 10px; color: #909399" label-width="100px">
      <el-form-item label="服务名称" style="width: 400px" prop="name">
        <el-input v-model="template.metadata.name" placeholder="请输入服务名称" size="small"></el-input>
      </el-form-item>
      <el-form-item label="服务类型" style="width: 700px" prop="" required>
        <el-radio-group v-model="template.spec.type"  size="small">
          <el-radio-button label="ClusterIP"></el-radio-button>
          <el-radio-button label="NodePort"></el-radio-button>
          <el-radio-button label="LoadBalancer"></el-radio-button>
          <el-radio-button label="ExternalName"></el-radio-button>
        </el-radio-group>
      </el-form-item>

      <el-form-item label="端口配置" required >
        <el-row style="margin-bottom: 5px; margin-top: 8px;">
          <el-col :span="4" style="background-color: #F5F7FA; padding-left: 10px;">
            <div class="border-span-header">
              <span  class="border-span-content">*</span>容器端口
            </div>
          </el-col>
          <el-col :span="4" style="background-color: #F5F7FA">
            <div class="border-span-header">
              名称
            </div>
          </el-col>
          <el-col :span="4" style="background-color: #F5F7FA">
            <div class="border-span-header">
              <span  class="border-span-content">*</span>服务端口
            </div>
          </el-col>
          <el-col v-if="template.spec.type == 'NodePort'" :span="4" style="background-color: #F5F7FA">
            <!-- <div class="border-span-header"> -->
              NodePort
            <!-- </div> -->
          </el-col>
          <el-col :span="4" style="background-color: #F5F7FA">
            <div class="border-span-header">
              协议
            </div>
          </el-col>
          <!-- <el-col :span="5"><div style="width: 100px;"></div></el-col> -->
        </el-row>
        <el-row style="padding-top: 0px;" v-for="(item, idx) in template.spec.ports" :key="idx">
          <el-col :span="4">
            <div class="border-span-header">
              <!-- <el-input v-model="item.containerPort" size="small" style="padding-right: 10px" placeholder="容器访问端口，如:80"></el-input> -->
              <el-select v-model="item.targetPort" placeholder="容器访问端口" size="small" style="width: 100%; padding-right: 10px">
                <!-- <el-option label="TCP" value="TCP"></el-option>
                <el-option label="UDP" value="UDP"></el-option>
                <el-option label="SCTP" value="SCTP"></el-option> -->
                <el-option v-for="p in containerPorts" :key="p.containerPort" :label="p.containerPort" :value="p.containerPort">
                </el-option>
              </el-select>
            </div>
          </el-col>
          <el-col :span="4">
            <div class="border-span-header">
              <el-input v-model="item.name" size="small" style="padding-right: 10px" placeholder="服务端口名称"></el-input>
            </div>
          </el-col>
          <el-col :span="4">
            <div class="border-span-header">
              <el-input v-model="item.port" size="small" style="padding-right: 10px" placeholder="服务暴露端口"></el-input>
            </div>
          </el-col>
          <el-col v-if="template.spec.type == 'NodePort'" :span="4">
            <!-- <div class="border-span-header"> -->
              <el-input v-model="item.nodePort" size="small" style="padding-right: 10px" placeholder="宿主机暴露端口"></el-input>
            <!-- </div> -->
          </el-col>
          <el-col :span="4">
            <div class="border-span-header">
              <el-select v-model="item.protocol" placeholder="端口所属协议" size="small">
                <el-option label="TCP" value="TCP"></el-option>
                <el-option label="UDP" value="UDP"></el-option>
                <el-option label="SCTP" value="SCTP"></el-option>
              </el-select>
            </div>
          </el-col>
          <el-col :span="2" style="padding-left: 10px">
            <el-button circle size="mini" style="padding: 5px;" 
              @click="template.spec.ports.splice(idx, 1)" icon="el-icon-close"></el-button>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="template.spec.type == 'NodePort' ? 20 : 16">
          <el-button style="width: 100%; border-radius: 0px; padding: 9px 15px; border-color: rgb(102, 177, 255); color: rgb(102, 177, 255)" plain size="mini" 
            @click="template.spec.ports.push({protocol: 'TCP'})" icon="el-icon-plus">添加服务端口</el-button>
          </el-col>
        </el-row>
      </el-form-item>
    </el-form>
    <el-form label-position="left" label-width="0" style="padding: 0px 25px; color: #909399">
      
    </el-form>
    
  </div>
</template>

<script>

export default {
  name: 'Service',
  components: {
  },
  data() {
    return {
      rules: {
        'name': [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
      },
    }
  },
  props: ['template', 'containers'],
  computed: {
    containerPorts() {
      let ports = []
      for(let c of this.containers) {
        for(let p of c.ports) {
          ports.push(p)
        }
      }
      console.log(ports)
      return ports
    }
  },
  methods: {
    
  }
}
</script>

<style scoped lang="scss">

.border-span-content {
  color: #F56C6C; 
  margin-right: 4px;
}
</style>

<style lang="scss">
</style>