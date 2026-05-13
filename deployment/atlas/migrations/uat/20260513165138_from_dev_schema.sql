-- Create "tbl_user_logs" table
CREATE TABLE "public"."tbl_user_logs" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "action" character varying NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY ("id")
);
-- Create "tbl_users" table
CREATE TABLE "public"."tbl_users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying NOT NULL,
  "age" integer NOT NULL,
  "email" character varying NOT NULL,
  "password" character varying NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now(),
  "deleted_at" timestamp NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "tbl_users_email_key" UNIQUE ("email")
);
