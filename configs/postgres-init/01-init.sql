
CREATE TABLE IF NOT EXISTS images (
    id SERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    object_key VARCHAR(500) NOT NULL UNIQUE,
    file_size BIGINT NOT NULL,
    upload_timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    content_type VARCHAR(100) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- index on upload_timestamp for faster retrieval of ready images
CREATE INDEX IF NOT EXISTS idx_images_upload_timestamp ON images(upload_timestamp DESC) WHERE status = 'ready';

-- index on filename for faster lookups
CREATE INDEX IF NOT EXISTS idx_images_filename ON images(filename);

-- index on status for cleanup of stale pending rows
CREATE INDEX IF NOT EXISTS idx_images_status ON images(status);

-- insert sample data (optional, for testing)
-- INSERT INTO images (filename, original_filename, object_key, file_size, content_type, status) 
-- VALUES 
--     ('sample1.jpg', 'sample1.jpg', 'images/sample1.jpg', 1024000, 'image/jpeg', 'ready'),
--     ('sample2.jpg', 'sample2.jpg', 'images/sample2.jpg', 2048000, 'image/jpeg', 'ready');
