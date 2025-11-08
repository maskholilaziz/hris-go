```sql
-- Pengguna Superadmin (Tim internal Anda)
CREATE TABLE "admin_users" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL,
  "email" VARCHAR(255) NOT NULL UNIQUE,
  "password" VARCHAR(255) NOT NULL,
  "remember_token" VARCHAR(100) NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Role untuk Superadmin
CREATE TABLE "admin_roles" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL UNIQUE,
  "description" TEXT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL
);

-- Pivot: Admin user punya role apa
CREATE TABLE "admin_role_user" (
  "admin_user_id" UUID NOT NULL REFERENCES "admin_users"("id") ON DELETE CASCADE,
  "admin_role_id" UUID NOT NULL REFERENCES "admin_roles"("id") ON DELETE CASCADE,
  PRIMARY KEY ("admin_user_id", "admin_role_id")
);

-- Pivot: Role admin punya permission apa
CREATE TABLE "admin_permission_role" (
  "permission_id" UUID NOT NULL REFERENCES "admin_permissions"("id") ON DELETE CASCADE,
  "admin_role_id" UUID NOT NULL REFERENCES "admin_roles"("id") ON DELETE CASCADE,
  PRIMARY KEY ("permission_id", "admin_role_id")
);

-- BARU: Master permission khusus untuk Superadmin
CREATE TABLE "admin_permissions" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL UNIQUE, -- e.g., "manage:tenants", "view:global_revenue"
  "group_name" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);
```

### \#\# 1. Platform & Billing

```sql
-- Klien/Perusahaan yang menggunakan aplikasi Anda
CREATE TABLE "tenants" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL,
  "slug" VARCHAR(255) NOT NULL UNIQUE,
  "logo_path" VARCHAR(255) NULL,
  "primary_color" VARCHAR(50) NULL,
  "company_email" VARCHAR(255) NULL,
  "phone_number" VARCHAR(50) NULL,
  "website" VARCHAR(255) NULL,
  "address" TEXT NULL,
  "city" VARCHAR(255) NULL,
  "province" VARCHAR(255) NULL,
  "postal_code" VARCHAR(20) NULL,
  "npwp" VARCHAR(255) NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'setup_pending' CHECK ("status" IN ('active', 'inactive', 'suspended', 'setup_pending')),
  "default_timezone" VARCHAR(100) NOT NULL DEFAULT 'Asia/Jakarta',
  "default_currency" VARCHAR(10) NOT NULL DEFAULT 'IDR',
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Katalog paket langganan
CREATE TABLE "plans" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL,
  "slug" VARCHAR(255) NOT NULL UNIQUE,
  "description" TEXT NULL,
  "price" DECIMAL(15, 2) NOT NULL,
  "billing_cycle" VARCHAR(50) NOT NULL DEFAULT 'monthly' CHECK ("billing_cycle" IN ('monthly', 'yearly')),
  "employee_limit" INT NOT NULL DEFAULT 50 CHECK ("employee_limit" >= 0),
  "is_active" BOOLEAN NOT NULL DEFAULT TRUE,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Catatan langganan aktif per tenant
CREATE TABLE "subscriptions" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "plan_id" UUID NOT NULL REFERENCES "plans"("id"),
  "status" VARCHAR(50) NOT NULL DEFAULT 'trialing' CHECK ("status" IN ('active', 'trialing', 'past_due', 'canceled')),
  "started_at" TIMESTAMPTZ NOT NULL,
  "ends_at" TIMESTAMPTZ NULL,
  "trial_ends_at" TIMESTAMPTZ NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL
);

-- BARU: Tabel untuk menagih tenant
CREATE TABLE "invoices" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id"),
  "subscription_id" UUID NOT NULL REFERENCES "subscriptions"("id"),
  "issue_date" DATE NOT NULL,
  "due_date" DATE NOT NULL,
  "total_amount" DECIMAL(15, 2) NOT NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'unpaid' CHECK ("status" IN ('unpaid', 'paid', 'overdue')),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL
);

-- BARU: Item di dalam setiap tagihan
CREATE TABLE "invoice_items" (
  "id" UUID PRIMARY KEY,
  "invoice_id" UUID NOT NULL REFERENCES "invoices"("id") ON DELETE CASCADE,
  "description" VARCHAR(255) NOT NULL,
  "quantity" INT NOT NULL DEFAULT 1,
  "unit_price" DECIMAL(15, 2) NOT NULL,
  "total_price" DECIMAL(15, 2) NOT NULL
);

-- Master daftar fitur
CREATE TABLE "features" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL,
  "slug" VARCHAR(255) NOT NULL UNIQUE,
  "description" TEXT NULL,
  "is_addon" BOOLEAN NOT NULL DEFAULT FALSE,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL
);

-- Pivot: Fitur apa saja yang ada di paket mana
CREATE TABLE "feature_plan" (
  "feature_id" UUID NOT NULL REFERENCES "features"("id") ON DELETE CASCADE,
  "plan_id" UUID NOT NULL REFERENCES "plans"("id") ON DELETE CASCADE,
  PRIMARY KEY ("feature_id", "plan_id")
);

-- Pivot: Addon apa saja yang dibeli oleh tenant
CREATE TABLE "tenant_addons" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "feature_id" UUID NOT NULL REFERENCES "features"("id") ON DELETE CASCADE,
  "expires_at" TIMESTAMPTZ NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  UNIQUE ("tenant_id", "feature_id")
);

-- DIPERBARUI: Menambahkan soft delete
CREATE TABLE "invoices" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id"),
  "subscription_id" UUID NOT NULL REFERENCES "subscriptions"("id"),
  "issue_date" DATE NOT NULL,
  "due_date" DATE NOT NULL,
  "total_amount" DECIMAL(15, 2) NOT NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'unpaid' CHECK ("status" IN ('unpaid', 'paid', 'overdue', 'void')),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL -- Ditambahkan
);

-- BARU: Tabel untuk mencatat pembayaran invoice (PRD 8.0)
CREATE TABLE "invoice_payments" (
  "id" UUID PRIMARY KEY,
  "invoice_id" UUID NOT NULL REFERENCES "invoices"("id") ON DELETE CASCADE,
  "payment_date" TIMESTAMPTZ NOT NULL,
  "amount_paid" DECIMAL(15, 2) NOT NULL,
  "payment_method" VARCHAR(100) NULL, -- e.g., 'bank_transfer', 'credit_card'
  "transaction_id" VARCHAR(255) NULL, -- Dari payment gateway
  "notes" TEXT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL
);

-- OPTIMASI: Indeks untuk Modul Billing & Platform
CREATE INDEX ON "subscriptions" ("tenant_id");
CREATE INDEX ON "subscriptions" ("plan_id");
CREATE INDEX ON "invoices" ("tenant_id");
CREATE INDEX ON "invoices" ("subscription_id");
CREATE INDEX ON "invoices" ("status");
CREATE INDEX ON "invoice_items" ("invoice_id");
CREATE INDEX ON "tenant_addons" ("tenant_id");
CREATE INDEX ON "tenant_addons" ("feature_id");

-- Indeks untuk pembayaran
CREATE INDEX ON "invoice_payments" ("invoice_id");
CREATE INDEX ON "invoice_payments" ("transaction_id");
```

---

### \#\# 2. Core HR & Perusahaan

```sql
-- Pengguna yang bisa login ke sistem
CREATE TABLE "users" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "email" VARCHAR(255) NOT NULL UNIQUE,
  "email_verified_at" TIMESTAMPTZ NULL,
  "password" VARCHAR(255) NOT NULL,
  "timezone" VARCHAR(100) NOT NULL DEFAULT 'UTC',
  "remember_token" VARCHAR(100) NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Strategic Business Unit
CREATE TABLE "sbus" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "description" TEXT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Lokasi kantor cabang
CREATE TABLE "branches" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "address" TEXT NULL,
  "latitude" DECIMAL(10, 8) NULL,
  "longitude" DECIMAL(11, 8) NULL,
  "attendance_radius_meters" INT NOT NULL DEFAULT 100 CHECK ("attendance_radius_meters" >= 0),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Master data departemen
CREATE TABLE "departments" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Master data jabatan
CREATE TABLE "positions" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Master data level jabatan
CREATE TABLE "job_levels" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Master data status kepegawaian
CREATE TABLE "employment_statuses" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Tabel inti data karyawan
CREATE TABLE "employees" (
  "id" UUID PRIMARY KEY,
  "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "branch_id" UUID NULL REFERENCES "branches"("id"),
  "department_id" UUID NULL REFERENCES "departments"("id"),
  "position_id" UUID NULL REFERENCES "positions"("id"),
  "job_level_id" UUID NULL REFERENCES "job_levels"("id"),
  "employment_status_id" UUID NULL REFERENCES "employment_statuses"("id"),
  "sbu_id" UUID NULL REFERENCES "sbus"("id"),
  "manager_id" UUID NULL REFERENCES "employees"("id") ON DELETE SET NULL,
  "employee_id_number" VARCHAR(255) NOT NULL, -- NIK Karyawan (bukan KTP)
  "join_date" DATE NOT NULL,
  "resign_date" DATE NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL,
  UNIQUE("tenant_id", "employee_id_number"),
  UNIQUE("user_id") -- Satu user hanya boleh jadi satu profil karyawan
);

-- BARU: Data Pribadi/Demografis Karyawan (PRD 2.1.1)
CREATE TABLE "employee_profiles" (
  "employee_id" UUID PRIMARY KEY REFERENCES "employees"("id") ON DELETE CASCADE, -- 1-to-1
  "full_name" VARCHAR(255) NOT NULL, -- Nama legal sesuai KTP
  "nik_ktp" VARCHAR(50) NULL,
  "place_of_birth" VARCHAR(255) NULL,
  "date_of_birth" DATE NULL,
  "gender" VARCHAR(50) NULL CHECK ("gender" IN ('male', 'female')),
  "marital_status" VARCHAR(50) NULL,
  "religion" VARCHAR(50) NULL,
  "address_ktp" TEXT NULL,
  "address_domicile" TEXT NULL,
  "emergency_contact_name" VARCHAR(255) NULL,
  "emergency_contact_phone" VARCHAR(50) NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- BARU: Data Finansial Karyawan (PRD 2.1.3 & 5.1.2)
CREATE TABLE "employee_financials" (
  "employee_id" UUID PRIMARY KEY REFERENCES "employees"("id") ON DELETE CASCADE, -- 1-to-1
  "bank_name" VARCHAR(100) NULL,
  "bank_account_number" VARCHAR(100) NULL,
  "bank_account_holder" VARCHAR(255) NULL,
  "npwp" VARCHAR(50) NULL,
  "ptkp_status" VARCHAR(10) NULL, -- e.g., 'TK/0', 'K/1'
  "bpjs_ketenagakerjaan_no" VARCHAR(50) NULL,
  "bpjs_kesehatan_no" VARCHAR(50) NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- BARU: Dokumen Karyawan (PRD 2.2.1 & 2.2.2)
CREATE TABLE "employee_documents" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "document_type" VARCHAR(255) NOT NULL, -- e.g., 'KTP', 'Kontrak Kerja', 'Ijazah'
  "file_path" VARCHAR(255) NOT NULL, -- Path ke S3/minio
  "file_name" VARCHAR(255) NOT NULL,
  "expiry_date" DATE NULL, -- Untuk pengingat masa berlaku (PRD 2.2.2)
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- OPTIMASI: Indeks untuk Core HR
CREATE INDEX ON "users" ("tenant_id");
CREATE INDEX ON "branches" ("tenant_id");
CREATE INDEX ON "departments" ("tenant_id");
CREATE INDEX ON "positions" ("tenant_id");
CREATE INDEX ON "job_levels" ("tenant_id");
CREATE INDEX ON "employment_statuses" ("tenant_id");
CREATE INDEX ON "sbus" ("tenant_id");

CREATE INDEX ON "employees" ("tenant_id");
CREATE INDEX ON "employees" ("branch_id");
CREATE INDEX ON "employees" ("department_id");
CREATE INDEX ON "employees" ("position_id");
CREATE INDEX ON "employees" ("manager_id");
CREATE INDEX ON "employees" ("full_name"); -- Untuk pencarian nama

CREATE INDEX ON "employee_documents" ("employee_id");
```

---

### \#\# 3. Access Control (RBAC)

```sql
-- Peran pengguna (dibatasi per tenant)
-- Peran pengguna (dibatasi per tenant)
CREATE TABLE "roles" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "description" TEXT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL,
  UNIQUE("tenant_id", "name") -- Role harus unik per tenant
);

-- Master hak akses (global)
CREATE TABLE "tenant_permissions" (
  "id" UUID PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL UNIQUE, -- e.g., "approve:leave", "view:payroll"
  "group_name" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL -- Ditambahkan
);

-- Pivot: User memiliki peran apa
CREATE TABLE "role_user" (
  "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "role_id" UUID NOT NULL REFERENCES "roles"("id") ON DELETE CASCADE,
  PRIMARY KEY ("user_id", "role_id")
);

-- Pivot: Peran memiliki hak akses apa
CREATE TABLE "permission_role" (
  "permission_id" UUID NOT NULL REFERENCES "tenant_permissions"("id") ON DELETE CASCADE,
  "role_id" UUID NOT NULL REFERENCES "roles"("id") ON DELETE CASCADE,
  PRIMARY KEY ("permission_id", "role_id")
);

CREATE TABLE "user_has_permissions" (
  "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "permission_id" UUID NOT NULL REFERENCES "tenant_permissions"("id") ON DELETE CASCADE,
  PRIMARY KEY ("user_id", "permission_id")
);

-- OPTIMASI: Indeks untuk RBAC
CREATE INDEX ON "roles" ("tenant_id");
CREATE INDEX ON "role_user" ("role_id"); -- Selain PK
CREATE INDEX ON "permission_role" ("role_id"); -- Selain PK
```

---

### \#\# 4. Customization

```sql
-- Definisi field kustom per tenant
CREATE TABLE "custom_fields" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "type" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Nilai field kustom per karyawan
CREATE TABLE "employee_custom_field_values" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "custom_field_id" UUID NOT NULL REFERENCES "custom_fields"("id") ON DELETE CASCADE,
  "value" TEXT NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL
);
```

---

### \#\# 5. Time Management

```sql
-- Master data shift kerja
CREATE TABLE "shifts" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "start_time" TIME NOT NULL,
  "end_time" TIME NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Penugasan shift per karyawan per hari
CREATE TABLE "employee_shifts" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "shift_id" UUID NOT NULL REFERENCES "shifts"("id") ON DELETE CASCADE,
  "date" DATE NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL
);

-- Catatan absensi
CREATE TABLE "attendances" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "clock_in" TIMESTAMPTZ NULL,
  "clock_out" TIMESTAMPTZ NULL,
  "clock_in_timezone" VARCHAR(100) NULL,
  "clock_out_timezone" VARCHAR(100) NULL,
  "clock_in_latitude" DECIMAL(10, 8) NULL,
  "clock_in_longitude" DECIMAL(11, 8) NULL,
  "clock_out_latitude" DECIMAL(10, 8) NULL,
  "clock_out_longitude" DECIMAL(11, 8) NULL,
  "liveness_validated" BOOLEAN NOT NULL DEFAULT FALSE,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
) PARTITION BY RANGE (clock_in);

-- Kebijakan/master data cuti
CREATE TABLE "time_off_policies" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "default_balance" DECIMAL(8, 2) NOT NULL DEFAULT 12.00,
  "is_unlimited" BOOLEAN NOT NULL DEFAULT FALSE,
  "is_prorated" BOOLEAN NOT NULL DEFAULT FALSE, -- PRD 4.1.2
  "can_carry_forward" BOOLEAN NOT NULL DEFAULT FALSE, -- PRD 4.1.2
  "carry_forward_max_days" DECIMAL(8, 2) NULL,
  "carry_forward_expiry_months" INT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- BARU: Ledger saldo cuti karyawan (PRD 4.3.1)
CREATE TABLE "employee_time_off_balances" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "policy_id" UUID NOT NULL REFERENCES "time_off_policies"("id") ON DELETE CASCADE,
  "period_year" INT NOT NULL,
  "balance" DECIMAL(8, 2) NOT NULL DEFAULT 0.00,
  "accrued" DECIMAL(8, 2) NOT NULL DEFAULT 0.00,
  "taken" DECIMAL(8, 2) NOT NULL DEFAULT 0.00,
  "carry_forward" DECIMAL(8, 2) NOT NULL DEFAULT 0.00,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,

  -- Kunci unik untuk mencegah duplikasi ledger
  UNIQUE("employee_id", "policy_id", "period_year")
);

-- Pengajuan cuti
CREATE TABLE "time_off_requests" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "policy_id" UUID NOT NULL REFERENCES "time_off_policies"("id"),
  "start_date" DATE NOT NULL,
  "end_date" DATE NOT NULL,
  "is_hourly" BOOLEAN NOT NULL DEFAULT FALSE,
  "start_time" TIME NULL,
  "end_time" TIME NULL,
  "reason" TEXT NOT NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK ("status" IN ('pending', 'approved_by_manager', 'approved_by_hr', 'rejected')),
  "approved_by" UUID NULL REFERENCES "users"("id"),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Pengajuan lembur
CREATE TABLE "overtime_requests" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "date" DATE NOT NULL,
  "start_time" TIME NOT NULL,
  "end_time" TIME NOT NULL,
  "reason" TEXT NOT NULL,
  "calculated_hours" DECIMAL(8, 2) NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK ("status" IN ('pending', 'approved', 'rejected')),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- OPTIMASI: Indeks untuk Modul Waktu
CREATE INDEX ON "time_off_policies" ("tenant_id");
CREATE INDEX ON "time_off_requests" ("employee_id");
CREATE INDEX ON "time_off_requests" ("policy_id");
CREATE INDEX ON "time_off_requests" ("status");
CREATE INDEX ON "time_off_requests" ("employee_id", "start_date");

CREATE INDEX ON "overtime_requests" ("employee_id");
CREATE INDEX ON "overtime_requests" ("status");
CREATE INDEX ON "overtime_requests" ("date");
```

---

### \#\# 6. Payroll & Finance

```sql
-- BARU: Pengaturan Payroll per tenant (PRD 5.1.2)
CREATE TABLE "payroll_settings" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "pph21_method" VARCHAR(50) NOT NULL DEFAULT 'gross' CHECK ("pph21_method" IN ('gross', 'gross_up', 'nett')),
  "payroll_period" VARCHAR(50) NOT NULL DEFAULT 'monthly',
  "payroll_cutoff_date" INT NOT NULL DEFAULT 25,
  "payroll_payment_date" INT NOT NULL DEFAULT 30,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL,

  -- Hanya boleh ada 1 setting per tenant
  UNIQUE("tenant_id")
);

-- Master komponen gaji
CREATE TABLE "payroll_components" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "type" VARCHAR(50) NOT NULL CHECK ("type" IN ('earning', 'deduction')),
  "is_taxable" BOOLEAN NOT NULL DEFAULT TRUE,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Komponen gaji spesifik per karyawan
CREATE TABLE "employee_payroll_components" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "component_id" UUID NOT NULL REFERENCES "payroll_components"("id") ON DELETE CASCADE,
  "amount" DECIMAL(15, 2) NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL
);

-- Batch penggajian per periode
CREATE TABLE "payrolls" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "period_name" VARCHAR(255) NOT NULL,
  "period_start" DATE NOT NULL,
  "period_end" DATE NOT NULL,
  "payment_date" DATE NOT NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK ("status" IN ('draft', 'processed', 'paid')),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Hasil akhir gaji per karyawan per periode
CREATE TABLE "payroll_details" (
  "id" UUID PRIMARY KEY,
  "payroll_id" UUID NOT NULL REFERENCES "payrolls"("id") ON DELETE CASCADE,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "gross_salary" DECIMAL(15, 2) NOT NULL,
  "total_deductions" DECIMAL(15, 2) NOT NULL,
  "net_salary" DECIMAL(15, 2) NOT NULL,
  "component_breakdown" JSONB NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Pengajuan pinjaman
CREATE TABLE "loans" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "principal_amount" DECIMAL(15, 2) NOT NULL,
  "installments_count" INT NOT NULL CHECK ("installments_count" > 0),
  "monthly_installment" DECIMAL(15, 2) NOT NULL,
  "start_date" DATE NOT NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'ongoing' CHECK ("status" IN ('ongoing', 'paid_off')),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Pengajuan klaim/reimbursement
CREATE TABLE "expense_claims" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "claim_date" DATE NOT NULL,
  "description" VARCHAR(255) NOT NULL,
  "amount" DECIMAL(15, 2) NOT NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK ("status" IN ('pending', 'approved', 'rejected')),
  "approved_by" UUID NULL REFERENCES "users"("id"),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Pengajuan akses gaji lebih awal (EWA)
CREATE TABLE "early_wage_access_requests" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "amount" DECIMAL(15, 2) NOT NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK ("status" IN ('pending', 'approved', 'rejected', 'deducted')),
  "approved_by" UUID NULL REFERENCES "users"("id"),
  "payroll_detail_id" UUID NULL REFERENCES "payroll_details"("id"),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- OPTIMASI: Indeks untuk Modul Payroll
CREATE INDEX ON "payroll_components" ("tenant_id");
CREATE INDEX ON "employee_payroll_components" ("employee_id");
CREATE INDEX ON "employee_payroll_components" ("component_id");

CREATE INDEX ON "payrolls" ("tenant_id");
CREATE INDEX ON "payrolls" ("status");
CREATE INDEX ON "payrolls" ("period_start", "period_end");

CREATE INDEX ON "payroll_details" ("payroll_id");
CREATE INDEX ON "payroll_details" ("employee_id");
CREATE INDEX ON "payroll_details" ("employee_id", "payroll_id"); -- Indeks komposit

CREATE INDEX ON "loans" ("employee_id");
CREATE INDEX ON "loans" ("status");

CREATE INDEX ON "expense_claims" ("employee_id");
CREATE INDEX ON "expense_claims" ("status");

CREATE INDEX ON "early_wage_access_requests" ("employee_id");
CREATE INDEX ON "early_wage_access_requests" ("status");
```

---

### \#\# 7. Talent Lifecycle

```sql
-- Perencanaan tenaga kerja (MPP)
CREATE TABLE "manpower_plans" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "period" VARCHAR(255) NOT NULL,
  "department_id" UUID NOT NULL REFERENCES "departments"("id"),
  "position_id" UUID NOT NULL REFERENCES "positions"("id"),
  "headcount_needed" INT NOT NULL CHECK ("headcount_needed" >= 0),
  "status" VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK ("status" IN ('draft', 'approved')),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Lowongan pekerjaan
CREATE TABLE "job_vacancies" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "position_id" UUID NOT NULL REFERENCES "positions"("id"),
  "title" VARCHAR(255) NOT NULL,
  "description" TEXT NOT NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'open' CHECK ("status" IN ('open', 'closed')),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Database kandidat
CREATE TABLE "candidates" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "full_name" VARCHAR(255) NOT NULL,
  "email" VARCHAR(255) NOT NULL UNIQUE,
  "phone_number" VARCHAR(255) NOT NULL,
  "resume_path" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Lamaran kerja dari kandidat
CREATE TABLE "job_applications" (
  "id" UUID PRIMARY KEY,
  "candidate_id" UUID NOT NULL REFERENCES "candidates"("id") ON DELETE CASCADE,
  "job_vacancy_id" UUID NOT NULL REFERENCES "job_vacancies"("id") ON DELETE CASCADE,
  "status" VARCHAR(50) NOT NULL CHECK ("status" IN ('applied', 'screening', 'interview', 'offered', 'hired', 'rejected')),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL
);

-- Template checklist On/Offboarding
CREATE TABLE "onboarding_offboarding_templates" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "type" VARCHAR(50) NOT NULL CHECK ("type" IN ('onboarding', 'offboarding')),
  "tasks" JSONB NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Penugasan On/Offboarding ke karyawan
CREATE TABLE "employee_onboarding_offboardings" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "template_id" UUID NOT NULL REFERENCES "onboarding_offboarding_templates"("id") ON DELETE CASCADE,
  "status" VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK ("status" IN ('pending', 'in_progress', 'completed')),
  "task_statuses" JSONB NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL
);
```

---

### \#\# 8. Employee Experience & Benefits

```sql
-- Katalog benefit fleksibel
CREATE TABLE "flex_benefit_catalogs" (
  "id" UUID PRIMARY KEY,
  "tenant_id" UUID NOT NULL REFERENCES "tenants"("id") ON DELETE CASCADE,
  "name" VARCHAR(255) NOT NULL,
  "description" TEXT NULL,
  "cost_in_points" DECIMAL(15, 2) NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Pilihan benefit oleh karyawan
CREATE TABLE "employee_flex_selections" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "catalog_id" UUID NOT NULL REFERENCES "flex_benefit_catalogs"("id") ON DELETE CASCADE,
  "selection_period" VARCHAR(255) NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL
);
```

---

### \#\# 9. Performance Management

```sql
-- Siklus penilaian kinerja
CREATE TABLE "performance_reviews" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "review_period" VARCHAR(255) NOT NULL,
  "manager_feedback" TEXT NULL,
  "employee_self_assessment" TEXT NULL,
  "final_score" DECIMAL(5, 2) NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK ("status" IN ('pending', 'in_progress', 'completed')),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Rencana pengembangan individu
CREATE TABLE "development_plans" (
  "id" UUID PRIMARY KEY,
  "employee_id" UUID NOT NULL REFERENCES "employees"("id") ON DELETE CASCADE,
  "plan_period" VARCHAR(255) NOT NULL,
  "career_goals" TEXT NOT NULL,
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Tabel 'goals' (polimorfik)
CREATE TABLE "goals" (
  "id" UUID PRIMARY KEY,
  "goalable_id" UUID NOT NULL,
  "goalable_type" VARCHAR(255) NOT NULL,
  "title" VARCHAR(255) NOT NULL,
  "description" TEXT NULL,
  "due_date" DATE NULL,
  "status" VARCHAR(50) NOT NULL DEFAULT 'not_started' CHECK ("status" IN ('not_started', 'in_progress', 'completed')),
  "created_at" TIMESTAMPTZ NULL,
  "updated_at" TIMESTAMPTZ NULL,
  "deleted_at" TIMESTAMPTZ NULL
);

-- Index untuk tabel polimorfik 'goals'
CREATE INDEX "goals_goalable_index" ON "goals" ("goalable_id", "goalable_type");
```
