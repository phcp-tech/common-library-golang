-- Example schema. Replace with your actual table definitions.
-- sqlc reads this file to understand column types for code generation.

CREATE TABLE IF NOT EXISTS users (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    login      INT             NOT NULL,
    name       VARCHAR(64)     NOT NULL DEFAULT '',
    `group`    VARCHAR(64)     NOT NULL DEFAULT '',
    tag        VARCHAR(64)         NULL,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_login (login)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
