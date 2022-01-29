<template>
  <div>
    <clusterbar :titleName="titleName" :editFunc="getSecretYaml" />
    <div class="dashboard-container">
      <div style="padding: 10px 8px 0px;">
        <div>基本信息</div>
        <el-form label-position="left" class="pod-item" v-if="secret.metadata" style="margin: 15px 10px 20px 10px; width: 100%">
          <el-form-item label="名称">
            <span>{{ secret.metadata.name }}</span>
          </el-form-item>
          <el-form-item label="创建时间">
            <span>{{ secret.metadata.creationTimestamp }}</span>
          </el-form-item>
          <el-form-item label="命名空间">
            <span>{{ secret.metadata.namespace }}</span>
          </el-form-item>
          <el-form-item label="类型">
            <span>{{ secret.type }}</span>
          </el-form-item>

          <el-form-item label="标签">
            <span v-if="!secret.metadata.labels">—</span>
            <template v-else v-for="(key, val) in secret.metadata.labels" >
              <span :key="key" class="back-class">{{key}}:{{val}}<br/></span>
            </template>
          </el-form-item>
        </el-form>
      </div>

      <div style="padding: 10px 8px 0px;">
        <div>数据</div>
        <div class="msgClass" style="margin-top: 15px;">
          <el-table
            v-if="secretData"
            :data="secretData"
            class="table-fix"
            tooltip-effect="dark"
            style="width: 100%"
            :cell-style="cellStyle"
            :default-sort = "{prop: 'lastProbeTime'}"
            >
            <el-table-column
              prop="key"
              label="键"
              min-width="30"
              show-overflow-tooltip>
              <template slot-scope={row}>
                <el-input placeholder="" v-model="row.key"></el-input>
              </template>
            </el-table-column>
            <el-table-column
              prop="value"
              label="值"
              min-width="90"
              show-overflow-tooltip>
              <template slot-scope={row}>
                <el-input type="textarea" placeholder="" v-model="row.value"></el-input>
              </template>
            </el-table-column>
            <el-table-column label="操作" min-width="20" align="center">
              <template slot-scope="scope">

                <el-button type="primary" size="mini" 
                  @click="getSecret(scope.row)">
                  {{ scope.row.decode ? '隐藏' : '显示' }}
                </el-button>
                <!-- <el-button size="mini" type="danger"
                  @click="handleDelete(scope.$index, scope.row)">删除</el-button> -->
              </template>
            </el-table-column>
          </el-table>
          <div v-else style="padding: 25px 15px ; color: #909399; text-align: center">暂无数据</div>
        </div>
      </div>

      <!-- <el-collapse v-model="activeNames" @change="handleChange">
        <el-collapse-item title="Data" name="1" v-if="secret.data"> -->
          <!-- <div v-for="(key, val) in secret.data" :key="val" class="msgClass" >
            <div class="dataDiv">
              <div>{{val}}
                <el-button type="text" icon="el-icon-view" circle size="mini" @click.once="getSecretDisplay(key,val)"></el-button>
              </div>
            </div>
            <el-input type="textarea"  :autosize='{ minRows: 2, maxRows: 4 }' readonly v-model="secret.data[val]">
              <i slot="suffix" class="el-input__icon el-icon-date"></i>
            </el-input>
          </div> -->
        <!-- </el-collapse-item>

      </el-collapse> -->


      <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
        <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
        <span slot="footer" class="dialog-footer">
          <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
          <el-button plain size="small">确 定</el-button>
        </span>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { Clusterbar, Yaml } from '@/views/components'
import { getSecret } from '@/api/secret'
import { Message } from 'element-ui'

let Base64 = require('js-base64').Base64

export default {
  name: 'SecretDetail',
  components: {
    Clusterbar,
    Yaml,
  },
  data() {
    return {
      yamlDialog: false,
      yamlValue: '',
      yamlLoading: true,
      cellStyle: { border: 0 },
      loading: true,
      originSecret: {},
      eventLoading: true,
      activeNames: ["1"]
    }
  },
  created() {
    this.fetchData()
  },
  watch: {},
  computed: {
    titleName: function() {
      return ['Secret', this.secretName]
    },
    secretName: function() {
      return this.$route.params ? this.$route.params.secretName : ''
    },
    cluster: function() {
      return this.$store.state.cluster
    },
    secret: function() {
      return this.originSecret
    },
    namespace: function() {
      return this.$route.params ? this.$route.params.namespace : ""
    },
    secretData: function() {
      if (!this.originSecret.data) return []
      let d = this.originSecret.data
      let dataTable = []
      Object.keys(d).forEach(key => {
      dataTable.push({
          key: key,
          value: d[key],
          decode: false
        })
      })
      return dataTable
    },
  },
  methods: {
    getSecret(row) {
      if(row.decode) {
        var res = Base64.encode(row.value)
      } else {
        var res = Base64.decode(row.value)
      }
      row['value'] = res
      row['decode'] = !row['decode']
    },
    handleChange(val) {
        console.log(val);
    },
    fetchData: function() {
      this.originSecret = {}
      this.loading = true
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error('获取集群参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.secretName) {
        Message.error('获取Secret名称参数异常，请刷新重试')
        this.loading = false
        this.eventLoading = false
        return
      }
      if (!this.namespace) {
        Message.error('获取获取Secret命名空间参数异常，请刷新重试')
      }
      getSecret(cluster, this.namespace, this.secretName).then(response => {
        this.loading = false
        this.originSecret = response.data
      }).catch(() => {
        this.loading = false
      })
    },
    getSecretYaml: function() {
      if (!this.secretName) {
        Message.error('获取Secret参数异常，请刷新重试')
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error('获取集群参数异常，请刷新重试')
        return
      }
      this.yamlValue = ''
      this.yamlDialog = true
      this.yamlLoading = true
      getSecret(cluster, this.namespace, this.secretName, 'yaml')
        .then((response) => {
          this.yamlLoading = false
          this.yamlValue = response.data
        })
        .catch(() => {
          this.yamlLoading = false
        })
    },
    updateSecret: function() {
      if (!this.Secret) {
        Message.error("获取Secret参数异常，请刷新重试")
        return
      }
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      console.log(this.yamlValue)
      console.log(this.secret)
    },
  },
}
</script>

<style lang="scss" scoped>
  .my-table >>> .el-table__row>td{
  /* 去除表格线 */
  border: none;
}

.my-table >>> .el-table::before {
	height: 0px;
}

.dashboard {
  &-container {
    margin: 10px 30px;
  }
  &-text {
    font-size: 30px;
    line-height: 46px;
  }

  .table-fix {
    height: calc(100% - 100px);
  }
}
.name-class {
  cursor: pointer;
}
.name-class:hover {
  color: #409eff;
}
</style>

<style>
/* .el-table__expand-icon {
  display: none;
} */
.el-table__expanded-cell[class*='cell'] {
  padding-top: 5px;
}
.table-expand {
  font-size: 0;
}
.table-expand label {
  width: 90px;
  color: #99a9bf;
  font-weight: 400;
}
.table-expand .el-form-item {
  margin-right: 0;
  margin-bottom: 0;
  width: 100%;
}

.pod-item {
  margin: 20px 5px 30px 5px;
  padding: 10px 20px;
  font-size: 0;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}
.pod-item label {
  width: 90px;
  color: #99a9bf;
  font-weight: 400;
}
.pod-item .el-form-item {
  margin-right: 0;
  margin-bottom: 0;
  width: 60%;
}
.pod-item span {
  color: #606266;
}
/* .el-collapse {
  border-top: 0px;
} */
.title-class {
  margin-left: 5px;
  color: #606266;
  font-size: 13px;
}
.podCollapse .el-collapse-item__content {
  padding: 0px 10px 15px;
  /* font-size: 13px; */
}
.el-dialog__body {
  padding-top: 5px;
}
.msgClass {
  margin: 0px 25px;
}
.msgClass .el-table::before {
  height: 0px;
}
.msgClass {
  margin: 8px 10px 15px 10px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}
.dataDiv {
  padding: 5px;
}
</style>
