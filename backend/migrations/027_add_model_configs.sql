-- 027_add_model_configs.sql
-- Store admin-controlled model enablement. Secrets and pricing stay in env/code.

CREATE TABLE IF NOT EXISTS model_configs (
    model_id VARCHAR(100) NOT NULL PRIMARY KEY COMMENT 'Provider model ID',
    enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT 'Whether the model is visible/selectable',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update time',
    INDEX idx_model_configs_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Admin model enablement config';

INSERT INTO model_configs (model_id, enabled)
VALUES
    ('gemini-3-pro-image-preview', TRUE),
    ('gemini-3.1-flash-image-preview', TRUE),
    ('doubao-seedream-4-5', TRUE),
    ('gpt-image-2', TRUE)
ON DUPLICATE KEY UPDATE model_id = VALUES(model_id);
