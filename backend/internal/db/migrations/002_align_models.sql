-- RadioPX Database Migration
-- Version: 002_align_models
-- Date: 2026-10-20
-- Aligns the schema with the models actually used by the backend:
-- users are registered without CPF today and carry OAuth/activation fields;
-- channels carry description, geographic coordinates and ownership.

-- users: CPF is documented (RF01) but not collected by the app yet
ALTER TABLE users ALTER COLUMN cpf DROP NOT NULL;

ALTER TABLE users ADD COLUMN IF NOT EXISTS provider VARCHAR(50) NOT NULL DEFAULT 'local';
ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(255) NOT NULL DEFAULT '';

-- channels: geo data feeds proximity discovery; created_by/updated_at feed ownership rules
ALTER TABLE channels ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE channels ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_channels_location ON channels(latitude, longitude);
