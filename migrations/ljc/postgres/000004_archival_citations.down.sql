-- Evidence is audit data. Application rollback must not delete it.
-- Deliberately retain both tables; no destructive automatic down migration.
SELECT 1;
