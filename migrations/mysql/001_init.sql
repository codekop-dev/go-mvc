-- MariaDB/MySQL migration for the learning sample.
-- Run from the database selected in DB_NAME.

CREATE TABLE IF NOT EXISTS penjualan (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    no_trx VARCHAR(64) NOT NULL,
    total DECIMAL(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_penjualan_created (created_at)
);

CREATE TABLE IF NOT EXISTS sample_items (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    total DECIMAL(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO penjualan (no_trx, total)
SELECT 'TRX-DEMO-001', 125000
WHERE NOT EXISTS (SELECT 1 FROM penjualan WHERE no_trx = 'TRX-DEMO-001');
INSERT INTO sample_items (name, total)
SELECT 'Item demo', 10000
WHERE NOT EXISTS (SELECT 1 FROM sample_items WHERE name = 'Item demo');
