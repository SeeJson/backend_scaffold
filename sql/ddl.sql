
drop table if exists users;
create table users (
    uid bigint not null auto_increment,
    name  varchar(64) not null default '',

    role   int(8) unsigned not null default 0 comment '角色',
    perm bigint(8) unsigned not null default 0 comment '权限',
    status tinyint(2) unsigned not null default 0 comment '状态',

    pwd_hash varchar(128) not null default '',
    salt     varchar(128) not null default '',

    ct timestamp not null default current_timestamp COMMENT '创建时间',
    ut timestamp not null default current_timestamp on update current_timestamp COMMENT '更新时间',

    primary key (uid),
    unique key (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 collate utf8mb4_bin COMMENT='用户';

drop table if exists tasks;
create table tasks (
    id bigint not null auto_increment,
    uid bigint not null,
    status tinyint(2) unsigned not null default 0 comment '状态',

    type  tinyint not null default 0,
    priority  tinyint not null default 0,
    input  varchar(255) not null default '',
    result varchar(255) not null default '',

    note varchar(255) not null default '',

    ct timestamp not null default current_timestamp COMMENT '创建时间',
    ut timestamp not null default current_timestamp on update current_timestamp COMMENT '更新时间',

    primary key (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 collate utf8mb4_bin COMMENT='任务';


-- 操作日志表
DROP TABLE IF EXISTS operation_logs;
CREATE TABLE `operation_logs` (
    id bigint unsigned NOT NULL AUTO_INCREMENT,

    api varchar(128) NOT NULL,
    operation varchar(32) NOT NULL,

    uid bigint not null default 0,

    value JSON COMMENT '操作要记录的其它内容',

    ct timestamp not null default current_timestamp COMMENT '创建时间',
    ut timestamp not null default current_timestamp on update current_timestamp COMMENT '更新时间',

    PRIMARY KEY(id),
    KEY(uid)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='操作日志';
