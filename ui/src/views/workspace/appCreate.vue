<template>
  <div>
    <clusterbar :titleName="titleName">
      <div slot="right-btn">
        <el-button size="small" class="bar-btn" type="" @click="cancelSaveApp">取 消</el-button>
        <el-button size="small" class="bar-btn" type="primary" @click="createAppDialog">保 存</el-button>
      </div>
    </clusterbar>
    <div v-loading="loading" class="dashboard-container app-create-container" :style="{margin: '0px', 'height': maxHeight + 'px'}" :max-height="maxHeight">
      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" class="base-info-class" :model="form" :rules="rules" style="padding: 10px 20px 0px" label-width="100px">
          <el-form-item label="应用名称" style="width: 400px" prop="name">
            <el-input v-model="form.name" :disabled="appVersionId?true:false" @change="appFormNameChange" placeholder="请输入应用名称" size="small"></el-input>
          </el-form-item>
          <el-form-item v-if="appVersionId" label="应用版本" style="width: 400px" prop="name">
            <span style="color: #606266">{{ chart.version }}</span>
          </el-form-item>
        </el-form>
      </div>
      <div style="padding: 0px 8px 0px;" class="app-template-class">
        <div style="margin-bottom: 13px;">
          <div style="display: inline-block">
            应用配置
          </div>
          <div style="display: inline-block; float: right; margin-right: 5px;">
            <el-dropdown>
              <span class="el-dropdown-link" style="cursor: pointer; color: #0c81f5;">
                添加资源<i class="el-icon-arrow-down el-icon--right"></i>
              </span>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item @click.native.prevent="resourceAddClick('Service')">Service</el-dropdown-item>
                <el-dropdown-item @click.native.prevent="resourceAddClick('ConfigMap')">ConfigMap</el-dropdown-item>
                <el-dropdown-item @click.native.prevent="resourceAddClick('Secret')">Secret</el-dropdown-item>
                <el-dropdown-item @click.native.prevent="resourceAddClick('PersistentVolumeClaim')">PersistentVolumeClaim</el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </div>
        </div>
        <el-tabs type="border-card" v-model="resourceTabVal" @tab-remove="resourceTabRemove" 
          :addable="false" style="padding: 0px;">
          <el-tab-pane v-for="(t, i) in form.templates" :key="i + ''" :name="i + ''" :closable="i==0  ? false : true">
            <div style="display: inherit; vertical-align: middle; font-size: 5px; line-height: 10px; font-weight: 400" slot="label">
              <div style="font-weight: 300;">{{ t.kind }}</div>
              <div style="display: block; font-size: 14px; line-height: 20px; align: center; font-weight: 500">{{ t.metadata.name ? t.metadata.name : ' unnaming'}}</div>
            </div>
            <div style="padding-bottom: 20px;">
              <workload v-if="workloadTypes.indexOf(t.kind) >= 0" :template="t" :noName="i == 0"></workload>
              <service v-if="t.kind == 'Service'" :template="t" :containers="form.templates[0].spec.template.spec.containers"></service>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>

      <el-dialog :title="updateFormVisible ? '升级应用' : '保存应用版本'" :visible.sync="createFormVisible"
      @close="closeFormDialog" :destroy-on-close="true" :close-on-click-modal="false">
        <div v-loading="installLoading">
          <div class="dialogContent" style="">
            <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
              <el-form-item label="应用名称" prop="name">
                <span>{{ form.name }}</span>
              </el-form-item>
              <el-form-item v-if="appVersionId" label="应用版本" prop="name">
                <span>{{ chart.version }}</span>
              </el-form-item>
              <el-form-item label="三位版本号" prop="version">
                <el-input v-model="form.version" placeholder="请输入应用三位版本号" size="small"></el-input>
              </el-form-item>
              <el-form-item label="第四位版本号">
                <el-input v-model="form.fourthVersion" placeholder="请输入应用第四位版本号" size="small"></el-input>
              </el-form-item>
              <el-form-item label="版本描述">
                <el-input type="textarea" v-model="form.description" placeholder="请输入应用版本描述" size="small"></el-input>
              </el-form-item>
            </el-form>
          </div>
          <div slot="footer" class="dialogFooter" style="padding-top: 25px;">
            <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
            <el-button type="primary" @click="updateFormVisible ? handleUpdateApp() : handleCreateApp()" >
              {{ createFormVisible ? '创 建' : '创 建' }}
            </el-button>
          </div>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { createApp, getAppVersion } from "@/api/project/apps";
import { Message } from "element-ui";
import { Workload, kindTemplate, Service, transferTemplate, resolveToTemplate } from '@/views/workspace/kinds'
import yaml from 'js-yaml'

export default {
  name: "appCreate",
  components: {
    Clusterbar,
    Workload,
    Service,
  },
  mounted: function () {
    const that = this;
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 130;
        that.maxHeight = heightStyle;
      })();
    };
  },
  data() {
    return {
      maxHeight: window.innerHeight - 130,
      cellStyle: { border: 0 },
      titleName: ["应用管理", "创建"],
      loading: false,
      createFormVisible: false,
      updateFormVisible: false,
      installLoading: false,
      resourceTabVal: 0,
      workloadTypes: ['Deployment', 'StatefulSet', 'DaemonSet', 'CronJob', 'Job'],
      chart: {
        templates: {},
        values: "",
        version: "",
      },
      form: {
        id: "",
        name: "",
        version: '0.0.1',
        fourthVersion: Math.ceil(Math.random() * 100000),
        description: '',
        templates: [],
      },
      rules: {
        name: [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
        version: [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
      },
    };
  },
  created() {
    if(this.appVersionId) {
      this.getAppVersion()
    } else {
      let defaultWorkloadTpl = kindTemplate('Workload')
      this.form.templates = [defaultWorkloadTpl]
    }
  },
  computed: {
    projectId() {
      return this.$route.params.workspaceId
    },
    appVersionId() {
      return this.$route.params.appVersionId
    }
  },
  methods: {
    kindTemplate,
    getAppVersion() {
      this.loading = true
      getAppVersion(this.appVersionId).then((resp) => {
        this.form.id = resp.data.id
        this.form.name = resp.data.name
        this.chart.version = resp.data.package_version
        this.form.version = this.chart.version.split('-')[0]
        let values = yaml.load(resp.data.values)
        for(let tpl of resp.data.templates) {
          let data = yaml.load(decodeURIComponent(atob(tpl.data)))
          resolveToTemplate(data)
          console.log(data)
          if(this.workloadTypes.indexOf(data.kind) > -1) {
            let podSpec = data.spec.template.spec
            for(let c of podSpec.containers) {
              c.image = values.workloads[data.metadata.name]['containers'][c.name]['image']
            }
          }
          this.form.templates.push(data)
        }
        this.loading = false
      }).catch((err) => {
        this.loading = false
      });
    },
    appFormNameChange(value) {
      this.form.templates[0].metadata.name = value
      this.form.templates[0].spec.template.spec.containers[0].name = value
      if(this.form.templates.length > 1 && this.form.templates[1].kind=='Service'){
        this.form.templates[1].metadata.name = value
      }
    },
    handleCreateApp() {
      if(!this.form.version) {
        Message.error("应用版本为空");
        return
      }
      let version = this.form.version
      if(this.form.fourthVersion) {
        version += '-' + this.form.fourthVersion
      }
      let chartDict = {
        apiVersion: 'v2',
        name: this.form.name,
        version: version,
        appVersion: version
      }
      let chartYaml = yaml.dump(chartDict)
      let data = {
        project_id: parseInt(this.projectId), 
        name: this.form.name, 
        version: version,
        chart: chartYaml,
        templates: this.chart.templates,
        values: this.chart.values
      }
      this.loading = true
      createApp(data).then(() => {
        this.loading = false
        Message.success("安装应用成功")
        this.$router.push({name: 'workspaceApp'})
      }).catch((err) => {
        this.loading = false
      });
    },
    closeFormDialog() {
      this.updateFormVisible = false; 
      this.createFormVisible = false;
    },
    createAppDialog() {
      if(!this.form.name) {
        Message.error("应用名称为空")
        return
      }
      this.chart.templates = {}
      let valuesDict = {workloads: {}}
      let idx = 0
      for(let template of this.form.templates) {
        let obj = transferTemplate(template)
        if(obj.err) {
          Message.error(obj.err)
          return
        }
        let tpl = obj.tpl
        if(this.workloadTypes.indexOf(template.kind) > -1) {
          let containers = {}
          for(let c of tpl.spec.template.spec.containers) {
            containers[c.name] = {image: c.image}
            c.image = `{{ .Values.workloads.${template.metadata.name}.containers.${c.name}.image }}`
          }
          valuesDict.workloads[template.metadata.name] = {containers: containers}
        }
        if(tpl.kind == 'Service') {
          tpl.spec.selector = {
            'kubespace.cn/app': this.form.name
          }
        }
        let tplName = `${idx < 10 ? '0'+idx : ''+idx}${tpl.metadata.name}-${tpl.kind}.yaml`.toLowerCase()
        this.chart.templates[tplName] = yaml.dump(tpl)
      }
      this.chart.values = yaml.dump(valuesDict)
      this.createFormVisible = true;
    },
    cancelSaveApp() {
      this.$router.push({name: 'workspaceApp'})
    },
    resourceTabRemove(removeIdx) {
      var i = parseInt(removeIdx)
      if(this.form.templates.length <= 1) return false
      var activeTab = this.resourceTabVal
      if (i < parseInt(this.resourceTabVal)) {
        activeTab = parseInt(this.resourceTabVal) - 1
      } else if(i === parseInt(this.resourceTabVal)) {
        activeTab = i - 1
        if (i <= 0) activeTab = 0;
      }
      
      this.form.templates.splice(i, 1)
      this.resourceTabVal = activeTab + '';
    },
    resourceAddClick(kind) {
      let tpl = kindTemplate(kind)
      if(kind == 'Service') {
        let hasService = false
        for(let tpl of this.form.templates) {
          if(tpl.kind == 'Service') {
            hasService = true
          }
        }
        if(!hasService) tpl.metadata.name = this.form.name
      }
      this.form.templates.push(tpl)
      this.resourceTabVal = (this.form.templates.length - 1) + ''
    },
  },
};
</script>


<style lang="scss" scoped>
@import "~@/styles/variables.scss";
.bar-btn {
  padding: 9px 25px
}

.table-fix {
  height: calc(100% - 100px);
}

.status-class {
  height: 13px; 
  width: 13px; 
  border: 1px solid; 
  border-color:rgb(121, 123, 129); 
  background-color: rgb(121, 123, 129);  
  display: inline-block;
  vertical-align: middle; 
  border-radius: 25px; 
  margin: 0px 5px 3px 0px;
}

.operator-btn {
  margin-right: 15px;
  color:#0c81f5
}

</style>

<style lang="scss">
.app-create-container {
  margin: 0px;
  padding: 10px 20px 0px;
  overflow: auto;

  input {
    border-radius: 0px;
  } 
  .el-form-item__label {
    width: 90px;
    color: #99a9bf;
    font-weight: 400;
  }
  .el-radio-button:first-child .el-radio-button__inner {
    border-radius: 0px;
  }
  .el-radio-button:last-child .el-radio-button__inner {
    border-radius: 0px;
  }
  .el-form-item {
    margin-bottom: 10px;
  }

  .app-template-class {
    .el-tabs--border-card>.el-tabs__content {
      padding: 0px;
    }
  }
}
</style>
