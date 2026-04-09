DROP TABLE IF EXISTS `table_constant_config`;
CREATE TABLE `table_constant_config` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Primary key ID',
  `table_name` VARCHAR(64) NOT NULL COMMENT 'Associated database table name',
  `column_name` VARCHAR(64) NOT NULL COMMENT 'Associated table column name',
  `const_name` VARCHAR(64) NOT NULL COMMENT 'Custom name for the constant (string)',
  `value` INT NOT NULL COMMENT 'Constant value (integer)',
  `value_type` TINYINT NOT NULL COMMENT 'Value type: 1=integer, 2=bit',
  `meaning` VARCHAR(255) NOT NULL COMMENT 'Meaning/description of the constant',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_table_column_constname` (`table_name`, `column_name`, `const_name`) COMMENT 'Unique constraint: table + column + const name cannot repeat',
  KEY `idx_table_column` (`table_name`, `column_name`) COMMENT 'Index for fast query by table + column'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Database table constant type configuration table';

