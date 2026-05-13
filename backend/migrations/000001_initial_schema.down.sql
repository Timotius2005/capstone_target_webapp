-- +goose Down
DROP INDEX IF EXISTS idx_transactions_loan_id;
DROP INDEX IF EXISTS idx_loans_status;
DROP INDEX IF EXISTS idx_loans_nasabah_id;
DROP INDEX IF EXISTS idx_nasabah_nik;
DROP INDEX IF EXISTS idx_nasabah_user_id;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;

DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS loans;
DROP TABLE IF EXISTS nasabah;
DROP TABLE IF EXISTS users;