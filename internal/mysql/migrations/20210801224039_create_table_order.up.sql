CREATE TABLE  IF NOT EXISTS `orders` (
  `id` varchar(36) COLLATE utf8_unicode_ci NOT NULL,
  `buyer_id` varchar(36) COLLATE utf8_unicode_ci NOT NULL,
  `seller_id` varchar(36) COLLATE utf8_unicode_ci NOT NULL,
  `description` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `source_address` text COLLATE utf8_unicode_ci NOT NULL,
  `destination_address` text COLLATE utf8_unicode_ci NOT NULL,
  `total_price` int COLLATE utf8_unicode_ci NOT NULL,
  `status` tinyint COLLATE utf8_unicode_ci NOT NULL,
  `created_time` datetime DEFAULT NULL,
  `updated_time` datetime DEFAULT NULL,
  `deleted_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
