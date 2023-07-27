<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="openCreateDialog" createDisplay="添加节点"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="originSpacelets"
        class="table-fix"
        :cell-style="cellStyle"
        v-loading="loading"
        :default-sort = "{prop: 'name'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column prop="hostname" label="主机名" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="roles" label="访问地址" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            http://{{ scope.row.hostip }}:{{ scope.row.port }}
          </template>
        </el-table-column>
        <el-table-column prop="update_user" label="状态" show-overflow-tooltip min-width="10">
          <template slot-scope="scope">
            {{ statusMap[scope.row.status] }}
          </template>
        </el-table-column>
        <el-table-column prop="update_time" label="更新时间" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.update_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template slot-scope="scope">
            <div class="tableOperate">
              <!-- <el-link :disabled="!$editorRole()" :underline="false" type="primary" style="margin-right: 15px;"  @click="updateSecretFormDialog(scope.row)">编辑</el-link> -->
              <el-link v-if="scope.row.status=='offline'" :disabled="!$editorRole()" :underline="false" type="danger" 
                @click="handleDeleteSpacelet(scope.row.id, scope.row.hostname)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-dialog title="添加Spacelet节点" :visible.sync="createDialogVisible" width="60%" @close="createDialogVisible=false;">
        <div style="margin-bottom: 15px;">
          <div>
            <span>Spacelet节点是用来执行流水线构建任务的， 如代码编译、发布等。通过添加Spacelet节点可以降低每个节点的负载，同时能够并发处理更多的构建任务。</span>
          </div>
          <div style="">
            <p>有以下两种方式添加Spacelet节点：</p>
          </div>
          <el-tabs type="border-card">
            <el-tab-pane label="Kubernetes集群节点">
              <div>登陆到KubeSpace所在集群，执行如下命令：</div>
              <p style="background-color: #fafafa; padding: 15px;" class="add-spacelt-p">
                # helm upgrade --set spacelet.replicaCount=&lt;num&gt; -n kubespace kubespace kubespace/kubespace
              </p>
              <p>
                如上，通过升级KubeSpace，并设置 spacelet.replicaCount 参数为&lt;num&gt;，该值默认为1，&lt;num&gt;值不能超过当前集群的节点数。
              </p>
              <p>
                同时，也可以通过 spacelet.nodeSelector 设置节点调度策略，将spacelet调度到相应节点上。
              </p>
            </el-tab-pane>
            <el-tab-pane label="物理节点">
              <div>Spacelet也可以单独部署在物理服务器上，登陆到机器，执行如下命令：</div>
              <p style="background-color: #fafafa; padding: 15px;" class="add-spacelt-p">
                # curl -sk {{ locationAddr }}/api/v1/spacelet/install.sh | sh -x
              </p>
              <p>
                *注意：请将上述访问地址「{{this.locationAddr}}」更改为机器可以访问的地址。
              </p>
            </el-tab-pane>
          </el-tabs>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { listSpacelet, deleteSpacelet } from "@/api/spacelet/spacelet";
import { Message } from "element-ui";

export default {
  name: "spacelet",
  components: {
    Clusterbar,
  },
  data() {
    return {
      maxHeight: window.innerHeight - this.$contentHeight,
      cellStyle: { border: 0 },
      titleName: ["Spacelet管理"],
      loading: true,
      originSpacelets: [],
      search_name: "",
      statusMap: {
        online: "在线",
        offline: "已下线"
      },
      createDialogVisible: false,
      locationAddr: window.location.origin,
    };
  },
  created() {
    this.fetchSpacelets();
  },
  mounted: function () {
    const that = this;
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - this.$contentHeight;
        that.maxHeight = heightStyle;
      })();
    };
  },
  computed: {
  },
  methods: {
    fetchSpacelets() {
      this.loading = true
      listSpacelet().then((resp) => {
        this.originSpacelets = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    nameSearch(val) {
      this.search_name = val;
    },
    openCreateDialog() {
      this.createDialogVisible = true;
    },
    handleDeleteSpacelet(id, name) {
      if(!id) {
        Message.error("获取id参数异常，请刷新重试");
        return
      }
      this.$confirm(`请确认是否删除「${name}」此Spacelet节点?`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.loading = true
        deleteSpacelet(id).then(() => {
          this.loading = false
          Message.success("删除Spacelet节点成功")
          this.fetchSpacelets()
        }).catch((err) => {
          console.log(err)
          this.loading = false
        });
      }).catch(() => {       
      });
    }
  },
};
</script>


<style lang="scss" scoped>
@import "~@/styles/variables.scss";
.add-spacelt-p {
  font-family: ui-monospace,SFMono-Regular,SF Mono,Menlo,Consolas,Liberation Mono,monospace
}

</style>
