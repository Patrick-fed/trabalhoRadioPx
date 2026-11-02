-- RadioPX Database Migration
-- Version: 004_user_last_location
-- Date: 2026-11-02
-- Admin panel (RF09): passive last-known user position. Updated only when the
-- app already sends lat/lng (nearby channels / create channel). No history.

ALTER TABLE users ADD COLUMN IF NOT EXISTS last_lat DOUBLE PRECISION;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_lng DOUBLE PRECISION;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_location_at TIMESTAMP;
