-- ─────────────────────────────────────────────────────────────────────────────
-- PT. Dana Sejahtera — Postgres Initialisation Script
--
-- Executed ONCE by the postgres Docker container on first volume creation.
-- Column names and types exactly match the GORM models in
-- backend/internal/models/ so that AutoMigrate is a clean no-op.
-- ─────────────────────────────────────────────────────────────────────────────

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── Users ────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    username       VARCHAR(100) UNIQUE NOT NULL,
    email          VARCHAR(255) UNIQUE NOT NULL,
    password_hash  VARCHAR(255) NOT NULL,
    role           VARCHAR(50)  NOT NULL DEFAULT 'nasabah'
                                CHECK (role IN ('admin', 'staff', 'nasabah')),
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    login_attempts INTEGER      NOT NULL DEFAULT 0,
    last_login_at  TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email    ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role     ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_deleted  ON users(deleted_at);

-- ─── Nasabah ──────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS nasabah (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID         UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    full_name     VARCHAR(255) NOT NULL,
    nik           VARCHAR(16)  UNIQUE NOT NULL,
    phone         VARCHAR(20),
    address       TEXT,
    date_of_birth TIMESTAMPTZ  NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_nasabah_user_id ON nasabah(user_id);
CREATE INDEX IF NOT EXISTS idx_nasabah_nik     ON nasabah(nik);
CREATE INDEX IF NOT EXISTS idx_nasabah_deleted ON nasabah(deleted_at);

-- ─── Loans ────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS loans (
    id            UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    nasabah_id    UUID           NOT NULL REFERENCES nasabah(id) ON DELETE CASCADE,
    amount        DECIMAL(15,2)  NOT NULL CHECK (amount > 0),
    interest_rate DECIMAL(5,2)   NOT NULL CHECK (interest_rate > 0),
    term_months   INTEGER        NOT NULL CHECK (term_months > 0),
    status        VARCHAR(50)    NOT NULL DEFAULT 'pending'
                                 CHECK (status IN ('pending','approved','rejected','active','closed')),
    approved_by   UUID           REFERENCES users(id),
    approved_at   TIMESTAMPTZ,
    notes         TEXT,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_loans_nasabah_id ON loans(nasabah_id);
CREATE INDEX IF NOT EXISTS idx_loans_status     ON loans(status);
CREATE INDEX IF NOT EXISTS idx_loans_created_at ON loans(created_at);

-- ─── Transactions ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS transactions (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    loan_id          UUID          NOT NULL REFERENCES loans(id) ON DELETE CASCADE,
    amount           DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    transaction_type VARCHAR(50)   NOT NULL
                                   CHECK (transaction_type IN ('disbursement','repayment','penalty')),
    description      TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tx_loan_id    ON transactions(loan_id);
CREATE INDEX IF NOT EXISTS idx_tx_created_at ON transactions(created_at);

-- ─── System Settings (singleton) ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS system_settings (
    id         UUID        PRIMARY KEY,
    mode       VARCHAR(20) NOT NULL DEFAULT 'secure'
                           CHECK (mode IN ('secure', 'vulnerable')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO system_settings (id, mode)
VALUES ('00000000-0000-0000-0000-000000000001', 'secure')
ON CONFLICT (id) DO NOTHING;

-- ─── Mode Change Audit Log ────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS mode_change_logs (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    previous_mode VARCHAR(20) NOT NULL,
    new_mode      VARCHAR(20) NOT NULL,
    ip_address    VARCHAR(45),
    user_agent    TEXT,
    changed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mode_change_logs_changed_at
    ON mode_change_logs(changed_at DESC);

-- ─────────────────────────────────────────────────────────────────────────────
-- Seed: default application accounts (all passwords: bcrypt cost 10)
--
-- USERNAME        PASSWORD            ROLE
-- admin           Admin@123           admin
-- staff01         Admin@123           staff
-- budi_santoso    Admin@Dana2025!     admin   (E2E / pentest test account)
-- dewi_rahayu     Staff@Dana2025!     staff   (E2E / pentest test account)
-- ─────────────────────────────────────────────────────────────────────────────

-- bcrypt hash of "Admin@123" at cost 10
INSERT INTO users (username, email, password_hash, role)
VALUES (
    'admin',
    'admin@danasejahtera.id',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'admin'
) ON CONFLICT (username) DO NOTHING;

INSERT INTO users (username, email, password_hash, role)
VALUES (
    'staff01',
    'staff01@danasejahtera.id',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'staff'
) ON CONFLICT (username) DO NOTHING;

-- bcrypt hash of "Admin@Dana2025!" at cost 10
INSERT INTO users (username, email, password_hash, role)
VALUES (
    'budi_santoso',
    'budi_santoso@danasejahtera.id',
    '$2a$10$qooJ1qA/ElATjuIwJmITI.D/EVr8YFJTIVuAxkO4veik2Cw6JDr9i',
    'admin'
) ON CONFLICT (username) DO NOTHING;

-- bcrypt hash of "Staff@Dana2025!" at cost 10
INSERT INTO users (username, email, password_hash, role)
VALUES (
    'dewi_rahayu',
    'dewi_rahayu@danasejahtera.id',
    '$2a$10$TF8lcPCXE10S60lBYF.EROCv6PUbOikbddsZJjmsHgz.w3WlaAiLC',
    'staff'
) ON CONFLICT (username) DO NOTHING;
