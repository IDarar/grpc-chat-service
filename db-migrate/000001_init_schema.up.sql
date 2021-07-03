CREATE TABLE `users` (
  `id` BIGINT PRIMARY KEY
);

CREATE TABLE `messages` (
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `inbox_hash` VARCHAR  (60),
  `created_at` TIMESTAMP,
  `sender_id` BIGINT,
  `text` TEXT ,
  `related_data` TINYINT (1) DEFAULT 0,
  `deleted_sender` TINYINT (1) DEFAULT 0,
  `deleted_receiver` TINYINT (1) DEFAULT 0
);

CREATE TABLE `images`(
  `id` VARCHAR (32) PRIMARY KEY,
  `message_id` INT,
  `created_at` TIMESTAMP,
  `inbox_hash` VARCHAR (60),
  `sended_at` TIMESTAMP,
  `sender_id` BIGINT
);

CREATE TABLE `inboxes` (
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `user_id` BIGINT,
  `sender_id` BIGINT,
  `inbox_hash` VARCHAR (60),
  `last_msg` VARCHAR (60),
  `seen` TINYINT (1),
  `unseen_number` INT
);

ALTER TABLE `messages` ADD FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`);

ALTER TABLE `inboxes` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `inboxes` ADD FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`);

ALTER TABLE `images` ADD FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`);

ALTER TABLE `images` ADD FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`);

CREATE INDEX `messages_index_0` ON `messages` (`inbox_hash`);

CREATE INDEX `inboxes_index_1` ON `inboxes` (`inbox_hash`);

CREATE INDEX `inboxes_index_2` ON `inboxes` (`sender_id`);

CREATE INDEX `images_index_0` ON `images` (`inbox_hash`);

CREATE INDEX `images_index_1` ON `images` (`message_id`);

INSERT INTO messages (text) VALUES ("some text");
SELECT LAST_INSERT_ID();