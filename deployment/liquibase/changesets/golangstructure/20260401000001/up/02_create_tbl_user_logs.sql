CREATE TABLE "tbl_user_logs" (
    "id"         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id"    UUID         NOT NULL,
    "action"     VARCHAR      NOT NULL,
    "created_at" TIMESTAMP    NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP    NOT NULL DEFAULT now()
);