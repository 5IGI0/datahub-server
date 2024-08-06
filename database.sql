
CREATE TABLE IF NOT EXISTS `schema_version` (
    `version_id` INT PRIMARY KEY,
    `applied_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO schema_version (version_id)
SELECT 1 WHERE NOT EXISTS (SELECT 1 FROM schema_version);

CREATE TABLE IF NOT EXISTS `individuals` (
    `id`            INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `data`          TEXT CHARACTER SET utf8mb4 NOT NULL DEFAULT '{}',
    `hash_id`       CHAR(40) CHARACTER SET ascii GENERATED ALWAYS AS (SHA1(CONCAT(`data`))),
    `first_seen`    DATETIME NOT NULL,
    `last_seen`     DATETIME NOT NULL,
    INDEX `hash_index`(`hash_id`),
    UNIQUE (`hash_id`)
);

CREATE TABLE IF NOT EXISTS `individual_emails` (
    `id`        BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `email`     TEXT CHARACTER SET utf8mb4 NOT NULL,
    `san_user`  VARCHAR(63) CHARACTER SET ascii GENERATED ALWAYS AS (
        LOWER(REGEXP_REPLACE(
            SUBSTRING_INDEX(SUBSTRING_INDEX(`email`, '@', 1), '+', 1),
            "[^a-zA-Z0-9]", ""))) VIRTUAL,
    `rev_host`  VARCHAR(255) CHARACTER SET ascii GENERATED ALWAYS AS (
        REVERSE(LOWER(SUBSTRING_INDEX(`email`, '@', -1)))) VIRTUAL,
    `individual_id` INT UNSIGNED NOT NULL,
    INDEX `san_user_index`(`san_user`,`rev_host`),
    INDEX `rev_host_ind`(`rev_host`),
    INDEX `individual_ind`(`individual_id`),
    FOREIGN KEY (`individual_id`)
        REFERENCES `individuals`(`id`)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `individual_sources` (
    `id`        BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `source`    VARCHAR(63) NOT NULL,
    `individual_id` INT UNSIGNED NOT NULL,
    INDEX `source_ind`(`source`),
    INDEX `individual_ind`(`individual_id`),
    FOREIGN KEY (`individual_id`)
        REFERENCES `individuals`(`id`)
        ON DELETE CASCADE,
    UNIQUE (`individual_id`, `source`)
);

CREATE TABLE IF NOT EXISTS `domains` (
    `id`            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `domain`        VARCHAR(255) CHARACTER SET ascii NOT NULL UNIQUE KEY,
    `rev_domain`    VARCHAR(255) CHARACTER SET ascii GENERATED ALWAYS AS (REVERSE(`domain`)) VIRTUAL,
    `is_active`     TINYINT NOT NULL DEFAULT 0,
    `cur_flags`     BIGINT UNSIGNED DEFAULT 0,
    `old_flags`     BIGINT UNSIGNED DEFAULT 0,
    `first_seen`    DATETIME DEFAULT NULL,
    `last_seen`     DATETIME DEFAULT NULL,
    `last_check`    DATETIME DEFAULT NULL,
    `check_ver`     SMALLINT UNSIGNED DEFAULT 0,
    INDEX `rev_domain_ind`(`rev_domain`),
    INDEX `last_check_ind`(`last_check`)
);

DELIMITER //
CREATE TRIGGER IF NOT EXISTS domains_insert_trigger
BEFORE INSERT ON `domains` FOR EACH ROW
BEGIN
   -- Remove trailing period from domain
   SET NEW.domain = REGEXP_REPLACE(LOWER(NEW.domain), '\\.$', '');
   SET NEW.old_flags = NEW.cur_flags;
END; //
CREATE TRIGGER IF NOT EXISTS domains_update_trigger
BEFORE UPDATE ON `domains` FOR EACH ROW
BEGIN
   -- Remove trailing period from domain
   SET NEW.domain = REGEXP_REPLACE(LOWER(NEW.domain), '\\.$', '');
   SET NEW.old_flags = OLD.old_flags | NEW.cur_flags;
END; //
DELIMITER ;

CREATE TABLE IF NOT EXISTS `domain_scan_archives` (
    `id`            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `domain_id`     BIGINT NOT NULL,
    `raw_result`    TEXT NOT NULL,
    INDEX `domain_id_ind`(`domain_id`)
)
PAGE_COMPRESSED=1
PAGE_COMPRESSION_LEVEL=9;

CREATE TABLE IF NOT EXISTS `dns_records` (
    `id`            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `domain_id`     BIGINT UNSIGNED NOT NULL,
    `is_active`     TINYINT UNSIGNED NOT NULL,
    `type`          SMALLINT UNSIGNED NOT NULL,
    `addr`          VARCHAR(255),
    `priority`      SMALLINT UNSIGNED,
    `hash_id`       CHAR(40) CHARACTER SET ascii GENERATED ALWAYS AS (SHA1(CONCAT(
                        `domain_id`,":",`type`,":",IFNULL(`addr`,''),":",IFNULL(`priority`,'0')))),
    UNIQUE (`hash_id`),
    INDEX `hash_id_ind`(`hash_id`),
    INDEX `domain_id_ind`(`domain_id`),
    INDEX `addr_ind`(`addr`),
    FOREIGN KEY (`domain_id`)
        REFERENCES `domains`(`id`)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `ssl_certificates` (
    `id`                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `hash_id`           CHAR(40) CHARACTER SET ascii GENERATED ALWAYS AS (SHA1(`certificate`)),
    `certificate`       BLOB NOT NULL,
    `row_ver`           TINYINT UNSIGNED NOT NULL,

    `issuer_rfc4514`    VARCHAR(255) CHARACTER SET utf8mb4 NOT NULL,
    `issuer_name`       VARCHAR(127) CHARACTER SET utf8mb4,
    `issuer_orga`       VARCHAR(127) CHARACTER SET utf8mb4,
    `subject_rfc4514`   VARCHAR(255) CHARACTER SET utf8mb4 NOT NULL,    
    `subject_name`      VARCHAR(127) CHARACTER SET utf8mb4,
    `subject_orga`      VARCHAR(127) CHARACTER SET utf8mb4,
    `valid_before`      DATETIME NOT NULL,
    `valid_after`       DATETIME NOT NULL,
    `public_key`        TEXT NOT NULL,
    INDEX `hash_id_ind`(`hash_id`),
    INDEX `issuer_name_ind`(`issuer_name`),
    INDEX `subject_name_ind`(`subject_name`),
    INDEX `issuer_orga_ind`(`issuer_orga`),
    INDEX `subject_orga_ind`(`subject_orga`),
    UNIQUE (`hash_id`)
)
PAGE_COMPRESSED=1
PAGE_COMPRESSION_LEVEL=9;

CREATE TABLE IF NOT EXISTS `ssl_certificate_dns_names` (
    `id`                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `certificate_id`    BIGINT UNSIGNED NOT NULL,
    `domain`            VARCHAR(255) CHARACTER SET ascii NOT NULL,
    `rev_domain`        VARCHAR(255) CHARACTER SET ascii GENERATED ALWAYS AS (REVERSE(`domain`)) VIRTUAL,
    INDEX `cert_id_ind`(`certificate_id`,`domain`),
    INDEX `domain_ind`(`domain`,`certificate_id`),
    INDEX `rev_domain_ind`(`rev_domain`,`certificate_id`),
    FOREIGN KEY (`certificate_id`)
        REFERENCES `ssl_certificates`(`id`)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `http_services` (
    `id`            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `domain_id`     BIGINT UNSIGNED NOT NULL,
    `is_active`     TINYINT NOT NULL,

    -- identity
    `domain`        varchar(255) NOT NULL,
    `rev_domain`    VARCHAR(255) CHARACTER SET ascii GENERATED ALWAYS AS (REVERSE(`domain`)) VIRTUAL,
    `secure`        TINYINT NOT NULL,
    `port`          SMALLINT NOT NULL,

    -- actual data
    `page_title`        VARCHAR(255) CHARACTER SET utf8mb4 ,
    `status_code`       SMALLINT NOT NULL,
    `actual_path`       VARCHAR(255) NOT NULL,
    `raw_result`        TEXT NOT NULL,
    `certificate_id`    BIGINT UNSIGNED,
    INDEX `domain_id_ind`(`domain_id`),
    INDEX `rev_domain_ind`(`rev_domain`),
    INDEX `page_title_ind`(`page_title`),
    INDEX `certificate_ind`(`certificate_id`),
    UNIQUE (`domain_id`, `secure`, `port`),
    FOREIGN KEY (`domain_id`)
        REFERENCES `domains`(`id`)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `http_document_meta` (
    `id`            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `service_id`    BIGINT UNSIGNED NOT NULL,
    `is_active`     TINYINT         NOT NULL,
    `property`      VARCHAR(127) CHARACTER SET utf8mb4 NOT NULL,
    `content`       VARCHAR(127) CHARACTER SET utf8mb4 NOT NULL,
    `hash_id`       CHAR(40) CHARACTER SET ascii GENERATED ALWAYS AS (SHA1(CONCAT(
                        `service_id`,':',`property`,':',`content`))),
    UNIQUE (`hash_id`),
    INDEX `hash_id_ind`(`hash_id`),
    INDEX `service_id_index`(`service_id`),
    INDEX `property_content_ind`(`property`,`content`),
    FOREIGN KEY (`service_id`)
        REFERENCES `http_services`(`id`)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `http_headers` (
    `id`            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `service_id`    BIGINT UNSIGNED NOT NULL,
    `is_active`     TINYINT UNSIGNED NOT NULL,
    `key`           VARCHAR(127) CHARACTER SET utf8mb4 NOT NULL,
    `value`         VARCHAR(127) CHARACTER SET utf8mb4 NOT NULL,
    `hash_id`       CHAR(40) CHARACTER SET ascii GENERATED ALWAYS AS (SHA1(CONCAT(
                        `service_id`,':',`key`,':',`value`))),
    UNIQUE (`hash_id`),
    INDEX `hash_id_ind`(`hash_id`),
    INDEX `service_id_index`(`service_id`),
    INDEX `key_value_ind`(`key`,`value`),
    FOREIGN KEY (`service_id`)
        REFERENCES `http_services`(`id`)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `http_robots_txt` (
    `id`            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `service_id`    BIGINT UNSIGNED NOT NULL,
    `is_active`     TINYINT UNSIGNED NOT NULL,
    `useragent`     VARCHAR(63)  CHARACTER SET utf8mb4 NOT NULL,
    `directive`     VARCHAR(127) CHARACTER SET utf8mb4 NOT NULL,
    `value`         VARCHAR(512) CHARACTER SET utf8mb4 NOT NULL,
    `hash_id`       CHAR(40) CHARACTER SET ascii GENERATED ALWAYS AS (SHA1(CONCAT(
                        `service_id`,':',`useragent`,':',`directive`,':',`value`))),
    UNIQUE (`hash_id`),
    INDEX `hash_id_ind`(`hash_id`),
    INDEX `service_id_index`(`service_id`),
    INDEX `directive_ind`(`directive`,`value`),
    FOREIGN KEY (`service_id`)
        REFERENCES `http_services`(`id`)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `http_certificate_history`(
    `id`                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `service_id`        BIGINT UNSIGNED NOT NULL,
    `certificate_id`    BIGINT UNSIGNED NOT NULL,
    `observed_at`       DATETIME NOT NULL,
    INDEX `service_ind`(`service_id`,`observed_at`),
    INDEX `certificate_ind`(`certificate_id`,`service_id`),
    FOREIGN KEY (`service_id`)
        REFERENCES `http_services`(`id`)
        ON DELETE CASCADE
);

/*
 * UPDATES
 */

DROP PROCEDURE IF EXISTS apply_schema_changes;
DELIMITER //
CREATE PROCEDURE apply_schema_changes()
BEGIN
    DECLARE current_version INT;
    SELECT MAX(version_id) INTO current_version FROM schema_version;

    -- to apply changes, do like that:
    -- IF current_version < [YOUR_NEW_VERSION] THEN
    --      ...
    --      INSERT INTO schema_version (version_id) VALUES ([YOUR_NEW_VERSION]);
    -- END IF

END //
DELIMITER ;
CALL apply_schema_changes();
DROP PROCEDURE IF EXISTS apply_schema_changes;