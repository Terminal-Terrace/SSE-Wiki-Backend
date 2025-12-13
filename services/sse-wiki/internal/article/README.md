# 文章权限

## 角色定义

| 角色 | 来源 | 说明 |
|------|------|------|
| Global_Admin | JWT role="admin" | 系统管理员，对文章仅有删除权限 |
| Author | articles.created_by | 文章创建者，拥有完全管理权限 |
| Admin | article_collaborators.role="admin" | 被授权的文章管理员 |
| Moderator | article_collaborators.role="moderator" | 文章协作者/审核员 |
| User | 已登录用户 | 普通用户，可提交修改 |
| Guest | 未登录 | 游客，仅可浏览 |

## 权限矩阵

Y = 允许, N = 禁止

| 操作 | Global_Admin | Author | Admin | Moderator | User | Guest |
|------|:------------:|:------:|:-----:|:---------:|:----:|:-----:|
| **浏览** |
| 查看文章列表 | Y | Y | Y | Y | Y | Y |
| 查看文章详情 | Y | Y | Y | Y | Y | Y |
| 查看版本历史 | Y | Y | Y | Y | Y | N |
| 查看版本内容 | Y | Y | Y | Y | Y | N |
| 查看版本 Diff | Y | Y | Y | Y | Y | N |
| **文章创建** |
| 创建文章 | Y | Y | Y | Y | Y | N |
| **内容编辑** |
| 提交内容修改 | Y | Y | Y | Y | Y | N |
| 直接发布(跳过审核) | N | Y | Y | Y | N | N |
| **基础信息编辑** |
| 编辑标题 | N | Y | Y | Y | N | N |
| 编辑标签 | N | Y | Y | Y | N | N |
| 设置审核开关 | N | Y | Y | Y | N | N |
| **审核操作** |
| 查看审核列表 | Y | Y | Y | Y | Y | N |
| 查看审核详情 | Y | Y | Y | Y | Y | N |
| 批准/驳回提交 | N | Y | Y | Y | N | N |
| 解决合并冲突 | N | Y | Y | Y | 仅自己的 | N |
| **协作者管理** |
| 查看协作者列表 | Y | Y | Y | Y | N | N |
| 添加 Admin | N | Y | N | N | N | N |
| 添加 Moderator | N | Y | Y | N | N | N |
| 移除协作者 | N | Y | Y | N | N | N |
| **文章删除** |
| 删除文章 | Y | Y | Y | N | N | N |
| **收藏** |
| 收藏/取消收藏 | Y | Y | Y | Y | Y | N |

## 审核流程

### 提交时处理逻辑

| 条件 | 处理方式 |
|------|----------|
| is_review_required = false | 直接发布（3路合并） |
| is_review_required = true + Author/Admin/Moderator | 直接发布（3路合并） |
| is_review_required = true + Global_Admin | 创建待审核提交（与普通用户相同） |
| is_review_required = true + User | 创建待审核提交 |

### 审核权限

| 角色 | 自己的提交 | 他人的提交 |
|------|-----------|-----------|
| Global_Admin | 需审核 | 不可审核 |
| Author | 直接发布 | 可审核 |
| Admin | 直接发布 | 可审核 |
| Moderator | 直接发布 | 可审核 |

## 关键约束

- Global_Admin 对文章仅有删除权限，编辑/审核与普通用户相同
- Author 不可被移除
- 只有 Author 可以添加 Admin 协作者
- Admin 可以添加 Moderator，但不能添加 Admin
- 角色命名统一使用 admin/moderator（需迁移现有 owner 为 admin）

## 数据迁移

article_collaborators 表角色迁移：

```sql
UPDATE article_collaborators SET role = 'admin' WHERE role = 'owner';
```
