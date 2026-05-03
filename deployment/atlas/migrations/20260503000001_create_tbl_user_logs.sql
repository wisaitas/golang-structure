CREATE TABLE "tbl_user_logs" (
    "id"         SERIAL       PRIMARY KEY,
    "user_id"    INTEGER      NOT NULL,
    "action"     VARCHAR      NOT NULL,
    "created_at" TIMESTAMP    NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP    NOT NULL DEFAULT now()
);