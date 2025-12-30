# Database Migrations

## 001_permission_refactor

权限系统重构迁移脚本。

### 变更内容

1. **角色名称迁移**: 将 `article_collaborators` 表中的 `role='owner'` 改为 `role='admin'`
2. **软删除支持**: 为 `articles` 表添加 `deleted_at` 字段

### 文件说明

| 文件 | 说明 |
|------|------|
| `001_permission_refactor.sql` | 正向迁移脚本 |
| `001_permission_refactor_rollback.sql` | 回滚脚本 |
| `001_permission_refactor_dryrun.sql` | 预览脚本（不修改数据） |

### 使用方法

#### 1. 预览变更（Dry Run）

```bash
# PostgreSQL
psql -d your_database -f 001_permission_refactor_dryrun.sql

# MySQL
mysql -u user -p your_database < 001_permission_refactor_dryrun.sql
```

#### 2. 执行迁移

```bash
# PostgreSQL
psql -d your_database -f 001_permission_refactor.sql

# MySQL
mysql -u user -p your_database < 001_permission_refactor.sql
```

#### 3. 回滚（如需要）

```bash
# PostgreSQL
psql -d your_database -f 001_permission_refactor_rollback.sql

# MySQL
mysql -u user -p your_database < 001_permission_refactor_rollback.sql
```

### 验证迁移

迁移后运行以下查询验证：

```sql
-- 检查角色分布
SELECT role, COUNT(*) FROM article_collaborators GROUP BY role;

-- 检查 deleted_at 字段
SELECT column_name FROM information_schema.columns 
WHERE table_name = 'articles' AND column_name = 'deleted_at';

-- 检查备份表
SELECT COUNT(*) FROM article_collaborators_backup_permission_refactor;
```

### 注意事项

- 迁移前会自动创建 `article_collaborators_backup_permission_refactor` 备份表
- 回滚脚本会从备份表恢复原始数据
- 建议在生产环境执行前先在测试环境验证
