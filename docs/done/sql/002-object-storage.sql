-- Migration for existing dev databases (the init script only runs on first volume init).
-- Apply with: psql "$DSN" -f docs/todo/sql/002-object-storage.sql

BEGIN;

ALTER TABLE images ADD COLUMN IF NOT EXISTS object_key VARCHAR(500);
ALTER TABLE images ADD COLUMN IF NOT EXISTS status VARCHAR(16) NOT NULL DEFAULT 'pending';

-- Existing rows were stored on local disk; mark them ready and derive a key from the filename.
UPDATE images
SET object_key = 'images/' || filename,
    status = 'ready'
WHERE object_key IS NULL;

ALTER TABLE images ALTER COLUMN object_key SET NOT NULL;
ALTER TABLE images DROP COLUMN IF EXISTS url;

DROP INDEX IF EXISTS idx_images_upload_timestamp;
CREATE INDEX IF NOT EXISTS idx_images_upload_timestamp ON images(upload_timestamp DESC) WHERE status = 'ready';
CREATE INDEX IF NOT EXISTS idx_images_status ON images(status);

ALTER TABLE images ADD CONSTRAINT images_object_key_key UNIQUE (object_key);

COMMIT;
