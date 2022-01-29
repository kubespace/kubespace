<template>
  <div>
    <el-row>
      <el-col :span="10" style="padding-right: 15px;">
        <div>
          <span :class="probe.type === 'http' ? 'probe-select' : 'probe-unselect'" 
            class="probe-class" @click="probe.type = 'http';">HTTP</span>
          <span :class="probe.type === 'https' ? 'probe-select' : 'probe-unselect'" 
            class="probe-class" @click="probe.type = 'https';">HTTPS</span>
          <span :class="probe.type === 'tcp' ? 'probe-select' : 'probe-unselect'" 
            class="probe-class" @click="probe.type = 'tcp';">TCP</span>
          <span :class="probe.type === 'command' ? 'probe-select' : 'probe-unselect'" 
            class="probe-class" @click="probe.type = 'command';">命令行</span>
        </div>
        <div style="margin-top: 10px;" v-if="['http', 'https'].indexOf(probe.type) >= 0">
          <el-row>
            <el-col :span="5">
              <span style="color: #8B959C; display: inline-block; width: 100px;">请求路径</span>
            </el-col>
            <el-col :span="19">
              <el-input v-model="probe.handle.path" size="small" placeholder="如：/health"
                ></el-input>
            </el-col>
          </el-row>
        </div>
        <div style="margin-top: 10px;" v-if="['http', 'https', 'tcp'].indexOf(probe.type) >= 0">
          <el-row>
            <el-col :span="5">
              <span style="color: #8B959C; display: inline-block; width: 100px;">端口</span>
            </el-col>
            <el-col :span="19">
              <el-input v-model="probe.handle.port" size="small" placeholder="如：80"
                ></el-input>
            </el-col>
          </el-row>
        </div>
        <div style="margin-top: 10px;" v-if="probe.type === 'command'">
          <el-row>
            <el-col :span="5">
              <span style="color: #8B959C; display: inline-block; width: 100px;">运行命令</span>
            </el-col>
            <el-col :span="19">
              <el-input v-model="probe.handle.command" size="small" placeholder='如：["/bin/bash", "-c", "ps -ef | grep 80"]'
                ></el-input>
            </el-col>
          </el-row>
        </div>
      </el-col>
      <el-col :span="12" style="border-left: 1px solid #DCDFE6; padding-left: 15px;">
        <el-row>
          <el-col :span="6" style="padding-right: 15px;">
            <div style="color: #8B959C; padding-bottom: 0px;">
              检查间隔
            </div>
            <el-input v-model="probe.periodSeconds" size="small" 
              placeholder="默认为10">
              <template slot="suffix">
                秒
              </template>  
            </el-input>
          </el-col>
          <el-col :span="6" style="padding-right: 15px;">
            <div style="color: #8B959C; padding-bottom: 0px;">
              初始延迟
            </div>
            <el-input v-model="probe.initialDelaySeconds" size="small" 
              placeholder="默认为0" controls-position="right">
              <template slot="suffix">
                秒
              </template>  
            </el-input>
          </el-col>
          <el-col :span="6" style="padding-right: 15px;">
            <div style="color: #8B959C; padding-bottom: 0px;">
              超时时间
            </div>
            <el-input v-model="probe.timeoutSeconds" size="small" 
              placeholder="默认为1">
              <template slot="suffix">
                秒
              </template>  
            </el-input>
          </el-col>
        </el-row>
        <el-row style="margin-top: 18px;">
          <el-col :span="12" style="padding-right: 15px;">
            <el-row>
              <el-col :span="8">
                <span style="color: #8B959C; display: inline-block; width: 100px;">成功阈值</span>
              </el-col>
              <el-col :span="16">
                <el-input-number controls-position="right" v-model="probe.successThreshold" size="small" placeholder="默认为1"
                  ></el-input-number>
              </el-col>
            </el-row>
          </el-col>
          <el-col :span="12" style="padding-right: 15px;">
            <el-row>
              <el-col :span="8">
                <span style="color: #8B959C; display: inline-block; width: 100px;">故障阈值</span>
              </el-col>
              <el-col :span="16">
                <el-input-number controls-position="right" v-model="probe.failureThreshold" size="small" placeholder="默认为1"
                  ></el-input-number>
              </el-col>
            </el-row>
          </el-col>
        </el-row>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { getToken } from '@/utils/auth'
import { Message } from 'element-ui'

export default {
  name: 'HealthProbe',
  data() {
    return {
    }
  },
  props: {
    probe: {
      type: Object,
      required: true,
      default: {}
    }
  },
  methods: {
  }
}
</script>

<style scoped>

.probe-class {
  border: 1px solid #DCDFE6; 
  border-radius: 3px; 
  padding: 5px 15px; 
  font-size: 12px; 
  cursor: pointer;
  color: #606266;
  font-weight: 500;
  margin-right: 10px;
}

.probe-unselect {
  cursor: pointer;
}

.probe-unselect:hover {
  border-color: #409EFF;
  color: #409EFF;
}

.probe-select {
  color: #409EFF;
  background: #ecf5ff;
  border-color: #b3d8ff;
}
</style>
