-- version 0

CREATE TABLE `options`(
	`name` VARCHAR(63) PRIMARY KEY,
	`value` VARCHAR(255) NOT NULL
	) WITHOUT ROWID;

CREATE TABLE `question`(
	`id` INTEGER PRIMARY KEY AUTOINCREMENT,
	`status` SMALLINT(8) NOT NULL,
	`asked_by` BIGINT(20) NOT NULL,
	`asked_at` DATETIME NOT NULL,
	`content` TEXT(32767) NOT NULL,
	`answer` TEXT(32767) NOT NULL,
	`answered_by` BIGINT(20) NOT NULL,
	`answered_at` DATETIME NOT NULL,
	`deleted_at` DATETIME NOT NULL
	);

INSERT INTO `options`(`name`, `value`) VALUES('schema_version', '0');