-- version 0

CREATE TABLE `option`(
	`name` VARCHAR(63) PRIMARY KEY,
	`value` VARCHAR(255) NOT NULL
	) WITHOUT ROWID;

CREATE TABLE `question`(
	`id` INTEGER PRIMARY KEY AUTOINCREMENT,
	`urgent` TINYINT(1) NOT NULL,
	`private` TINYINT(1) NOT NULL,
	`featured` TINYINT(1) NOT NULL,
	`asker` BIGINT(20) NOT NULL,
	`asked_at` DATETIME NOT NULL,
	`content` TEXT(32767) NOT NULL,
	`reply` TEXT(32767) NOT NULL,
	`replier` BIGINT(20) NOT NULL,
	`replied_at` DATETIME NOT NULL,
	`deleted_at` DATETIME NOT NULL
	);

CREATE TABLE `tag`(
	`id` INTEGER PRIMARY KEY AUTOINCREMENT,
	`name` VARCHAR(63) NOT NULL UNIQUE,
	`color` VARCHAR(31) NOT NULL,
	`created_at` DATETIME NOT NULL,
	`created_by` INTEGER NOT NULL
);

CREATE TABLE `user`(
	`id` INTEGER PRIMARY KEY AUTOINCREMENT,
	`wx_open_id` VARCHAR(31) NOT NULL UNIQUE,
	`wx_union_id` VARCHAR(31) NOT NULL UNIQUE,
	`role` TINYINT(2) NOT NULL,
	`nick_name` VARCHAR(127) NOT NULL,
	`avatar` VARCHAR(511) NOT NULL,
	`sign_up_at` DATETIME NOT NULL
);

CREATE TABLE `question_tag`(
	`tag_id` INTEGER NOT NULL,
	`question_id` INTEGER NOT NULL,
	`tagged_at` DATETIME NOT NULL,
	`tagged_by` INTEGER NOT NULL,
	PRIMARY KEY(`tag_id`, `question_id`)
);
