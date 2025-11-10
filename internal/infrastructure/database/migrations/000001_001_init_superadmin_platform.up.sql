-- Ekstensi untuk UUID (jika belum ada)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. SUPERADMIN
CREATE TABLE "admin_users" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "name" VARCHAR(255) NOT NULL,
  "email" VARCHAR(255) NOT NULL UNIQUE,
  "password" VARCHAR(255) NOT NULL,
  "remember_token" VARCHAR(100) NULL,
  "created_at" TIMESTAMPTZ DEFAULT (NOW()),
  "updated_at" TIMESTAMPTZ DEFAULT (NOW()),
  "deleted_at" TIMESTAMPTZ NULL
);

CREATE TABLE "admin_permissions" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "name" VARCHAR(255) NOT NULL UNIQUE, -- e.g., "manage:tenants"
  "group_name" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMPTZ DEFAULT (NOW()),
  "updated_at" TIMESTAMPTZ DEFAULT (NOW()),
  "deleted_at" TIMESTAMPTZ NULL
);

CREATE TABLE "admin_roles" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "name" VARCHAR(255) NOT NULL UNIQUE,
  "description" TEXT NULL,
  "created_at" TIMESTAMPTZ DEFAULT (NOW()),
  "updated_at" TIMESTAMPTZ DEFAULT (NOW()),
  "deleted_at" TIMESTAMPTZ NULL
);

CREATE TABLE "admin_permission_role" (
  "permission_id" UUID NOT NULL REFERENCES "admin_permissions"("id") ON DELETE CASCADE,
  "admin_role_id" UUID NOT NULL REFERENCES "admin_roles"("id") ON DELETE CASCADE,
  PRIMARY KEY ("permission_id", "admin_role_id")
);

CREATE TABLE "admin_role_user" (
  "admin_user_id" UUID NOT NULL REFERENCES "admin_users"("id") ON DELETE CASCADE,
  "admin_role_id" UUID NOT NULL REFERENCES "admin_roles"("id") ON DELETE CASCADE,
  PRIMARY KEY ("admin_user_id", "admin_role_id")
);

-- 2. PLATFORM (Manajemen Tenant)
CREATE TABLE "tenants" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL,
  "slug" VARCHAR(255) NOT NULL UNIQUE,
  "company_email" VARCHAR(255) NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'setup_pending' CHECK ("status" IN ('active', 'inactive', 'suspended', 'setup_pending')),
  "created_at" TIMESTAMPTZ DEFAULT (now()),
  "updated_at" TIMESTAMPTZ DEFAULT (now()),
  "deleted_at" TIMESTAMPTZ NULL
);

CREATE TABLE "plans" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL,
  "slug" VARCHAR(255) NOT NULL UNIQUE,
  "description" TEXT NULL,
  "price" DECIMAL(15, 2) NOT NULL,
  "billing_cycle" VARCHAR(50) NOT NULL DEFAULT 'monthly' CHECK ("billing_cycle" IN ('monthly', 'yearly')),
  "employee_limit" INT NOT NULL DEFAULT 50,
  "is_active" BOOLEAN NOT NULL DEFAULT TRUE,
  "created_at" TIMESTAMPTZ DEFAULT (now()),
  "updated_at" TIMESTAMPTZ DEFAULT (now()),
  "deleted_at" TIMESTAMPTZ NULL
);

CREATE TABLE "features" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL,
  "slug" VARCHAR(255) NOT NULL UNIQUE,
  "description" TEXT NULL,
  "is_addon" BOOLEAN NOT NULL DEFAULT FALSE,
  "created_at" TIMESTAMPTZ DEFAULT (now()),
  "updated_at" TIMESTAMPTZ DEFAULT (now()),
  "deleted_at" TIMESTAMPTZ NULL
);

CREATE TABLE "feature_plan" (
  "feature_id" UUID NOT NULL REFERENCES "features"("id") ON DELETE CASCADE,
  "plan_id" UUID NOT NULL REFERENCES "plans"("id") ON DELETE CASCADE,
  PRIMARY KEY ("feature_id", "plan_id")
);

CREATE INDEX ON "admin_role_user" ("admin_user_id");
CREATE INDEX ON "admin_role_user" ("admin_role_id");
CREATE INDEX ON "admin_permission_role" ("permission_id");
CREATE INDEX ON "admin_permission_role" ("admin_role_id");
CREATE INDEX ON "tenants" ("slug");