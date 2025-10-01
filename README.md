# go大仓

## 文件结构

一个包的结构大概这样:

- config 存放配置
- internal 存放包内部使用的代码, 无法被其它包导入
  - pkg 内部使用的工具
  - ... 其它
- pkg 外部可使用的工具
- cmd 可执行的文件/命令

service最好不要导出东西, 导入package里的就可以了. 

## 运行

在子包运行

```sh
make install
```

```sh
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
```

```sh
go run xxx
```

## 暂时的预期

### auth-service

鉴于软工集市登录复用的麻烦, 这里把认证服务单独抽出来了. 预计有这些接口:

- login

预计返回refreshToken和accessToken
- register
- oauth/{provider}/authorize?redirect_url=xxx

预计返回一个授权链接, 比如GitHub的, 那么跳到https://github.com/login/oauth/authorize?client_id=xxx&state=xxx, 在授权链接授权后, 重定向到xxx, 并在url里带上认证code, 前端需要把这个code发给后端
- oauth/{provider}/callback

预计接收code, 并通过这个code询问第三方平台, 最后返回登录结果
- refresh

使用refreshToken刷新token. 
redis会存储refreshToken:accessToken, refresh后删除该token对, 再生成一份
- validate

需要传入accessToken

### auth-sdk

预计导出这些东西供外部使用:

- authMiddleware

认证中间件, 处理用户鉴权