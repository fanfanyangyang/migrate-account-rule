package service

import (
	"fmt"
	"golang.org/x/exp/slog"
	"migrate_account_rule/util"
	"strconv"
	"strings"
)

func FilterMigratePriv(appWhere string, exclude []AppUser) ([]string, []string, error) {
	all := make([]*PrivModule, 0)
	uids := make([]string, 0)
	mysqlUids := make([]string, 0)
	vsql := fmt.Sprintf("select uid,app,db_module,user "+
		" from tb_app_priv_module where app in (%s);", appWhere)
	err := util.DB.Self.Debug().Raw(vsql).Scan(&all).Error
	if err != nil {
		slog.Error(vsql, "execute error", err)
		return mysqlUids, uids, err
	}

	for _, module := range all {
		var excludeFlag bool
		for _, v := range exclude {
			if module.App == v.App && module.User == v.User {
				excludeFlag = true
				break
			}
		}
		if excludeFlag == false {
			suid := strconv.FormatInt(module.Uid, 10)
			uids = append(uids, suid)
			if module.DbModule != "spider_master" && module.DbModule != "spider_slave" {
				mysqlUids = append(mysqlUids, suid)
			}
		}
	}
	if len(uids) == 0 {
		slog.Warn("no rule should be migrated")
	}
	return mysqlUids, uids, err
}

func GetUsers(key string, uids []string) ([]*PrivModule, error) {
	users := make([]*PrivModule, 0)
	vsql := fmt.Sprintf("select distinct app,user,AES_DECRYPT(psw,'%s') as psw"+
		" from tb_app_priv_module where uid in (%s);",
		key, strings.Join(uids, ","))
	err := util.DB.Self.Debug().Raw(vsql).Scan(&users).Error
	if err != nil {
		slog.Error(vsql, "execute error", err)
		return users, err
	}
	// todo 8.0解析出的密码是16进制的
	return users, nil
}

func GetRules(uids []string) ([]*PrivModule, error) {
	users := make([]*PrivModule, 0)
	vsql := fmt.Sprintf("select distinct app,user,privileges,dbname "+
		" from tb_app_priv_module where uid in (%s);", strings.Join(uids, ","))
	err := util.DB.Self.Debug().Raw(vsql).Scan(&users).Error
	if err != nil {
		slog.Error(vsql, "execute error", err)
		return users, err
	}
	return users, nil
}

// FormatPriv
// SELECT,INSERT,UPDATE,DELETE 转换为 {"dml":"select,update","ddl":"create","global":"REPLICATION SLAVE"}
func FormatPriv(source string) (map[string]string, error) {
	target := make(map[string]string, 4)
	if source == "" {
		return target, fmt.Errorf("privilege is null")
	}
	source = strings.ToLower(source)
	privs := strings.Split(source, ",")
	var dml, ddl, global []string
	var allPrivileges bool
	for _, p := range privs {
		p = strings.TrimPrefix(p, " ")
		p = strings.TrimSuffix(p, " ")
		if p == "select" || p == "insert" || p == "update" || p == "delete" {
			dml = append(dml, p)
		} else if p == "create" || p == "alter" || p == "drop" || p == "index" || p == "execute" || p == "create view" {
			ddl = append(ddl, p)
		} else if p == "file" || p == "trigger" || p == "event" || p == "create routine" || p == "alter routine" ||
			p == "replication client" || p == "replication slave" {
			global = append(global, p)
		} else if p == "all privileges" {
			global = append(global, p)
			allPrivileges = true
		} else {
			return target, fmt.Errorf("privilege: %s not allowed", p)
		}
	}
	if allPrivileges && (len(global) > 1 || len(dml) > 0 || len(ddl) > 0) {
		return target, fmt.Errorf("[all privileges] should not be granted with others")
	}
	target["dml"] = strings.Join(dml, ",")
	target["ddl"] = strings.Join(ddl, ",")
	target["global"] = strings.Join(global, ",")
	return target, nil
}
