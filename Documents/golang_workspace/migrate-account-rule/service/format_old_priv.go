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

func DoAddAccounts(apps map[string]int64, users []*PrivModule, clusterType string) error {
	testpsw := "xbhESrkOF+ZSKjqHTzvB3KtnQs97oD5hDvfWxt4RksqYfnR/dr2UF3c27hGXJuTBvX4OUSa8FlpuTSuP0ekesASVmIY9LXrILwaRL9hSeFpNAWYJd34b7G372z8EOGjLeQB8FPvOV/2XuVZJd8br3dOsAmVoxwlfRvVrVNqmCAI="
	for _, user := range users {
		account := AccountPara{BkBizId: apps[user.App], User: user.User,
			//	Psw: user.Psw, Operator: "migrate", ClusterType: &tendbcluster}
			Psw: testpsw, Operator: "migrate", ClusterType: &clusterType}
		err := AddAccount(account)
		if err != nil {
			slog.Error("add account error", account, err)
			return err
		}
	}
	return nil
}

func DoAddAccountRule(rule *PrivModule, apps map[string]int64, clusterType string, priv map[string]string) error {
	id, err := GetAccount(AccountPara{BkBizId: apps[rule.App], User: rule.User, ClusterType: &clusterType})
	if err != nil {
		return fmt.Errorf("add rule failed when get account: %s", err.Error())
	}
	//23
	err = AddAccountRule(AccountRulePara{BkBizId: apps[rule.App], ClusterType: &clusterType, AccountId: id,
		Dbname: rule.Dbname, Priv: priv, Operator: "migrate"})
	if err != nil {
		return fmt.Errorf("add rule failed: %s", err.Error())
	}
	return nil
}
