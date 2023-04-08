/*
SQLyog Ultimate v12.4.1 (64 bit)
MySQL - 10.6.5-MariaDB-log : Database - lscan
*********************************************************************
*/

/*!40101 SET NAMES utf8 */;

/*!40101 SET SQL_MODE=''*/;

/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
CREATE DATABASE /*!32312 IF NOT EXISTS*/`lscan` /*!40100 DEFAULT CHARACTER SET utf8mb3 COLLATE utf8mb3_unicode_ci */;

USE `lscan`;

/*Table structure for table `segments` */

DROP TABLE IF EXISTS `segments`;

CREATE TABLE `segments` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `base` varchar(10) COLLATE utf8mb3_unicode_ci DEFAULT NULL,
  `seg` varchar(10) COLLATE utf8mb3_unicode_ci DEFAULT NULL,
  `type` varchar(1) COLLATE utf8mb3_unicode_ci DEFAULT NULL,
  `tsize` varchar(10) COLLATE utf8mb3_unicode_ci DEFAULT NULL,
  `status` varchar(1) COLLATE utf8mb3_unicode_ci DEFAULT NULL,
  `flags` varchar(10) COLLATE utf8mb3_unicode_ci DEFAULT NULL,
  `membox` varchar(10) COLLATE utf8mb3_unicode_ci DEFAULT NULL,
  `period` int(11) DEFAULT NULL,
  `macro` int(11) DEFAULT NULL,
  `size` int(11) DEFAULT NULL,
  `files` int(11) DEFAULT NULL,
  `procent` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fp` (`status`,`period`)
) ENGINE=InnoDB AUTO_INCREMENT=8863 DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_unicode_ci;

/*Data for the table `segments` */

/*Table structure for table `segmentsperiod` */

DROP TABLE IF EXISTS `segmentsperiod`;

CREATE TABLE `segmentsperiod` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `date` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=25 DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_unicode_ci;

/*Data for the table `segmentsperiod` */


/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
