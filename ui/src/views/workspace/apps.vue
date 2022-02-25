<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="openCreateFormDialog" createDisplay="创建应用"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="originApps"
        class="table-fix"
        :cell-style="cellStyle"
        v-loading="loading"
        :default-sort = "{prop: 'name'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column prop="name" label="名称" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.id)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="package_version" label="版本" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="type" label="类型" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ typeNameMap[scope.row.type] }}
          </template>
        </el-table-column>
        <el-table-column prop="update_user" label="操作人" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="update_time" label="更新时间" show-overflow-tooltip min-width="20">
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.update_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            <div class="status-class" :style="{'border-color': statusColorMap[scope.row.status], 'background-color': statusColorMap[scope.row.status]}"></div>
            <span :style="{'font-weight': 430}">{{ statusNameMap[scope.row.status] }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :underline="false" class="operator-btn"
                v-if="scope.row.status=='UnInstall'" @click="nameClick(scope.row.id)">安装</el-link>
              <el-link :underline="false" class="operator-btn"
                v-if="scope.row.status=='Running'" @click="nameClick(scope.row.id)">升级</el-link>
              <el-link :underline="false" class="operator-btn"
                @click="openUpdateFormDialog(scope.row)">编辑</el-link>
              <el-link :underline="false" class="operator-btn" style="color: #F56C6C"
                v-if="scope.row.status=='Running'" @click="nameClick(scope.row.id)">销毁</el-link>
              <el-link :underline="false" style="color: #F56C6C" v-if="scope.row.status=='UnInstall'"
                @click="handleDeleteApp(scope.row.id, scope.row.name)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="updateFormVisible ? '修改空间' : '创建空间'" :visible.sync="createFormVisible"
      @close="closeFormDialog" :destroy-on-close="true">
        <div class="dialogContent" style="">
          <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
            <el-form-item label="名称" prop="name" autofocus>
              <el-input v-model="form.name" autocomplete="off" placeholder="请输入空间名称" size="small"></el-input>
            </el-form-item>
            <el-form-item label="描述" prop="description">
              <el-input v-model="form.description" type="textarea" autocomplete="off" placeholder="请输入空间描述" size="small"></el-input>
            </el-form-item>
            <el-form-item label="命名空间" prop="" :required="true">
              <el-select v-model="form.namespace" placeholder="请选择要绑定的命名空间" size="small" style="width: 100%;"
                :disabled="updateFormVisible">
                <el-option
                  v-for="item in namespaces"
                  :key="item.name"
                  :label="item.name"
                  :value="item.name">
                </el-option>
              </el-select>
            </el-form-item>
            <el-form-item label="负责人" prop="" :required="true">
              <el-input v-model="form.owner" autocomplete="off" placeholder="请输入该项目空间负责人" size="small"></el-input>
            </el-form-item>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter">
          <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="updateFormVisible ? handleUpdateApp() : handleCreateApp()" >确 定</el-button>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { listApps } from "@/api/project/apps";
import { Message } from "element-ui";

export default {
  name: "projectApps",
  components: {
    Clusterbar,
  },
  mounted: function () {
    const that = this;
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 150;
        that.maxHeight = heightStyle;
      })();
    };
  },
  data() {
    return {
      maxHeight: window.innerHeight - 150,
      cellStyle: { border: 0 },
      titleName: ["应用管理"],
      loading: true,
      createFormVisible: false,
      updateFormVisible: false,
      clusters: [],
      namespaces: [],
      form: {
        id: "",
        name: "",
        description: "",
        cluster_id: "",
        namespace: "",
        owner: "",
      },
      rules: {
        name: [{ required: true, message: '请输入空间名称', trigger: 'blur' },],
      },
      statusNameMap: {
        "UnInstall": "未安装",
        "UnReady": "未就绪",
        "RunningFault": "运行故障",
        "Running": "运行中"
      },
      statusColorMap: {
        "UnInstall": "",
        "UnReady": "#E6A23C",
        "RunningFault": "#F56C6C",
        "Running": "#67C23A"
      },
      typeNameMap: {
        "ordinary_app": "普通应用",
        "middleware": "中间件",
        "import_app": "导入应用"
      },
      originApps: [],
      search_name: "",
    };
  },
  created() {
    this.fetchApps();
  },
  computed: {
    projectId() {
      return this.$route.params.workspaceId
    }
  },
  methods: {
    nameClick: function(id) {
      this.$router.push({name: 'workspaceOverview', params: {'workspaceId': id}})
    },
    handleEdit(index, row) {
      console.log(index, row);
    },
    fetchApps() {
      this.loading = true
      listApps({project_id: this.projectId}).then((resp) => {
        this.originApps = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    getAppStatusName(status) {

    },
    handleCreateApp() {
      if(!this.form.name) {
        Message.error("空间名称不能为空");
        return
      }
      if(!this.form.cluster_id) {
        Message.error("请选择要绑定的集群");
        return
      }
      if(!this.form.namespace) {
        Message.error("请选择要绑定的集群命名空间");
        return
      }
      let project = {
        name: this.form.name, 
        description: this.form.description, 
        cluster_id: this.form.cluster_id,
        namespace: this.form.namespace,
        owner: this.form.owner
      }
      createProject(project).then(() => {
        this.createFormVisible = false;
        Message.success("创建项目空间成功")
        this.fetchApps()
      }).catch((err) => {
        console.log(err)
      });
    },
    handleUpdateApp() {
      if(!this.form.id) {
        Message.error("获取空间id参数异常，请刷新重试");
        return
      }
      let project = {
        name: this.form.name, 
        description: this.form.description, 
        owner: this.form.owner
      }
      updateSecret(this.form.id, project).then(() => {
        this.createFormVisible = false;
        Message.success("更新项目空间成功")
        this.fetchApps()
      }).catch((err) => {
        console.log(err)
      });
    },
    handleDeleteApp(id, name) {
      if(!id) {
        Message.error("获取密钥id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${name}」此项目空间?`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          deleteSecret(id).then(() => {
            Message.success("删除项目空间成功")
            this.fetchApps()
          }).catch((err) => {
            console.log(err)
          });
        }).catch(() => {       
        });
    },
    nameSearch(val) {
      this.search_name = val;
    },
    openCreateFormDialog() {
      this.createFormVisible = true;
      this.fetchClusters()
    },
    openUpdateFormDialog(project) {
      this.form = {
        id: project.id,
        name: project.name,
        description: project.description,
        cluster_id: project.cluster_id,
        namespace: project.namespace,
        owner: project.owner,
      }
      this.updateFormVisible = true;
      this.createFormVisible = true;
      // this.fetchClusters()
    },
    closeFormDialog() {
      this.form = {
        name: "",
        description: "",
        cluster_id: "",
        namespace: "",
        owner: "",
      }
      this.updateFormVisible = false; 
      this.createFormVisible = false;
    }
  },
};
</script>


<style lang="scss" scoped>
@import "~@/styles/variables.scss";

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
