-- RadioPX Database Migration
-- Version: 003_admin_users
-- Date: 2026-10-27
-- Admin panel (RF09): users gain an admin flag bootstrapped from ADMIN_EMAILS.

ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_users_admin ON users(is_admin) WHERE is_admin;
