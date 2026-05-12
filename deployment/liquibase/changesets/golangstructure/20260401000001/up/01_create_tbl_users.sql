CREATE TABLE "tbl_users" (
    "id"         SERIAL       PRIMARY KEY,
    "name"       VARCHAR      NOT NULL,
    "age"        INTEGER      NOT NULL,
    "email"      VARCHAR      NOT NULL UNIQUE,
    "password"   VARCHAR      NOT NULL,
    "created_at" TIMESTAMP    NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP    NOT NULL DEFAULT now(),
    "deleted_at" TIMESTAMP
);
