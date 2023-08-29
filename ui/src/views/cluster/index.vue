<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="createClusterDialog" createDisplay="添加集群"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        :data="clusters"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort = "{prop: 'name'}"
        row-key="name"
      >
        <el-table-column prop="name" label="名称" show-overflow-tooltip>
          <template slot-scope="scope">
            <span v-if="scope.row.status === 'Connect'" class="name-class" v-on:click="nameClick(scope.row.name)">
              {{ scope.row.name1 }}
            </span>
            <span v-else>
              {{ scope.row.name1 }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="集群版本" show-overflow-tooltip>
          <template slot-scope="scope">
            {{ scope.row.version }}
          </template>
        </el-table-column>
        <el-table-column prop="created_by" label="创建人" show-overflow-tooltip>
          <template slot-scope="scope">
            {{ scope.row.created_by }}
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" show-overflow-tooltip>
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.create_time) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          label="状态"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            <span :style="{'color': (scope.row.status === 'Connect' ? '#67c23a' : '#F56C6C')}">
              
              <template v-if="scope.row.status === 'Connect'">
                <i class="el-icon-success" style="font-size: 16px;" ></i> 
                <span class="correct">连接成功</span>
                <!-- <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="correct" /> -->
              </template>
              <template v-else-if="scope.row.status === 'Pending'">
                <i class="el-icon-warning" style="font-size: 16px; color: #E6A23C" ></i> 
                <span class="wrong" style="color: #E6A23C">未连接</span>
                <!-- <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="wrong" /> -->
              </template>
              <template v-else-if="scope.row.status === 'Failed'">
                <i class="el-icon-error" style="font-size: 16px;" ></i>
                <span class="wrong">连接失败：{{ scope.row.connect_error }}</span>
                <!-- <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="wrong" /> -->
              </template>
            </span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="150">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 15px;" 
                @click="cluster=scope.row; clusterConnectToken=scope.row.token; clusterConnectDialog = true" 
                v-if="scope.row.status !== 'Connect'">
                导入集群
              </el-link>
              <el-link :underline="false" type='primary' style="margin-right: 15px;" 
                @click="nameClick(scope.row.id)"
                v-if="scope.row.status === 'Connect'">集群详情</el-link>
              <el-link :disabled="!$adminRole()" :underline="false" type="danger" @click="deleteCluster({id: scope.row.id, name1: scope.row.name1})">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <div>
      <el-dialog :title="inviteForm ? '邀请用户':'创建集群'" :visible.sync="createClusterFormVisible" :close-on-click-modal="false" 
        :destroy-on-close="true" @close="form={'name': '', 'members': []}; inviteForm=false;">
        <el-form :model="form">
          <el-form-item label="集群名称">
            <el-input v-model="form.name" :disabled="inviteForm" autocomplete="off" placeholder="请输入集群名称"></el-input>
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button @click="createClusterFormVisible = false; form={'name': '', 'members': []}; inviteForm=false;">取 消</el-button>
          <el-button type="primary" @click="inviteForm ? handleClusterMembers() : handleCreateCluster();">确 定</el-button>
        </div>
      </el-dialog>
    </div>

    <el-dialog title="集群导入" :visible.sync="clusterConnectDialog" :close-on-click-modal="false">
      
      <el-tabs v-model="activeName" type="border-card" >
        <el-tab-pane label="KubeConfig" name="KubeConfig">
          <yaml v-model="kubeconfig" :loading="yamlLoading" :height="200"></yaml>
        </el-tab-pane>
        <el-tab-pane label="Agent" name="Agent" style="padding: 0px 0px;">
          <div style="font-size: 15px;">请在Kubernetes集群master节点运行下面的kubeclt命令，以连接KubeSpace平台：</div>
            <div style="margin-top: 15px;">
              <el-tag type="info" style="font-size: 14px; border-radius: 4px 0px 0px 4px;  border-right: 0px">
                {{ copyCluster }}
              </el-tag>
              <el-button plain size="small" slot="append" 
                  style="height: 32px; border-radius: 0px 4px 4px 0px; padding: 10px 8px;"
                  v-clipboard:copy="copyCluster" v-clipboard:success="onCopy" v-clipboard:error="onError">
                  复制
              </el-button>
            </div>
            <div style="font-size: 13px; margin-top: 8px; color: #e6a23c;">
              *注意：请将上述访问地址「{{this.locationAddr}}」更改为Kubernetes集群可以访问的地址。
            </div>
        </el-tab-pane>
      </el-tabs>

      
      <div slot="footer" class="dialogFooter" style="text-align: right">
        <el-button type="primary" @click="clusterConnect">确 定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { listCluster, createCluster, updateCluster, deleteCluster, clusterMembers } from '@/api/cluster'
import { getUser } from "@/api/user";
import { Message } from 'element-ui'

export default {
  name: 'SettingsCluster',
  components: {
    Clusterbar,
    Yaml,
  },
  data() {
    return {
      titleName: ["集群管理"],
      search_name: '',
      cellStyle: {border: 0},
      maxHeight: window.innerHeight - this.$contentHeight,
      loading: true,
      clusters: [],
      createClusterFormVisible: false,
      inviteForm: false,
      clusterConnectDialog: false,
      clusterConnectToken: '',
      form: {
        name: '',
        members: [],
      },
      locationAddr: window.location.origin,
      kubeconfig: "",
      yamlLoading: false,
      activeName: "KubeConfig",
      cluster: {}
    }
  },
  created() {
    this.fetchData();
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - this.$contentHeight
        console.log(heightStyle)
        that.maxHeight = heightStyle
      })()
    }
  },
  watch: {
  },
  computed: {
    copyCluster() {
      return `curl -sk ${ this.locationAddr }/cluster/agent/import/${ this.clusterConnectToken } | kubectl apply -f -`;
    },
  },
  methods: {
    fetchData() {
      this.loading = true
      listCluster()
        .then((response) => {
          this.loading = false
          this.clusters = response.data || [];
        })
        .catch(() => {
          this.loading = false
        })
    },
    nameClick: function(name) {
      this.$router.push({name: 'cluster', params: {clusterId: name}})
      // parent.location.href = '/ui/cluster/' + name
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    createClusterDialog() {
      this.createClusterFormVisible = true;
    },
    handleCreateCluster() {
      if (!this.form.name) {
        Message.error('集群名称不能为空！')
        return
      }
      createCluster(this.form)
        .then((response) => {
          Message.success("集群添加成功")
          this.createClusterFormVisible = false
          this.loading = false
          this.fetchData()
          this.clusterConnectToken = response.data.token
          this.clusterConnectDialog = true;
        })
        .catch(() => {
          // this.createClusterFormVisible = false
          this.loading = false
        })
    },
    handleClusterMembers() {
      if (!this.form.name) {
        Message.error('集群名称不能为空！')
        return
      }
      clusterMembers(this.form)
        .then((response) => {
          Message.success("邀请用户成功")
          this.createClusterFormVisible = false
          this.inviteForm = false
          this.loading = false
          this.fetchData()
        })
        .catch(() => {
          // this.createClusterFormVisible = false
          this.loading = false
        })
    },
    deleteCluster(cluster) {
      if(!cluster || !cluster.id) {
        Message.error('请选择要删除的集群')
        return
      }
      this.$confirm(`请确认是否删除「${cluster.name1}」集群?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        deleteCluster(cluster.id).then((response) => {
          this.fetchData()
        }).catch((e) => {
          console.log(e)
        })
      }).catch(() => {       
      });
    },
    onCopy(e) {
      Message.success("复制成功")
    },
    onError(e) {

    },
    clusterConnect() {
      if(this.kubeconfig!="") {
        updateCluster(this.cluster.id, {"kubeconfig": this.kubeconfig}).then((response) => {
            this.fetchData()
        }).catch((e) => {
          console.log(e)
        })
      }
      this.kubeconfig = ""
      this.clusterConnectDialog = false
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
// .dashboard {
//   &-container {
//     margin: 10px 30px;
//     height: calc(100%);
//   }
//   &-text {
//     font-size: 30px;
//     line-height: 46px;
//   }
// }

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
