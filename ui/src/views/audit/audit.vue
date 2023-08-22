<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" />
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="audits"
        class="table-fix"
        :cell-style="cellStyle"
        v-loading="loading"
        :max-height="maxHeight"
        :default-sort = "{prop: 'name'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column v-if="scope==''" label="所属范围" show-overflow-tooltip min-width="12">
          <template slot-scope="scope">
            {{ scopeDisplay(scope.row) }}
          </template>
        </el-table-column>
        <el-table-column prop="operation" label="操作类型" show-overflow-tooltip min-width="8">
        </el-table-column>
        <el-table-column prop="resource_type" label="资源类型" show-overflow-tooltip min-width="10">
        </el-table-column>
        <el-table-column prop="resource_name" label="资源名称" show-overflow-tooltip min-width="14">
        </el-table-column>
        <el-table-column v-if="scope!='pipeline'" prop="namespace" label="命名空间" show-overflow-tooltip min-width="10">
        </el-table-column>
        <el-table-column prop="operate_detail" label="操作详情" show-overflow-tooltip min-width="25">
        </el-table-column>
        <el-table-column prop="code" label="操作结果" show-overflow-tooltip min-width="10">
          <template slot-scope="scope">
            {{ scope.row.code == 'Success' ? '成功': '失败：' + scope.row.message }}
          </template>
        </el-table-column>
        <el-table-column prop="update_user" label="操作人" show-overflow-tooltip min-width="8">
          <template slot-scope="scope">
            {{ scope.row.operator }}
          </template>
        </el-table-column>
        <el-table-column prop="update_time" label="操作时间" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.create_time) }}
          </template>
        </el-table-column>
      </el-table>
      <div style="float: right">
        <el-pagination
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          :hide-on-single-page="singlePage"
          :page-sizes="[20, 50, 80, 120]"
          :page-size="20"
          layout="total, sizes, prev, pager, next"
          :total="totalRecords">
        </el-pagination>
      </div>

    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { listAuditOperate } from "@/api/audit/audit";
import { Message } from "element-ui";

export default {
  name: "audit",
  components: {
    Clusterbar,
  },
  props: ['scope', 'scopeId'],
  mounted: function () {
    const that = this;
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - this.$contentHeight;
        that.maxHeight = heightStyle;
      })();
    };
  },
  data() {
    return {
      maxHeight: window.innerHeight - this.$contentHeight,
      cellStyle: { border: 0 },
      titleName: ["操作审计"],
      loading: true,
      audits: [],
      search_name: "",
      scopeNameMap: {
        cluster: "集群",
        project: "工作空间",
        pipeline: "流水线",
        platform: "平台配置",
        appstore: "应用商店"
      },
      singlePage: true,
      totalRecords: 0,
      page: {
        size: 20,
        pageNo: 1,
      }
    };
  },
  created() {
    this.fetchAuditOperates();
  },
  computed: {
  },
  methods: {
    fetchAuditOperates() {
      this.loading = true
      let query = {
        scope: this.scope,
        scope_id: this.scopeId,
        fuzzy_name: this.search_name,
        page_size: this.page.size,
        page_no: this.page.pageNo,
      }
      console.log(query)
      listAuditOperate(query).then((resp) => {
        if(resp.data.pagination && resp.data.pagination.pages > 1) {
          this.singlePage = false
          this.maxHeight = window.innerHeight - this.$contentHeight - 28
        }
        if(resp.data.pagination) {
          this.totalRecords = resp.data.pagination.records
        }
        this.audits = resp.data.data ? resp.data.data : []
        // console.log(this.audits)
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    nameSearch(val) {
      this.search_name = val;
      this.page.pageNo = 1
      this.fetchAuditOperates()
    },
    scopeDisplay(row) {
      let s = this.scopeNameMap[row.scope]
      if(row.scope_name) {
        s += "/" + row.scope_name
      }
      return s
    },
    handleSizeChange(size) {
      this.page.size = size
      this.fetchAuditOperates()
    },
    handleCurrentChange(val) {
      this.page.pageNo = val
      this.fetchAuditOperates()
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
