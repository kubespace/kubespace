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
            {{ scope.row.type == 'code' ? scope.row.code ? scope.row.code.clone_url : '' : scope.row.description }}
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
        <div v-loading="dialogLoading">
          <div class="dialogContent">
            <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
              <el-form-item label="空间类型" prop="type">
                <span v-if="updateFormVisible">{{ workspaceTypeMap[form.type] }}</span>
                <el-radio-group v-else v-model="form.type">
                  <el-radio label="code">代码空间</el-radio>
                  <el-radio label="custom">自定义空间</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="代码托管" prop="codeType" v-if="form.type == 'code'" @change="codeTypeChange">
                <el-radio-group v-model="form.codeType" @change="codeTypeChange" :disabled="updateFormVisible">
                  <el-radio label="github">GitHub</el-radio>
                  <el-radio label="gitlab">GitLab</el-radio>
                  <el-radio label="gitee">Gitee</el-radio>
                  <el-radio label="other">其他</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="GitLab地址" prop="apiUrl" v-if="form.codeType == 'gitlab'">
                <el-input v-model="form.apiUrl" autocomplete="off" :disabled="updateFormVisible"
                  placeholder="请输入Gitlab访问地址，如: https://gitlab.com" 
                  size="small"></el-input>
              </el-form-item>
              <el-form-item label="代码地址" prop="codeUrl" v-if="updateFormVisible || form.codeType == 'other'">
                <el-input v-model="otherCodeUrl" autocomplete="off" :disabled="updateFormVisible"
                  placeholder="请输入代码地址，如: git@github.com:kubespace/kubespace.git" 
                  size="small"></el-input>
              </el-form-item>
              <el-form-item :label="form.codeType=='other'?'访问密钥':'AccessToken'" prop="codeSecretId" v-if="form.type=='code'">
                <el-select v-model="form.codeSecretId" placeholder="请选择代码访问密钥" size="small" style="width: 100%"
                  @change="codeSecretChange">
                  <el-option v-for="secret in codeSecrets" :key="secret.id" :label="secret.name" :value="secret.id">
                  </el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="代码仓库" prop="codeUrl" v-if="!updateFormVisible && form.type=='code' && form.codeType!='other'">
                <el-select v-model="form.codeUrl" placeholder="请选择代码仓库" size="small" style="width: 100%" :loading="codeRepoLoading">
                  <el-option
                    v-for="repo in codeRepos"
                    :key="repo.clone_url"
                    :label="repo.full_name"
                    :value="repo.clone_url">
                  </el-option>
                </el-select>
              </el-form-item>
              <el-form-item v-if="form.type == 'custom'" label="名称" prop="name">
                <el-input v-model="form.name" autocomplete="off" placeholder="请输入空间名称" size="small"></el-input>
              </el-form-item>
              <el-form-item v-if="form.type == 'custom'" label="描述" prop="description">
                <el-input v-model="form.description" type="textarea" autocomplete="off" placeholder="请输入空间描述" size="small"></el-input>
              </el-form-item>
            </el-form>
          </div>
          <div slot="footer" class="dialogFooter" style="margin-top: 20px;">
            <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
            <el-button type="primary" @click="updateFormVisible ? handleUpdateWorkspace() : handleCreateWorkspace()" >确 定</el-button>
          </div>
        </div>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listWorkspaces, createWorkspace, deleteWorkspace, updateWorkspace, listGitRepos } from '@/api/pipeline/workspace'
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
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      dialogLoading: false,
      workspaces: [],
      secrets: [],
      createFormVisible: false,
      updateFormVisible: false,
      codeRepos: [],
      codeRepoLoading: false,
      otherCodeType: "",
      otherCodeUrl: "",
      form: {
        name: '',
        type: 'code',
        codeType: "github",
        codeUrl: "",
        codeSecretId: "",
        apiUrl: "https://gitlab.com",
        description: "",
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
        codeType: [{ required: true, message: '请选择代码仓库类型', trigger: 'blur' },],
        apiUrl: [{ required: true, message: '请输入Gitlab访问地址', trigger: 'blur' },],
        codeSecretId: [{ required: true, message: '请选择仓库密钥', trigger: 'blur' },],
        codeUrl: [{ required: true, message: '请输入仓库地址', trigger: 'blur' },],
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
        that.maxHeight = window.innerHeight - this.$contentHeight
      })()
    }
  },
  watch: {
    otherCodeUrl(n) {
      if(this.updateFormVisible) return
      this.form.codeUrl = n
      if(n.length >= 5 && n.substr(0, 5) == "https") {this.otherCodeType ="https"; return}
      if(n.length >= 4 && n.substr(0, 4) == "http") {this.otherCodeType ="https"; return}
      if(n.length >= 3 && n.substr(0, 3) =="git") {this.otherCodeType = "git"; return}
      this.otherCodeType = ''
    },
    otherCodeType() {
      if(this.updateFormVisible) return
      this.form.codeSecretId = '' 
    }
  },
  computed: {
    codeSecrets() {
      if(this.form.codeType != 'other') return this.secretAccessTokens;
      if(this.otherCodeType == "https") return this.codePasswordSecrets;
      if(this.otherCodeType == "git") return this.codeKeySecrets;
      return []
    },
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
    },
    secretAccessTokens() {
      let secrets = [];
      for(let s of this.secrets) {
        if(s.type=="token") secrets.push(s)
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
      console.log(this.form)
      let workspace = {type: this.form.type}
      if(this.form.type == 'code') {
        if (this.form.codeType == "") {
          Message.error("请选择代码托管类型")
          return
        }
        if(this.form.codeType =="other") {
          workspace['code_url'] = this.otherCodeUrl
          workspace['code_type'] = this.otherCodeType
        } else {
          workspace['code_url'] = this.form.codeUrl
          workspace['code_type'] = this.form.codeType
        }
        workspace['code_secret_id'] = this.form.codeSecretId
        if (this.form.codeType == "gitlab") {
          workspace['api_url'] = this.form.apiUrl
        }
        if(!workspace['code_url']) {
          Message.error("代码仓库不能为空");
          return
        }
        if (!workspace['code_secret_id']) {
          Message.error("请选择代码仓库密钥")
          return
        }
        if (workspace['code_type'] == "gitlab" && workspace['api_url'] == "") {
          Message.error("请输入gitlab访问地址")
          return
        }

      } else {
        if(!this.form.name) {
          Message.error("流水线空间名称不能为空");
          return
        }
        workspace['name'] = this.form.name
        workspace['description'] = this.form.description
      }
      this.dialogLoading = true
      createWorkspace(workspace).then(() => {
        this.dialogLoading = false
        this.createFormVisible = false;
        Message.success("创建流水线空间成功")
        this.fetchWorkspaces()
      }).catch((err) => {
        this.dialogLoading = false
        console.log(err)
      });
    },
    handleUpdateWorkspace() {
      if(!this.form.id) {
        Message.error("获取流水线空间id参数错误，请刷新重试");
        return
      }
      let workspace = {}
      if(this.form.type == 'code') {
        if(!this.form.codeUrl) {
          Message.error("代码地址不能为空");
          return
        }
        workspace['code_secret_id'] = this.form.codeSecretId
      } else {
        if(!this.form.name) {
          Message.error("流水线空间名称不能为空");
          return
        }
        workspace['description'] = this.form.description
      }
      this.dialogLoading = true
      updateWorkspace(this.form.id, workspace).then(() => {
        this.dialogLoading = false
        this.closeFormDialog()
        Message.success("更新流水线空间成功")
        this.fetchWorkspaces()
      }).catch((err) => {
        this.dialogLoading = false
        console.log(err)
      });
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
          this.loading = true
          deleteWorkspace(id).then(() => {
            this.loading = false
            Message.success("删除流水线空间成功")
            this.fetchWorkspaces()
          }).catch((err) => {
            this.loading = false
            console.log(err)
          });
        }).catch(() => {       
        });
    },
    openCreateFormDialog() {
      this.form = {
        name: '',
        type: 'code',
        codeUrl: "",
        codeType: "github",
        codeSecretId: "",
        apiUrl: "https://gitlab.com"
      }
      this.otherCodeUrl = ''
      this.otherCodeType = ''
      this.createFormVisible = true;
      this.updateFormVisible = false
      this.fetchSecrets()
    },
    openUpdateFormDialog(object) {
      this.form = {
        id: object.id,
        name: object.name,
        type: object.type,
        description: object.description,
        codeType: '',
        apiUrl: '',
        codeSecretId: 0,
        codeUrl: '',
      }
      if(object.type == 'code' && object.code) {
        if(['https', 'git'].indexOf(object.code.type)> -1) {
          this.form.codeType = 'other'
          this.otherCodeUrl = object.code.clone_url
          this.otherCodeType = object.code.type
        } else {
          this.form.codeType = object.code.type
          this.otherCodeUrl = object.code.clone_url
          if(this.form.codeType == 'gitlab') {
            this.form.apiUrl = object.code.api_url
          }
        }
        this.form.codeUrl = object.code.clone_url
        this.form.codeSecretId = object.code.secret_id
      }
      this.updateFormVisible = true;
      this.createFormVisible = true;
      this.fetchSecrets()
    },
    closeFormDialog() {
      this.createFormVisible = false;
    },
    codeTypeChange() {
      this.form.codeSecretId = ''
      this.codeSecretChange()
    },
    codeSecretChange() {
      if(this.form.codeType == "other") return
      if(this.updateFormVisible) return
      this.codeRepos = []
      this.form.codeUrl = ""
      if (this.form.codeSecretId) {
        let params = {
          git_type: this.form.codeType,
          secret_id: this.form.codeSecretId,
          api_url: this.form.apiUrl,
        }
        this.codeRepoLoading = true
        listGitRepos(params).then((res) => {
          this.codeRepos = res.data || []
          this.codeRepoLoading = false
        }).catch((err) => {
          console.log(err)
          this.codeRepoLoading = false
        });
      }
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
