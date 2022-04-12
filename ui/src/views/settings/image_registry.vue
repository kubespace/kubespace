<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="openCreateFormDialog" createDisplay="添加仓库"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="originImageRegistry"
        class="table-fix"
        :cell-style="cellStyle"
        v-loading="loading"
        :default-sort = "{prop: 'name'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column prop="registry" label="仓库地址" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="user" label="认证用户" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="update_user" label="操作人" show-overflow-tooltip min-width="10">
        </el-table-column>
        <el-table-column prop="update_time" label="更新时间" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.update_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template slot-scope="scope">
            <div class="tableOperate">
              <!-- <el-link :underline="false" style="margin-right: 13px; color:#409EFF" @click="nameClick(scope.row.id)">详情</el-link> -->
              <el-link :underline="false" style="margin-right: 13px; color:#409EFF" @click="openUpdateFormDialog(scope.row)">编辑</el-link>
              <el-link :underline="false" style="color: #F56C6C" @click="handleDeleteImageRegistry(scope.row.id, scope.row.registry)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="updateFormVisible ? '修改仓库' : '添加仓库'" :visible.sync="createFormVisible"
      @close="closeFormDialog" :destroy-on-close="true">
        <div class="dialogContent" style="">
          <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
            <el-form-item label="仓库地址" prop="registry" autofocus>
              <el-input v-model="form.registry" autocomplete="off" placeholder="请输入镜像仓库地址，如: docker.io" size="small"></el-input>
            </el-form-item>
            <el-form-item label="用户" prop="user">
              <el-input v-model="form.user" autocomplete="off" placeholder="请输入认证用户" size="small"></el-input>
            </el-form-item>
            <el-form-item label="密码" prop="password" :required="true">
              <el-input v-model="form.password" type="password" autocomplete="off" placeholder="请输入认证密码" size="small"></el-input>
            </el-form-item>
          </el-form>
        </div>
        <div slot="footer" class="dialogFooter">
          <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
          <el-button type="primary" @click="updateFormVisible ? handleUpdateImageRegistry() : handleCreateImageRegistry()" >确 定</el-button>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { createImageRegistry, listImageRegistry, updateImageRegistry, deleteImageRegistry } from "@/api/settings/image_registry";
import { Message } from "element-ui";

export default {
  name: "ImageRegistry",
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
      titleName: ["镜像仓库"],
      loading: true,
      createFormVisible: false,
      updateFormVisible: false,
      form: {
        id: "",
        registry: "",
        user: "",
        password: "",
      },
      rules: {
        registry: [{ required: true, message: '请输入镜像仓库地址', trigger: 'blur' },],
        user: [{ required: true, message: '请输入镜像仓库认证用户', trigger: 'blur' },],
        password: [{ required: true, message: '请输入镜像仓库认证密码', trigger: 'blur' },],
      },
      originImageRegistry: [],
      search_name: "",
    };
  },
  created() {
    this.fetchImageRegistry();
  },
  computed: {
  },
  methods: {
    fetchImageRegistry() {
      this.loading = true
      listImageRegistry().then((resp) => {
        this.originImageRegistry = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    handleCreateImageRegistry() {
      if(!this.form.registry) {
        Message.error("镜像仓库地址不能为空");
        return
      }
      if(!this.form.user) {
        Message.error("请输入镜像仓库认证用户");
        return
      }
      if(!this.form.password) {
        Message.error("请输入镜像仓库认证密码");
        return
      }
      let image_registry = {
        registry: this.form.registry, 
        user: this.form.user, 
        password: this.form.password,
      }
      createImageRegistry(image_registry).then(() => {
        this.createFormVisible = false;
        Message.success("创建镜像仓库成功")
        this.fetchImageRegistry()
      }).catch((err) => {
        console.log(err)
      });
    },
    handleUpdateImageRegistry() {
      if(!this.form.id) {
        Message.error("获取镜像仓库id参数异常，请刷新重试");
        return
      }
      let image_registry = {
        user: this.form.user, 
        password: this.form.password, 
      }
      updateImageRegistry(this.form.id, image_registry).then(() => {
        this.createFormVisible = false;
        Message.success("更新镜像仓库成功")
        this.fetchImageRegistry()
      }).catch((err) => {
        console.log(err)
      });
    },
    handleDeleteImageRegistry(id, registry) {
      if(!id) {
        Message.error("获取镜像仓库id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${registry}」镜像仓库?`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          deleteImageRegistry(id).then(() => {
            Message.success("删除镜像仓库成功")
            this.fetchImageRegistry()
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
    },
    openUpdateFormDialog(obj) {
      this.form = {
        id: obj.id,
        registry: obj.registry,
        user: obj.user,
        password: ''
      }
      this.updateFormVisible = true;
      this.createFormVisible = true;
      // this.fetchClusters()
    },
    closeFormDialog() {
      this.form = {
        id: "",
        registry: "",
        user: "",
        password: "",
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

</style>
