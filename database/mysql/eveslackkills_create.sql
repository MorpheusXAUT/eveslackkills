/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;

-- Dumping database structure for eveslackkills
CREATE DATABASE IF NOT EXISTS `eveslackkills` /*!40100 DEFAULT CHARACTER SET utf8 */;
USE `eveslackkills`;


-- Dumping structure for table eveslackkills.corporations
CREATE TABLE IF NOT EXISTS `corporations` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `evecorporationid` int(11) NOT NULL,
  `lastkillid` int(11) NOT NULL,
  `lastlossid` int(11) NOT NULL,
  `name` varchar(64) NOT NULL,
  `killcomment` varchar(256) NOT NULL,
  `losscomment` varchar(256) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- Data exporting was unselected.


-- Dumping structure for table eveslackkills.ignoredsolarsystems
CREATE TABLE IF NOT EXISTS `ignoredsolarsystems` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `corporationid` int(11) NOT NULL,
  `solarsystemid` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_ignoredregions_corporation` (`corporationid`),
  CONSTRAINT `fk_ignoredregions_corporation` FOREIGN KEY (`corporationid`) REFERENCES `corporations` (`id`) ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- Data exporting was unselected.
/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IF(@OLD_FOREIGN_KEY_CHECKS IS NULL, 1, @OLD_FOREIGN_KEY_CHECKS) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
