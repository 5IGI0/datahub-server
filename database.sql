
CREATE TABLE IF NOT EXISTS `schema_version` (
    `version_id` INT PRIMARY KEY,
    `applied_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO schema_version (version_id)
SELECT 1 WHERE NOT EXISTS (SELECT 1 FROM schema_version);

CREATE TABLE IF NOT EXISTS `individuals` (
    `id`            INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `data`          TEXT CHARACTER SET utf8 NOT NULL DEFAULT '{}',
    `hash_id`       CHAR(40) CHARACTER SET ascii GENERATED ALWAYS AS (SHA1(CONCAT(`data`))),
    `first_seen`    DATETIME NOT NULL,
    `last_seen`     DATETIME NOT NULL,
    INDEX `hash_index`(`hash_id`),
    UNIQUE (`hash_id`)
);

CREATE TABLE IF NOT EXISTS `individual_emails` (
    `id`        BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `email`     TEXT CHARACTER SET utf8 NOT NULL,
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