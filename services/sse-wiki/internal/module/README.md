# 模块权限

## 角色定义

| 角色 | 来源 | 说明 |
|------|------|------|
| Global_Admin | JWT role="admin" | 系统全局管理员，拥有所有模块权限 |
| Owner | modules.owner_id | 模块创建者，拥有该模块完全管理权限 |
| Admin | module_moderators.role="admin" | 被授权的模块管理员 |
| Moderator | module_moderators.role="moderator" | 模块协作者，权限受限 |
| User | 已登录用户 | 普通用户，无管理权限 |
| Guest | 未登录 | 游客，仅可浏览 |

## 权限矩阵

Y = 允许, N = 禁止

| 操作 | Global_Admin | Owner | Admin | Moderator | User | Guest |
|------|:------------:|:-----:|:-----:|:---------:|:----:|:-----:|
| **浏览** |
| 查看模块树 | Y | Y | Y | Y | Y | Y |
| 查看模块详情 | Y | Y | Y | Y | Y | Y |
| 查看面包屑导航 | Y | Y | Y | Y | Y | Y |
| **模块管理** |
| 创建顶级模块 | Y | N | N | N | N | N |
| 创建子模块 | Y | Y | Y | Y | N | N |
| 编辑模块名称 | Y | Y | Y | Y | N | N |
| 编辑模块描述 | Y | Y | Y | Y | N | N |
| 移动模块 | Y | Y | Y | N | N | N |
| 删除模块 | Y | Y | Y | N | N | N |
| **协作者管理** |
| 查看协作者列表 | Y | Y | Y | Y | N | N |
| 添加 Admin | Y | Y | N | N | N | N |
| 添加 Moderator | Y | Y | Y | N | N | N |
| 移除协作者 | Y | Y | Y | N | N | N |
| **编辑锁** |
| 获取/释放编辑锁 | Y | Y | Y | Y | N | N |

## 权限继承规则

当创建子模块时，权限按以下规则继承：

| 父模块角色 | 子模块继承角色 |
|-----------|---------------|
| 子模块创建者 | Owner |
| 父模块 Owner | Admin |
| 父模块 Admin | Admin |
| 父模块 Moderator | Moderator |

## is_moderator 计算规则

`GetModuleTree` 接口返回的 `is_moderator` 字段计算顺序：

1. 检查 JWT role 是否为 "admin" (Global_Admin)
2. 检查 modules.owner_id 是否等于当前用户 ID
3. 检查 module_moderators 表是否有该用户记录

任一条件满足则返回 true。

## 关键约束

- Moderator 不能移动或删除模块
- Admin 不能添加其他 Admin，只有 Owner 和 Global_Admin 可以
- 删除模块会级联删除所有子模块
