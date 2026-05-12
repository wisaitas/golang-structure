DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = current_schema()
          AND table_name = 'tbl_users'
    ) THEN
        RAISE EXCEPTION 'verify failed: table tbl_users does not exist in schema %', current_schema();
    ELSE
        RAISE NOTICE 'verify OK: table tbl_users exists in schema %', current_schema();
    END IF;
END $$;
