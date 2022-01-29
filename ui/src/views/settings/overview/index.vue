<template>
  <div>
    <router-view />
    <h3>全局总览</h3>
    <el-row :gutter="12" class="card">
      <el-col :span="8" class="card-item">
        <el-card shadow="hover">
          <div slot="header" class="clearfix">
            <span>集群数量</span>
            <el-button style="float: right; padding: 3px 0" type="text">查看详情</el-button>
          </div>
          <div class="text">24</div>
        </el-card>  
      </el-col>
      <el-col :span="8" class="card-item">
        <el-card shadow="hover">
          <div slot="header" class="clearfix">
            <span>成员数量</span>
            <el-button style="float: right; padding: 3px 0" type="text">查看详情</el-button>
          </div>
          <div class="text">24</div>
        </el-card>
      </el-col>
      <el-col :span="8" class="card-item">
        <el-card shadow="hover">
          <div slot="header" class="clearfix">
            <span>告警信息</span>
            <el-button style="float: right; padding: 3px 0" type="text">查看详情</el-button>
          </div>
          <div class="text">24</div>
        </el-card>
      </el-col>
    </el-row>

    <el-divider></el-divider>


  </div>
</template>

<style lang="scss" scoped>
.card {
  display: flex;
  margin: 10px;
  flex-direction: row;
  justify-content: space-around;
}
.card-item {
  align-content: center;
  margin: 20px;
}
.text {
    text-align: center;
    font-size: xx-large;
    flex-direction: row-reverse;
    color: #67C23A;
}
</style>