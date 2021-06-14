CREATE TABLE `users` (
  `id` BIGINT PRIMARY KEY
);

CREATE TABLE `messages` (
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `inbox_hash` VARCHAR  (60),
  `created_at` TIMESTAMP,
  `sender_id` BIGINT,
  `text` TEXT ,
  `deleted_sender` TINYINT (1) DEFAULT 0,
  `deleted_receiver` TINYINT (1) DEFAULT 0
);

CREATE TABLE `inboxes` (
  `id` INT PRIMARY KEY AUTO_INCREMENT,
  `user_id` BIGINT,
  `sender_id` BIGINT,
  `inbox_hash` VARCHAR  (60),
  `last_msg` VARCHAR  (60),
  `seen`  TINYINT (1),
  `unseen_nubmber` INT
);

ALTER TABLE `messages` ADD FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`);

ALTER TABLE `inboxes` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `inboxes` ADD FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`);

CREATE INDEX `messages_index_0` ON `messages` (`inbox_hash`);

CREATE INDEX `messages_index_1` ON `messages` (`sender_id`);

CREATE INDEX `messages_index_2` ON `messages` (`inbox_hash`, `sender_id`);

CREATE INDEX `inboxes_index_3` ON `inboxes` (`inbox_hash`);

CREATE INDEX `inboxes_index_4` ON `inboxes` (`sender_id`);

CREATE INDEX `inboxes_index_5` ON `inboxes` (`inbox_hash`, `sender_id`);