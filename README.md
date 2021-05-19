# gopractice
## 关于
1. 提供TarsGo标准化服务的公共库，集成日志、监控、调用链、JSON网关、错误码、配置、初始化等功能。
2. 提供TarsGo创建服务的命令行工具(tarsgo)。

## TarsGo使用标准化方案
- 使用go mod管理依赖
- 默认初始化日志、调用链、监控和配置文件
- 管理golang/tars2go/makefile的版本
- 错误码区分客户端还是服务端错误
- 集群外使用json网关来访问tars服务

## tarsgo 命令行使用示例

### init - 初始化项目

```
tarsgo init github.com/yourname/example
```

参数：go mod名、profile名（默认standard）

创建项目包含基本目录结构及文件
- go.mod
- meta/servers.yaml meta/config.yaml 项目元数据信息
- apps/demoserver 示例项目


### create - 创建服务

```
tarsgo create TestApp DemoServer HelloObj
```

参数: App/Server/Obj名

在apps目录下创建一个服务，并有项目的基本标准化代码

### build - 编译服务并生成代码包或镜像

参数: tgz/img

### upload - 编译并上传部署服务

- 对于vm类型服务：调用tarsweb接口部署
- 对于k8s服务：调用kubectl部署
