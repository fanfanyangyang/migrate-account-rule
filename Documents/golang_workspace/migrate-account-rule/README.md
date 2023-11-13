mysql/tendbcluster权限规则从gcs/scr迁移到blueking-dbm
====
服务启动环境变量
----
MYSQL_PRIV_MANAGER_APIGW_DOMAIN
DB_HOST
DB_PORT
DB_USER
DB_NAME
DB_PASSWORD
MODE
APPS

环境变量使用
----
MYSQL_PRIV_MANAGER_APIGW_DOMAIN：权限服务
DB_HOST: 源db
MODE：执行模式,建议先执行check，解决gcs、scr无法迁移的权限规则， 再run
* check：仅检查，输出检查日志
* run：检查，如果存在不可迁移的user、权限规则，不迁移
* force-run，检查，如果存在不可迁移的user、权限规则，过滤不可迁移的，迁移其他user和权限规则。

APPS：需要迁移的业务，gcs业务app与dbm中bk_biz_id对应关系
{"tesapp":14}
