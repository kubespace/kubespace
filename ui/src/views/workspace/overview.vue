<template>
  <div>
    <div class="dashboard-container" ref="tableCot">
    </div>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { listWorkspaces } from '@/api/pipeline/workspace'
import { getUser } from "@/api/user";
import { Message } from 'element-ui'

export default {
  name: 'ProjectOverview',
  components: {
    Clusterbar
  },
  data() {
    return {
      titleName: ["项目空间"],
      search_name: '',
      users: [],
      cellStyle: {border: 0},
      maxHeight: window.innerHeight - 150,
      loading: true,
      workspaces: [],
      createClusterFormVisible: false,
      inviteForm: false,
      clusterConnectDialog: false,
      clusterConnectToken: '',
      form: {
        name: '',
        members: [],
      },
      locationAddr: window.location.origin,
    }
  },
  created() {
    this.fetchData();
    this.fetchUsers();
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
  },
  methods: {
    fetchData() {
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
    fetchUsers() {
      getUser({}).then((response) => {
        this.users = response.data;
      });
    },
    nameClick: function(id) {
      parent.location.href = '/ui/pipespace/' + id
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
    deleteClusters(delClusters) {
      if(delClusters.length <= 0) {
        Message.error('请选择要删除的集群')
        return
      }
      deleteCluster(delClusters).then((response) => {
          this.fetchData()
      }).catch((e) => {
        console.log(e)
      })
    },
    onCopy(e) {
      Message.success("复制成功")
    },
    onError(e) {

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
