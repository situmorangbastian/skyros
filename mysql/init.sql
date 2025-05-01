CREATE DATABASE IF NOT EXISTS `userservice`;
CREATE DATABASE IF NOT EXISTS `productservice`;
CREATE DATABASE IF NOT EXISTS `orderservice`;
GRANT ALL ON `userservice`.* TO 'user'@'%';
GRANT ALL ON `productservice`.* TO 'user'@'%';
GRANT ALL ON `orderservice`.* TO 'user'@'%';
