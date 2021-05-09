CREATE TABLE `t_score_total` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `book_id` int(11) NOT NULL DEFAULT '0' COMMENT '图书id',
  `total_count` int(11) NOT NULL DEFAULT '0' COMMENT '实际总人数',
  `total_score` int(11) NOT NULL DEFAULT '0' COMMENT '实际总评分',
  `level_score` int(11) NOT NULL DEFAULT '0' COMMENT 'level分数',
  `create_time` datetime NOT NULL COMMENT '创建时间',
  `update_time` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `ux_book_id` (`book_id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8 COMMENT='总分表'