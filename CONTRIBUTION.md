# CONTRIBUTION

### 一、开发流程

#### 1. Fork

访问https://github.com/kubespace/kubespace，点击`Fork`在自己的账号仓库创建一份项目代码。

#### 2. 克隆代码到本地

```bash
export user={你自己的github仓库账号}
git clone https://github.com/$user/kubespace
cd kubespace
git remote add upstream https://github.com/kubespace/kubespace

git remote set-url --push upstream no_push
git remote -v
```

#### 3. 与远程代码库同步

```bash
git fetch upstream
git checkout master
git rebase upstream/master
```

#### 4. 本地开发

```bash
# 本地创建分支，并在该分支进行开发
git checkout -b feature
```

#### 5. 推送代码

```bash
# 先同步下远程代码，以免有冲突
git checkout master
git fetch upstream
git rebase upstream/master

git checkout feature
git rebase -i master

# 提交本地修改
git add <file>
git commit -s -m "add code modify message"

# 推送到Fork远程仓库
git push -f origin feature
```

#### 6. 创建PR

* 访问你的Fork仓库 https://github.com/$user/kubespace
* 在`Pull requests`中点击`New pull request`创建PR
* 在PR中选择推送的分支feature，查看提交信息，确认无误后点击`Create pull request`

### 二、整体架构

### 三、本地运行

#### 1. 启动数据库

KubeSpace依赖Mysql以及Redis数据库，其中：
* Mysql用来存储如工作空间、应用、流水线等非Kubernetes数据；
* Redis用来当作事件通知订阅的消息中间件；

在本地直接通过docker快速启动mysql以及redis：
```bash
# 本地运行mysql
docker run --name kubespace-mysql -p 3306:3306 -v ~/kubespace/mysql:/var/lib/mysql -e MYSQL_ROOT_PASSWORD=kubespace@2022 -e MYSQL_DATABASE=kubespace -d mysql

# 本地运行redis
docker run --name kubespace-redis -p 6379:6379 -v ~/kubespace/redis:/data -d redis
```

如果已有mysql，需要创建kubespace数据库。

#### 2. 运行KubeSpace相关服务

克隆代码到本地目录：
```
git clone git@github.com:kubespace/kubespace.git ~/workspace/source/kubespace
```

##### 前端

前端是通过Vue框架开发，首先需要安装`node`以及`npm`，然后进入到代码目录。
```bash
# 安装前端依赖
cd ui && npm install
# 启动前端，port=9527指定前端端口号
port=9527 npm run serve
```

##### Server

Server是后端的接口服务，通过golang的gin框架开发，首先需要安装`go`，之后进入到代码目录。
```bash
# 设置go代理
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
# 安装依赖
go mod init
# 启动server
MYSQL_HOST=localhost:3306 MYSQL_PASSWORD=kubespace@2022 go run cmd/server
```

##### Controller-manager

Controller-manager是用来监听发送的事件，后台执行一些异步任务。
```bash
# 启动controller-manager
MYSQL_HOST=localhost:3306 MYSQL_PASSWORD=kubespace@2022 DATA_DIR=./data go run cmd/server
```

##### Spacelet

Spacelet节点是用来执行流水线构建任务的， 如代码编译、发布等。通过添加Spacelet节点可以降低每个节点的负载，同时能够并发处理更多的构建任务。

```bash
# 启动spacelet
DATA_DIR=./data SERVER_URL=http://127.0.0.1 go run cmd/spacelet
```

### 3. 代码模块

```
kubespace
├── apps    // 应用商店内置的应用，server启动时会读取该目录创建应用
├── build   // kubespace镜像构建Dockerfile
│   ├── amd64
│   └── arm64
├── chart   // helm chart构建目录
├── cmd     
│   ├── controller-manager  // controller-manager启动目录
│   ├── kube-agent          // kube-agent启动目录
│   ├── server              // server启动目录
└── pkg
    ├── controller          // 各controller模块
    │   └── pipelinerun     // 流水线构建controller
    │       └── plugins     // 流水线内置插件
    ├── core
    │   ├── datatype
    │   ├── db              // mysql以及redis实例化
    │   └── lock            // 内存锁，controller对执行的对象加锁
    ├── informer            // 事件通知模块，各controller调用不同的informer来监听不同的对象
    │   └── listwatcher     // 各事件对象的list、watch以及notify操作
    │       ├── cluster
    │       ├── config      // listwatcher配置
    │       ├── pipeline
    │       └── storage     // 基于redis消息通知以及事件
    ├── kubeagent           // kubeagent模块，通过代理访问k8s集群
    │   └── config.go       // kubeagent配置
    │   └── agent.go        // kubeagent启动入口
    │   └── tunnel.go       // 通过websocket连接server，接收server发送的命令
    ├── kubernetes          // kubernetes调用模块
    │   ├── config          // kubernetes配置
    │   ├── kubeclient      // 客户端连接
    │   ├── resource        // k8s各资源调用
    │   └── types           // kubernetes调用依赖的类型
    │   └── factory.go      // kubernetes资源工厂
    ├── model               // 数据库gorm类型以及操作
    │   ├── manager         // orm数据库操作
    │   │   ├── cluster
    │   │   ├── pipeline
    │   │   └── project
    │   ├── migrate         // 数据库升级迁移
    │   │   ├── migration
    │   │   └── v1_1
    │   └── types           // 数据库orm类型
    ├── server              // server启动入口
    │   ├── config          // server配置
    │   ├── router          // server路由配置
    │   └── views           // server服务接口
    │       ├── cluster     // 集群相关接口
    │       ├── pipeline    // 流水线相关接口
    │       ├── project     // 工作空间以及应用相关接口
    │       ├── serializers // 接口参数序列化
    │       ├── settings    // 平台配置接口
    │       └── user        // 用户相关接口
    ├── service             // 业务逻辑模块
    │   ├── cluster         // 集群业务
    │   ├── config          // service配置
    │   ├── pipeline        // 流水线相关业务逻辑
    │   │   └── schemas
    │   └── project         // 工作空间
    └── utils      
        ├── code            // 错误码
        └── git             // github/gitlab/gitee相关操作
```

一个Pod list（通过kubeconfig连接k8s）接口调用模块流转如下：
```
server/views/cluster/resource.go::list
--> service/cluster/kube_client.go::List
    --> service/cluster/kube_client.go::get_client      // 获取调用k8s集群的客户端（direct/agent）
--> service/cluster/kube_client.go::Request             // 客户端请求k8s资源资源
    --> service/cluster/direct_client.go::request       // 通过direct kubeconfig调用k8s
        --> kubernetes/factory.go::GetResource          // 通过kubernetes资源工厂获取对应的资源的handler
    --> kubernetes/factory.go::ResourceHandler.Handle
        --> kubernetes/resource/pod.go::Handle          // Pod handler处理请求
        --> kubernetes/resource.resource.go::Handle     // 统一的k8s资源处理
```

一个流水线构建调用模块流转如下：
```
// api接口调用
server/views/pipeline/pipeline_run.go::build
--> service/pipeline/pipeline_run.go::Build             // 调用service流水线构建逻辑
    --> model/manager/pipeline/pipeline_run.go::CreatePipelineRun   // orm创建构建任务
        --> informer/listwatcher/pipeline/pipelinerun_listwatcher.go::Nofiy // listwatcher调用nofity通知到对应controller
        --> informer/listwatcher/storage/redis.go::Nofity           // 底层通过redis pubsub事件通知
        
// 以下部分是通过controller-manager启动运行
controller/pipelinerun/pipelinerun_controller.go::Run
--> informer/informer.go::Run                           // informer统一运行入口
    --> informer/listwatcher/storage/redis.go::Run      // 通过redis pubsub监听事件
--> controller/pipelinerun/pipelinerun_controller.go::Handle    // controller后台执行流水线构建
    --> controller/pipelinerun/pipelinerun_controller.go::executeStage      // 执行流水线阶段
        --> controller/pipelinerun/pipelinerun_controller.go::executeJob    // 并发执行流水线任务
        --> controller/pipelinerun/plugins/plugin.go::Execute               // 执行流水线任务对应插件
```