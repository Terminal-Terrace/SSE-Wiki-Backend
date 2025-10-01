# SSE-Wiki-Backend

## 文件结构

- cmd
  - server 存放服务器相关的命令
- config 配置文件
- internal 内部使用, 外部无法导入的代码文件
  - database 数据库相关的初始化操作
  - dto 统一的返回格式, 约定接口的request和response
  - handler 与http层交互, 进行简单的数据校验, 如判空等; 不处理业务, 将数据交给service层处理
  - service 业务实际处理的地方
  - middleware 中间件, 还没写
  - route 路由
  - model 数据库表的结构