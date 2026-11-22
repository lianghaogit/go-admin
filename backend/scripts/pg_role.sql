-- ============================================================
-- go-admin 专用最小权限数据库账号（PostgreSQL）
-- 用途：应用连接串不再使用 postgres 超级用户，降低误操作与提权风险。
-- 用法：
--   docker exec -i my-postgres psql -U postgres -d go_admin -v ON_ERROR_STOP=1 < backend/scripts/pg_role.sql
-- 执行后用 backend/config.yaml 把 user/password 改成 go_admin_app/<自定义密码>。
-- 注意：账号只授予 DML 权限；若关闭 database.autoMigrate 后需要变更表结构，
--       请由 postgres 超级用户单独执行 DDL（本账号无 CREATE 权限）。
-- ============================================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'go_admin_app') THEN
        CREATE ROLE go_admin_app LOGIN PASSWORD 'change-me-in-production';
    END IF;
END $$;

-- 连接与 schema 使用权限
GRANT CONNECT ON DATABASE go_admin TO go_admin_app;
GRANT USAGE ON SCHEMA public TO go_admin_app;

-- 现有表与序列的读写权限（含 pg_trgm 等扩展对象）
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO go_admin_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO go_admin_app;

-- 之后新建的表/序列自动继承权限（由 postgres 执行 DDL 后生效）
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO go_admin_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO go_admin_app;

-- 只读统计视图所需权限（数据监控里的连接数/缓存命中率由 pg_stat_activity、pg_stat_database 提供，
-- 这两者是 PUBLIC 可读，无需额外授权）
