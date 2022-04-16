<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="openCreateFormDialog" createDisplay="创建空间"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        :data="workspaces"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort = "{prop: 'name'}"
        row-key="name"
      >
        <el-table-column prop="name" label="名称" show-overflow-tooltip min-width="20">
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.id)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip min-width="40">
          <template slot-scope="scope">
            {{ scope.row.type == 'code' ? scope.row.code_url : scope.row.description }}
          </template>
        </el-table-column>
        <el-table-column prop="update_user" label="操作人" min-width="20" show-overflow-tooltip>
        </el-table-column>
        <el-table-column prop="update_time" label="更新时间" min-width="20" show-overflow-tooltip>
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.update_time) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="180">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :underline="false" style="margin-right: 15px; color:#409EFF" @click="nameClick(scope.row.id)">流水线</el-link>
              <el-link :disabled="!$editorRole(scope.row.id)" :underline="false" type="primary" style="margin-right: 15px;"  @click="openUpdateFormDialog(scope.row)">编辑</el-link>
              <el-link :disabled="!$adminRole(scope.row.id)" :underline="false" type="danger" @click="handleDeleteWorkspace(scope.row.id, scope.row.name)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="updateFormVisible ? '修改流水线空间' : '创建流水线空间'" :visible.sync="createFormVisible"
        @close="closeFormDialog" :destroy-on-close="true">
        <div class="dialogContent">
          <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
            <el-form-item label="空间类型" prop="type">
              <span v-if="updateFormVisible">{{ workspaceTypeMap[form.type] }}</span>
              <el-radio-group v-else v-model="form.type">
                <el-radio label="code">代码空间</el-radio>
                <!-- <el-radio label="custom">自定义空间</el-radio> -->
              </el-radio-group>
            </el-form-item>
            <el-form-item label="代码类型" prop="codeType" v-if="form.type == 'code'" @change="codeTypeChange">
              <!-- <span v-if="updateFormVisible">{{ codeTypeMap[form.codeType] }}</span> -->
              <el-radio-group v-model="form.codeType" @change="codeTypeChange" :disabled="updateFormVisible">
                <el-radio label="https">HTTPS</el-radio>
                <el-radio label="git">GIT</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="代码地址" prop="codeUrl" v-if="form.type == 'code'">
              <el-input v-model="form.codeUrl" autocomplete="off" :disabled="updateFormVisible"
                :placeholder="form.codeType == 'https' ? '请输入代码地址，如: https://github.com/kubespace/kubespace.git' : '请输入代码地址，如: git@github.com:kubespace/kubespace.git'" 
                size="small"></el-input>
            </el-form-item>
            <el-form-item label="访问密钥" prop="codeSecretId" v-if="form.type=='code'">
              <el-select v-model="form.codeSecretId" placeholder="请选择代码访问密钥" size="small" style="width: 100%">
                <el-option
                  v-for="secret in form.codeType == 'https' ? codePasswordSecrets : codeKeySecrets"
                  :key="secret.id"
                  :label="secret.name"
                  :value="secret.id">
                </el-option>
              </el-select>
            </el-form-item>
            <el-form-item v-if="form.type == 'custom'" label="名称" prop="name">
              <el-input v-model="form.name" autocomplete="off" placeholder="请输入空间名称" size="small"></el-input>
            </el-form-item>
            <el-form-item v-if="form.type == 'custom'" label="描述" prop="description">
              <el-input v-model="form.name" type="textarea" autocomplete="off" placeholder="请输入空间描述" size="small"></el-input>
            </el-form-item>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter">
          <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="updateFormVisible ? handleUpdateWorkspace() : handleCreateWorkspace()" >确 定</el-button>
        </div>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listWorkspaces, createWorkspace, deleteWorkspace, updateWorkspace } from '@/api/pipeline/workspace'
import { listSecret } from "@/api/settings/secret";
import { Message } from 'element-ui'

export default {
  name: 'PipelineWorkspace',
  components: {
    Clusterbar
  },
  data() {
    return {
      titleName: ["流水线空间"],
      search_name: '',
      users: [],
      cellStyle: {border: 0},
      maxHeight: window.innerHeight - 150,
      loading: true,
      workspaces: [],
      secrets: [],
      createFormVisible: false,
      updateFormVisible: false,
      form: {
        name: '',
        type: 'code',
        codeType: "https",
        codeUrl: "",
        codeSecretId: ""
      },
      workspaceTypeMap: {
        code: "代码空间",
        custom: "自定义空间"
      },
      codeTypeMap: {
        https: "HTTPS",
        git: "GIT"
      },
      rules: {
        name: [{ required: true, message: '请输入空间名称', trigger: 'blur' },],
        type: [{ required: true, message: '请选择空间类型', trigger: 'blur' },],
        codeType: [{ required: true, message: '', trigger: 'blur' },],
        codeUrl: [{ required: true, message: ' ', trigger: 'blur' },],
        codeSecretId: [{ required: true, message: ' ', trigger: 'blur' },],
      },
    }
  },
  created() {
    this.fetchWorkspaces();
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 150
        console.log(heightStyle)
        that.maxHeight = heightStyle
      })()
    }
  },
  computed: {
    codePasswordSecrets() {
      let secrets = [];
      for(let s of this.secrets) {
        if(s.type=='password') secrets.push(s)
      }
      return secrets
    },
    codeKeySecrets() {
      let secrets = [];
      for(let s of this.secrets) {
        if(s.type=='key') secrets.push(s)
      }
      return secrets
    }
  },
  methods: {
    fetchWorkspaces() {
      this.loading = true
      listWorkspaces()
        .then((response) => {
          this.loading = false
          this.workspaces = response.data || [];
        })
        .catch(() => {
          this.loading = false
        })
    },
    fetchSecrets() {
      listSecret().then((resp) => {
        this.secrets = resp.data ? resp.data : []
      }).catch((err) => {
        console.log(err)
      })
    },
    nameClick: function(id) {
      this.$router.push({name: 'pipeline', params: {'workspaceId': id}})
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    handleCreateWorkspace() {
      if(!this.form.type) {
        Message.error("请选择流水线空间类型");
        return
      }
      let workspace = {type: this.form.type}
      if(this.form.type == 'code') {
        if(!this.form.codeUrl) {
          Message.error("代码地址不能为空");
          return
        }
        workspace['code_url'] = this.form.codeUrl
        workspace['code_type'] = this.form.codeType
        workspace['code_secret_id'] = this.form.codeSecretId
      } else {
        if(!this.form.name) {
          Message.error("流水线空间名称不能为空");
          return
        }
        workspace['name'] = this.form.name
        workspace['description'] = this.form.description
      }
      createWorkspace(workspace).then(() => {
        this.createFormVisible = false;
        Message.success("创建流水线空间成功")
        this.fetchWorkspaces()
      }).catch((err) => {
        console.log(err)
      });
    },
    handelUpdateWorkspace() {

    },
    handleDeleteWorkspace(id, name) {
      if(!id) {
        Message.error("获取空间id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${name}」此流水线空间?同时会删除该空间下的所有流水线。`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          deleteWorkspace(id).then(() => {
            Message.success("删除流水线空间成功")
            this.fetchWorkspaces()
          }).catch((err) => {
            console.log(err)
          });
        }).catch(() => {       
        });
    },
    openCreateFormDialog() {
      this.form = {
        name: '',
        type: 'code',
        codeType: "https",
        codeUrl: "",
        codeSecretId: ""
      }
      this.createFormVisible = true;
      this.fetchSecrets()
    },
    openUpdateFormDialog(object) {
      this.form = {
        name: object.name,
        type: object.type,
        codeType: object.code_type,
        codeUrl: object.code_url,
        codeSecretId: object.code_secret_id,
      }
      this.updateFormVisible = true;
      this.createFormVisible = true;
      this.fetchSecrets()
    },
    closeFormDialog() {
      this.updateFormVisible = false; 
      this.createFormVisible = false;
    },
    codeTypeChange() {
      this.form.codeUrl = ''
      this.form.codeSecretId = ''
    }
  },
}
</script>

<style lang="scss" scoped>
.member-bar {
  transition: width 0.28s;
  height: 55px;
  overflow: hidden;
  box-shadow: inset 0 0 4px rgba(0, 21, 41, 0.1);
  margin: 20px 20px 0px;

  .app-breadcrumb.el-breadcrumb {
    display: inline-block;
    font-size: 20px;
    line-height: 55px;
    margin-left: 8px;

    .no-redirect {
      // color: #97a8be;
      cursor: text;
      margin-left: 15px;
      font-size: 23px;
      font-family: Avenir, Helvetica Neue, Arial, Helvetica, sans-serif;
    }
  }

  .icon-create {
    display: inline-block;
    line-height: 55px;
    margin-left: 20px;
    width: 1.8em;
    height: 1.8em;
    vertical-align: 0.8em;
    color: #bfbfbf;
  }

  .right {
    float: right;
    height: 100%;
    line-height: 55px;
    margin-right: 25px;

    .el-input {
      width: 195px;
      margin-left: 15px;
    }

    .el-select {
      .el-select__tags {
        white-space: nowrap;
        overflow: hidden;
      }
    }
  }
}
.dashboard {
  &-container {
    margin: 10px 30px;
    height: calc(100%);
  }
  &-text {
    font-size: 30px;
    line-height: 46px;
  }
}

.table-fix {
  height: calc(100% - 100px);
}

.name-class {
  cursor: pointer;
}
.name-class:hover {
  color: #409EFF;
}
</style>
