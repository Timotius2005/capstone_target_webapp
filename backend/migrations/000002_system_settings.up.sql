-- PT. Dana Sejahtera — System Settings & Mode Change Audit Log
-- Adds persistent mode storage and change history for the public mode-switch API.

-- ─── System Settings (singleton) ──────────────────────────────────────────────
-- One row with a fixed UUID acts as the single global configuration record.
-- Application code always upserts on id = SystemSettingsID.
CREATE TABLE IF NOT EXISTS system_settings (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    mode       VARCHAR(20) NOT NULL DEFAULT 'secure'
                           CHECK (mode IN ('secure', 'vulnerable')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed: insert the singleton row only if the table is empty.
INSERT INTO system_settings (id, mode)
VALUES ('00000000-0000-0000-0000-000000000001', 'secure')
ON CONFLICT (id) DO NOTHING;

-- ─── Mode Change Audit Log ─────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS mode_change_logs (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    previous_mode VARCHAR(20) NOT NULL,
    new_mode      VARCHAR(20) NOT NULL,
    ip_address    VARCHAR(45),
    user_agent    TEXT,
    changed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mode_change_logs_changed_at ON mode_change_logs(changed_at DESC);
