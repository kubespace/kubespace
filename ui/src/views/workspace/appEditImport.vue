<template>
  <div>
    <clusterbar :titleName="titleName" :titleLink="['workspaceApp']">
      <div slot="right-btn">
        <el-button size="small" class="bar-btn" type="" @click="cancelSave">取 消</el-button>
        <el-button size="small" class="bar-btn" type="primary" @click="createAppDialog">保 存</el-button>
      </div>
    </clusterbar>
    <div class="dashboard-container" :style="{height: maxHeight + 'px'}" ref="tableCot" v-loading="loading">
      <div class="chart-content">
        <div class="menu">
          <div class="menu-head">
            <span>Chart资源目录</span>
            <span class="menu-operator">
              <el-tooltip class="item" effect="dark" content="新建文件" placement="top">
                <el-link :underline="false" class="menu-operator-item" @click="treeAddFile('add_file')" icon="el-icon-document-add"></el-link>
              </el-tooltip>
              <el-tooltip class="item" effect="dark" content="新建文件夹" placement="top">
                <el-link :underline="false" class="menu-operator-item"  @click="treeAddFile('add_dir')" icon="el-icon-folder-add"></el-link>
              </el-tooltip>
              <el-tooltip class="item" effect="dark" content="删除选中" placement="top">
                <el-link :underline="false" class="menu-operator-item" @click="treeDel" icon="el-icon-circle-close"></el-link>
              </el-tooltip>
            </span>
          </div>
          <div class="menu-body">
            <el-tree
              ref="tree-menu"
              :data="chartFilesData"
              node-key="id"
              icon-class="el-icon-arrow-right"
              highlight-current
              @node-click="handleNodeClick">
                <div slot-scope="{node, data}" class="custom-tree-node" :title="node.label">
                  <div v-if="!data.op_type">
                    <i v-if="data.children" style="margin-right: 3px;"
                        :class="data.children.length && node.expanded ? 'el-icon-folder-opened' : 'el-icon-folder'"></i>
                    <span :class="data.id == selectFileObject.id ? 'tree-select-file': ''">{{ node.label }}</span>
                  </div>
                  <div v-else>
                    <i v-if="data.op_type=='add_dir'" style="margin-right: 3px;" class="el-icon-folder"></i>
                    <el-input v-model="data.label" :ref="'treeInput-' + data.id" size="mini" placeholder="新建文件"
                    @blur="treeInputBlur" @change="treeInputBlur"></el-input>
                  </div>
                </div>
            </el-tree>
          </div>
        </div>
        <div class="editor" :style="'--yamlHeight: ' + (maxHeight - 32) + 'px'">
          <el-tabs v-model="editableTabsValue" type="card" :closable="true"
            @tab-remove="removeTab" @tab-click="tabClick">
            <el-tab-pane
              v-for="(item, index) in editableTabs"
              :key="item.id"
              :label="item.label"
              :name="item.id"
            >
            </el-tab-pane>
          </el-tabs>
          <div :style="{display: init || editableTabs.length ? 'block': 'none'}">
            <textarea  ref="textarea" />
          </div>
          <div :style="{display: init || editableTabs.length ? 'none': 'block', }">
            <el-empty description="空 空 如 也"></el-empty>
          </div>
        </div>
      </div>
    </div>
    <el-dialog title="保存应用版本" :visible.sync="createFormVisible"
       :destroy-on-close="true" :close-on-click-modal="false">
      <div>
        <div class="dialogContent" style="">
          <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
            <el-form-item label="应用名称" prop="name">
              <span>{{ form.name }}</span>
            </el-form-item>
            <el-form-item label="三位版本号" prop="version" required>
              <el-input v-model="form.version" placeholder="请输入应用三位版本号" size="small"></el-input>
            </el-form-item>
            <el-form-item label="第四位版本号" required>
              <el-input v-model="form.fourthVersion" placeholder="请输入应用第四位版本号" size="small"></el-input>
            </el-form-item>
            <el-form-item label="版本说明" required>
              <el-input type="textarea" v-model="form.version_description" placeholder="请输入应用版本说明" size="small"></el-input>
            </el-form-item>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter" style="padding-top: 25px;">
          <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="handleCreateApp" >创 建</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>
<script>
import { Clusterbar, Yaml } from "@/views/components";
import { getAppChartFiles } from '@/api/project/apps'
import {Base64} from 'js-base64';
import { Message } from "element-ui";
import CodeMirror from 'codemirror'
import 'codemirror/addon/lint/lint.css'
import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/base16-light.css'
import 'codemirror/mode/yaml/yaml'
import 'codemirror/addon/lint/lint'
import { createApp } from "@/api/project/apps"

export default {
  name: "workspace",
  components: {
    Clusterbar,
    Yaml,
  },
  mounted: function () {
    const that = this;
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - this.$contentHeight;
        that.maxHeight = heightStyle;
      })();
    };
    var ops = {
      lineNumbers: true, // 显示行号
      mode: 'text/x-yaml', // 语法model
      lint: false, // 开启语法检查
      lineWrapping: true,
    }
    this.yamlEditor = CodeMirror.fromTextArea(this.$refs.textarea, ops)
    
    this.yamlEditor.on('change', (cm) => {
      if(typeof(this.selectFileObject.content) == 'string') {
        this.selectFileObject.content = cm.getValue()
      }
    })
  },
  data() {
    return {
      init: true,
      maxHeight: window.innerHeight - this.$contentHeight,
      cellStyle: { border: 0 },
      titleName: ["应用管理", "编辑"],
      loading: true,
      treeId: 0,
      app: {},
      appVersion: {},
      chartFiles: {},
      chartFilesData: [],

      selectFileObject: {},

      treeAddObject: {},

      yamlEditor: null,

      tabClosable: false,
      editableTabs: [],
      editableTabsValue: "",

      form: {
        id: "",
        name: "",
        type: "ordinary_app",
        version: '0.0.1',
        fourthVersion: Math.ceil(Math.random() * 100000),
        description: '',
        version_description: "",
        templates: [],
      },
      rules: {
        name: [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
        version: [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
        description: [{ required: true, message: ' ', trigger: ['blur', 'change'] },],
      },
      createFormVisible: false,
    };
  },
  created() {
    this.fetchChartFiles();
  },
  computed: {
    projectId() {
      return parseInt(this.$route.params.workspaceId)
    },
    appVersionId() {
      return this.$route.params.appVersionId
    }
  },
  methods: {
    fetchChartFiles() {
      this.loading = true
      getAppChartFiles(this.appVersionId).then((resp) => {
        this.app = resp.data.app
        this.appVersion = resp.data.app_version
        this.chartFiles = resp.data.chart_files

        this.chartFilesData = this.chartFilesToTree(this.chartFiles)

        this.titleName = ["应用管理", "编辑", this.app.name]
        this.form.name = this.app.name
        this.form.type = this.app.type
        if(this.appVersion.app_version.indexOf("-")>0) {
          this.form.version = this.appVersion.app_version.split('-')[0]
        }

        for(let f of this.chartFilesData) {
          if(f.label == 'values.yaml') {
            this.$nextTick(() => {
                this.$refs['tree-menu'].setCurrentKey(f);
            });
            this.editableTabsValue = f.id
            this.editableTabs.push(f)
            this.selectFileObject = f
            this.yamlEditor.setValue(f.content)
            break
          }
        }
        this.init = false
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    treeDataSort(data) {
      return data.sort((x, y) => {
          if (x.label.toLowerCase() < y.label.toLowerCase()) {
              return 1;
          } else {
              return -1;
          }
      }).sort((x, y) => {
          if (!x.children && y.children) {
              return 1;
          } else {
              return -1;
          }
      });
    },
    chartFilesToTree(files) {
      let filesData = []
      for(let k in files) {
        if(typeof(files[k]) == 'object') {
          filesData.push({
            label: k,
            children: this.chartFilesToTree(files[k]),
            id: `${++this.treeId}`,
          })
        } else {
          let decodeOk = true
          let content = ''
          try{
            content = Base64.decode(files[k])
          }catch(e) {
            // base64解析失败
            console.log(e)
            decodeOk = false
            content = files[k]
          }
          filesData.push({
            label: k,
            decodeOk: decodeOk,
            content: content,
            id: `${++this.treeId}`,
          })
        }
      }
      return this.treeDataSort(filesData)
    },
    treeToChartFiles(tree) {
      let chartFiles = {}
      for(let t of tree) {
        if(t.children) {
          chartFiles[t.label] = this.treeToChartFiles(t.children)
        } else {
          if(!t.label) {
            continue
          }
          if(t.decodeOk) {
            chartFiles[t.label] = Base64.encode(t.content)
          } else {
            chartFiles[t.label] = t.content
          }
        }
      }
      return chartFiles
    },
    handleNodeClick(data) {
      // console.log(data)
      if(data.op_type) return
      if(!data.children) {
        if(data.id == this.selectFileObject.id) return
        let found = false
        for(let e of this.editableTabs) {
          if(e.id == data.id) {
            this.editableTabsValue = data.id
            found = true
            break
          }
        }
        if(!found) {
          this.editableTabs.push(data)
          this.editableTabsValue = data.id
        }
        if(this.editableTabs.length > 1) this.tabClosable = true
        this.selectFileObject = data
        // console.log(this.selectFileObject)
        this.yamlEditor.setValue(data.content)
      }
    },
    cancelSave() {
      this.$router.push({name: 'workspaceApp'})
    },
    createAppDialog() {
      // console.log(this.chartFilesData)
      this.form.fourthVersion = (Math.floor(new Date() / 1000)).toString(36)
      this.createFormVisible = true;
    },
    handleCreateApp() {
      if(!this.form.version) {
        Message.error("应用版本为空");
        return
      }
      if(!this.form.version_description) {
        Message.error("请输入版本说明")
        return
      }
      let version = this.form.version
      if(this.form.fourthVersion) {
        version += '-' + this.form.fourthVersion
      }
      let chartFiles = this.treeToChartFiles(this.chartFilesData)
      let data = {
        scope: "project_app",
        scope_id: parseInt(this.projectId), 
        name: this.form.name, 
        type: this.form.type,
        from: "import",
        version: version,
        values: Base64.decode(chartFiles["values.yaml"]),
        chart_files: chartFiles,
        description: this.form.description,
        version_description: this.form.version_description
      }
      this.loading = true
      createApp(data).then(() => {
        this.loading = false
        Message.success("创建应用成功")
        this.$router.push({name: 'workspaceApp'})
      }).catch((err) => {
        this.loading = false
      });
    },
    removeTab(targetName) {
      let tabs = this.editableTabs;
      let activeName = this.editableTabsValue;
      if (activeName === targetName) {
        tabs.forEach((tab, index) => {
          if (tab.id === targetName) {
            let nextTab = tabs[index + 1] || tabs[index - 1];
            if (nextTab) {
              this.editableTabsValue = nextTab.id;
              this.selectFileObject = nextTab
              this.yamlEditor.setValue(this.selectFileObject.content)
            }
          }
        });
      }
      this.editableTabs = tabs.filter(tab => tab.id !== targetName);
      // if(this.editableTabs.length == 1) this.tabClosable = false
      if(this.editableTabs.length == 0) {
        this.selectFileObject = {}
        this.$refs['tree-menu'].setCurrentKey(null)
      }
      
    },
    tabClick(tabObj) {
      for(let tab of this.editableTabs) {
        if(tab.id == tabObj.name) {
          // this.$refs['tree-menu'].setCurrentNode(tab)
          this.selectFileObject = tab
          this.yamlEditor.setValue(this.selectFileObject.content)
          break
        }
      }
    },
    treeInsertFile(type, parent) {
      let addFile = {op_type: type, id: `${++this.treeId}`, label: ''}
      if(type=='add_file') {
        addFile['content'] = ''
        addFile['decodeOk'] = true
      } else {
        addFile['children'] = []
      }
      this.treeAddObject = addFile
      let insertBefore 
      if(type == 'add_file') {
        for(let n of parent.children) {
          if(!n.children) {
            insertBefore = n
            break
          }
        }
      } else {
        if(parent.children.length > 0){
          insertBefore = parent.children[0]
        }
      }
      if(insertBefore) {
        this.$refs['tree-menu'].insertBefore(addFile, insertBefore)
      } else {
        this.$refs['tree-menu'].append(addFile, parent.id)
      }
      // this.chartFilesData.push(addFile)
      this.$nextTick(() => {
        this.$refs['treeInput-'+addFile.id].focus()
        // console.log(this.$refs['treeInput-'+addFile.id])
        // console.log(this.$refs['treeInput-'+addFile.id].focused)
      });
    },
    treeAddFile(type) {
      if(typeof(this.treeAddObject.op_type) == 'string') {
        Message.error(`当前有一个${this.treeAddObject.op_type=='add_dir' ? '文件夹': '文件'}正在新建`)
        return
      }
      let curNode = this.$refs['tree-menu'].getCurrentNode()
      // console.log(curNode)
      let parent
      if(!curNode) {
        parent = {children: this.chartFilesData}
      } else {
        let treeNode = this.$refs['tree-menu'].getNode(curNode.id)
        // console.log(treeNode)
        if(curNode.children) {
          parent = curNode
          let that = this
          treeNode.expand(function(){
            that.treeInsertFile(type, parent)
          }, true)
          return
        } else {
          parent = treeNode.parent
        }
        if(!parent.parent) {
          parent = {children: parent.data}
        } else {
          parent = parent.data
        }
      }
      // console.log(parent)
      // console.log(typeof(parent))
      this.treeInsertFile(type, parent)
    },
    // 获取tree当前节点以及所有子节点的文件
    getChildrenFiles(files) {
      let curFiles = []
      for(let f of files) {
        if(f.children) {
          let subFiles = this.getChildrenFiles(f.children)
          for(let sf of subFiles) {
            curFiles.push(sf)
          }
        } else {
          curFiles.push(f)
        }
      }
      return curFiles
    },
    treeDel() {
      let curNode = this.$refs["tree-menu"].getCurrentNode()
      if(!curNode) return
      // console.log(curNode)
      this.$refs["tree-menu"].remove(curNode)
      if(curNode.children) {
        let allFiles = this.getChildrenFiles(curNode.children)
        for(let f of allFiles) {
          this.removeTab(f.id)
        }
      } else {
        this.removeTab(curNode.id)
      }
    },
    treeInputBlur(e) {
      // console.log(this.treeAddObject)
      if(!this.treeAddObject.label) {
        this.$refs['tree-menu'].remove(this.treeAddObject)
      } else {
        this.treeAddObject.op_type = ''
        this.$refs['tree-menu'].setCurrentNode(this.treeAddObject)
        this.handleNodeClick(this.treeAddObject)
        // this.handleNodeClick(this.treeAddObject)
      }
      this.treeAddObject = {}
    },
  },
};
</script>

<style scoped lang="scss">
.chart-content {
  display: flex;
  height: 100%;
  border-radius: 10px;

  .menu {
    width: 20%;
    height: 100%;

    .menu-head {
        height: 32px;
        line-height: 32px;
        background: #fff;
        border-bottom: 1px solid #E4E7ED;
        font-weight: 450;
        border-top-left-radius: 10px;
        padding-left: 10px;
        font-size: 14px;
    }

    .menu-body {
        height: calc(100% - 32px);
    }

    .menu-operator {
      float: right; 
      margin-right: 13px; 
      margin-top: -1px;
    }
    .menu-operator-item {
      font-size: 18px;
      margin-left: 4px;
    }

    .tree-select-file {
      color: #409EFF;
    }
  }
  .editor {
    margin-left: 0px;
    height: 100%;
    width: 80%;
  }

}
</style>
<style lang="scss">
.chart-content {
  .menu {
    .el-tree {
      overflow: auto;
      height: 100%;
      width: 100%;
      border-bottom-left-radius: 10px;
    }
    .menu-operator {
      .el-link {
        // vertical-align: baseline;
      }
    }
    .el-input--mini .el-input__inner {
      height: 24px;
      line-height: 24px;
    }
  }

  .editor { 
    .CodeMirror {
      height: var(--yamlHeight);
      font-size: 13px;
      line-height: 1.2;
      // border-top-right-radius: 10px;
      border-bottom-right-radius: 10px;
    }

    .CodeMirror-scroll{
      height: var(--yamlHeight);
      overflow-y: hidden;
      overflow-x: auto;
    }
    .CodeMirror-linenumber {
      padding-left: 8px;
    }

    .el-tabs__header {
      height: 32px;
      margin-bottom: 0px;
      background-color: #fff;
      border-top-right-radius: 10px;
    }

    .el-tabs--card>.el-tabs__header .el-tabs__nav {
      border-radius: 0px;
    }
    .el-tabs__item {
      height: 31px;
      line-height: 31px;
    }
  }
}
</style>
