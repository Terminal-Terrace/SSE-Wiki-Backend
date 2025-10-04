# go大仓

## 项目结构

```
SSE-Wiki-Backend/
├── services/              # 微服务目录
│   ├── sse-wiki/          # SSE Wiki 主服务
│   ├── auth-service/      # 认证服务
│   └── template/          # 服务模板（参考标准）
├── packages/              # 共享包目录
│   ├── database/          # 统一数据库连接管理
│   ├── response/          # 统一响应格式
│   └── auth-sdk/          # 认证 SDK
├── .env.example           # 环境变量模板
├── go.work                # Go Workspace 配置
├── Makefile               # 构建脚本
└── README.md              # 本文件
```

## 文件结构

一个包的结构大概这样:

- config 存放配置
- internal 存放包内部使用的代码, 无法被其它包导入
  - pkg 内部使用的工具
  - ... 其它
- pkg 外部可使用的工具
- cmd 可执行的文件/命令

service最好不要导出东西, 导入package里的就可以了. 

## 快速开始

### 环境

待补充各软件版本

### 配置

`.env.example` 为环境变量模板，需要配置拷贝并命名为`.env`，完成内部相关配置

### 运行

在子包运行

```sh
make install
make run
```

或者在根目录运行

调用子包的`make install`

```sh
make install 子包名
```

调用子包的`make run`

```sh
make run 子包名
```

如果没有make, 也可以跟平时一样运行. 

```sh
go mod tidy
go run xxx
```

## 暂时的预期

### auth-service

参见[飞书文档](https://mcn0xmurkm53.feishu.cn/docx/C4z7dMc0co932PxyyXkcPT1Fn5e)

### auth-sdk

预计导出这些东西供外部使用:

- authMiddleware

认证中间件, 所有的服务应该都使用这个中间件. 处理用户鉴权. 顺便将一些用户信息存到上下文里. 