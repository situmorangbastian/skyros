CREATE TABLE IF NOT EXISTS `product` (
  `id` CHAR(36) COLLATE utf8_unicode_ci NOT NULL,
  `name` VARCHAR(255) COLLATE utf8_unicode_ci NOT NULL,
  `description` VARCHAR(255) COLLATE utf8_unicode_ci NOT NULL,
  `price` BIGINT(30) NOT NULL,
  `seller_id` VARCHAR(255) COLLATE utf8_unicode_ci NOT NULL,
  `created_time` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  `updated_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
  `deleted_time` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  INDEX `idx_seller_id` (`seller_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
