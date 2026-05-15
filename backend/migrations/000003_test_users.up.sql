-- ─── Seed: test users for E2E and integration tests ─────────────────────────
-- Passwords hashed with bcrypt cost 10.
-- Run this migration once against the test database before executing E2E suites.
--
-- Admin  : budi_santoso  /  Admin@Dana2025!
-- Staff  : dewi_rahayu   /  Staff@Dana2025!

INSERT INTO users (username, email, password_hash, role)
VALUES (
    'budi_santoso',
    'budi_santoso@danasejahtera.id',
    '$2a$10$qooJ1qA/ElATjuIwJmITI.D/EVr8YFJTIVuAxkO4veik2Cw6JDr9i',
    'admin'
) ON CONFLICT (username) DO NOTHING;

INSERT INTO users (username, email, password_hash, role)
VALUES (
    'dewi_rahayu',
    'dewi_rahayu@danasejahtera.id',
    '$2a$10$TF8lcPCXE10S60lBYF.EROCv6PUbOikbddsZJjmsHgz.w3WlaAiLC',
    'staff'
) ON CONFLICT (username) DO NOTHING;
