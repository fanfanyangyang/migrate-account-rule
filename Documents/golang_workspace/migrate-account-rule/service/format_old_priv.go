package service

import (
	"fmt"
	"golang.org/x/exp/slog"
	"migrate_account_rule/util"
	"strconv"
	"strings"
)

func FilterMigratePriv(appWhere string, exclude []AppUser) ([]*PrivModule, []string, error) {
	all := make([]*PrivModule, 0)
	need := make([]*PrivModule, 0)
	uids := make([]string, 0)
	vsql := fmt.Sprintf("select uid,app,db_module,module,user,dbname,psw,privileges "+
		" from tb_app_priv_module where app in (%s);", appWhere)
	err := util.DB.Self.Debug().Raw(vsql).Scan(&all).Error
	if err != nil {
		slog.Error(vsql, "execute error", err)
		return need, uids, err
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
			need = append(need, module)
			uids = append(uids, strconv.FormatInt(module.Uid, 10))
		}
	}
	if len(uids) == 0 {
		slog.Warn("no rule should be migrated")
	}
	return need, uids, err
}

func GetUsers(key string, uids []string) ([]*PrivModule, []*PrivModule, error) {
	spiderUsers := make([]*PrivModule, 0)
	users := make([]*PrivModule, 0)

	vsql := fmt.Sprintf("select distinct app,user,AES_DECRYPT(psw,'%s') as psw"+
		" from tb_app_priv_module where uid in (%s) and db_module in ('spider_master','spider_slave');",
		key, strings.Join(uids, ","))
	err := util.DB.Self.Debug().Raw(vsql).Scan(&spiderUsers).Error
	if err != nil {
		slog.Error(vsql, "execute error", err)
		return spiderUsers, users, err
	}
	// todo 8.0解析出的密码是16进制的
	vsql = fmt.Sprintf("select distinct app,user,AES_DECRYPT(psw,'%s') as psw"+
		" from tb_app_priv_module where uid in (%s) and db_module not in ('spider_master','spider_slave');",
		key, strings.Join(uids, ","))
	err = util.DB.Self.Debug().Raw(vsql).Scan(&users).Error
	if err != nil {
		slog.Error(vsql, "execute error", err)
		return spiderUsers, users, err
	}
	return spiderUsers, users, nil
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
	var dml, ddl, global, all []string
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
			all = append(all, p)
		} else {
			return target, fmt.Errorf("privilege: %s not allowed", p)
		}
	}
	target["dml"] = strings.Join(dml, ",")
	target["ddl"] = strings.Join(ddl, ",")
	target["global"] = strings.Join(global, ",")
	target["all"] = strings.Join(all, ",")
	return target, nil
}

func DoAddAccounts(apps map[string]int64, users []*PrivModule, clusterType string) error {
	testpsw := "l7BmcNiE48aMvTLzCkI6nJiCh8TYSxHOLtSqIdRFFv13bS6gCsjt8dYysn4uoawtkxhVlm1PezQbUYtdWRvschahFjcWMhCPnju/SJ9oh1ET6FkIgg2cSFCa7zVJsubuHZbQEj2W2xqDhp+CRUhd5a7hGxsupmWoE8408pfTitg="
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

	/* todo id, err := GetAccount(AccountPara{BkBizId: apps[rule.App], User: rule.User, ClusterType: &clusterType})
	if err != nil {
		return fmt.Errorf("add rule failed when get account: %s", err.Error())
	}

	*/
	err := AddAccountRule(AccountRulePara{BkBizId: apps[rule.App], ClusterType: &clusterType, AccountId: 23,
		Dbname: rule.Dbname, Priv: priv, Operator: "migrate"})
	if err != nil {
		return fmt.Errorf("add rule failed: %s", err.Error())
	}
	return nil
}
