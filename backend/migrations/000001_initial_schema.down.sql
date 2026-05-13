DROP INDEX IF EXISTS idx_tx_created_at;
DROP INDEX IF EXISTS idx_tx_loan_id;
DROP TABLE IF EXISTS transactions;

DROP INDEX IF EXISTS idx_loans_created_at;
DROP INDEX IF EXISTS idx_loans_status;
DROP INDEX IF EXISTS idx_loans_nasabah_id;
DROP TABLE IF EXISTS loans;

DROP INDEX IF EXISTS idx_nasabah_deleted;
DROP INDEX IF EXISTS idx_nasabah_nik;
DROP INDEX IF EXISTS idx_nasabah_user_id;
DROP TABLE IF EXISTS nasabah;

DROP INDEX IF EXISTS idx_users_deleted;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;