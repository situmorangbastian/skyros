CREATE TABLE  IF NOT EXISTS `orders_product` (
  `id` varchar(36) COLLATE utf8_unicode_ci NOT NULL,
  `order_id` varchar(36) COLLATE utf8_unicode_ci NOT NULL,
  `product_id` varchar(36) COLLATE utf8_unicode_ci NOT NULL,
  `quantity` int COLLATE utf8_unicode_ci NOT NULL,
  `created_time` datetime DEFAULT NULL,
  `updated_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
