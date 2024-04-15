CREATE TABLE `ord_dict` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `key` varchar(100) DEFAULT NULL,
  `value` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni-key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `ord_address` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `address` varchar(100) DEFAULT NULL,
  `tick` varchar(100) DEFAULT NULL,
  `available` varchar(100) DEFAULT NULL,
  `transferable` varchar(100) DEFAULT NULL,
  `block` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni-addr-tick` (`address`,`tick`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `ord_tick` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) DEFAULT NULL,
  `dec` int DEFAULT NULL,
  `supply` varchar(100) DEFAULT NULL,
  `mint_limit` varchar(100) DEFAULT NULL,
  `minted` varchar(100) DEFAULT NULL,
  `deploy_tx` varchar(100) DEFAULT NULL,
  `deploy_pos` int(10) DEFAULT NULL,
  `deploy_by` varchar(100) DEFAULT NULL,
  `deploy_time` int unsigned DEFAULT NULL,
  `finish_mint_time` int unsigned DEFAULT NULL,
  `finish_mint_tx` varchar(100) DEFAULT NULL,
  `block` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni-name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `ord_tx` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `txid` varchar(100) DEFAULT NULL,
  `inscription_id` varchar(100) DEFAULT NULL,
  `op` varchar(100) DEFAULT NULL,
  `tick` varchar(100) DEFAULT NULL,
  `amt` varchar(100) DEFAULT NULL,
  `valid_amt` varchar(100) DEFAULT NULL,
  `from` varchar(100) DEFAULT NULL,
  `to` varchar(100) DEFAULT NULL,
  `sat_offset` varchar(100) DEFAULT NULL,
  `block_height` int unsigned DEFAULT NULL,
  `block_time` int unsigned DEFAULT NULL,
  `pos` int DEFAULT NULL,
  `input_idx` int DEFAULT NULL,
  `output_idx` int DEFAULT NULL,
  `status` int DEFAULT NULL,
  `reason` varchar(255) DEFAULT NULL,
  `meta` varchar(255) DEFAULT NULL,
  `content` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni-tx-op-idx` (`txid`,`op`,`input_idx`),
  KEY `idx-block-pos-input` (`block_height`,`pos`,`input_idx`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;