package service

import (
	"fmt"
	"golang.org/x/exp/slog"
	"migrate_account_rule/util"
	"strings"
)

func CheckOldPriv(key, appWhere string, exclude []AppUser) bool {
	err1 := CheckDifferentPasswordsForOneUser(key, appWhere, exclude)
	err2 := CheckEmptyPassword(key, appWhere, exclude)
	err3 := CheckPasswordConsistentWithUser(key, appWhere, exclude)
	err4 := CheckDifferentPrivileges(appWhere, exclude)
	err5 := CheckPrivilegesFormat(appWhere, exclude)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return false
	}
	return true
}

func CheckDifferentPasswordsForOneUser(key, appWhere string, exclude []AppUser) error {
	slog.Info("check 1: different passwords for one user")
	count := make([]*Count, 0)
	vsql := fmt.Sprintf("select app,user,count(distinct(AES_DECRYPT(psw,'%s'))) as cnt "+
		" from tb_app_priv_module where app in (%s) group by app,user order by 1,2", key, appWhere)
	err := util.DB.Self.Debug().Raw(vsql).Scan(&count).Error
	if err != nil {
		slog.Error(vsql, "execute error", err)
		return err
	}

	check1 := make([]string, 0)
	for _, distinct := range count {
		if distinct.Cnt > 1 {
			check1 = append(check1,
				fmt.Sprintf("%s    %s     %d",
					distinct.App, distinct.User, distinct.Cnt))
			exclude = append(exclude, AppUser{distinct.App, distinct.User})
		}
	}
	if len(check1) > 0 {
		msg := "app:    user:     different_passwords_count:"
		msg = fmt.Sprintf("\n%s\n%s", msg, strings.Join(check1, "\n"))
		slog.Error(msg)
		slog.Error("[ check 1 Fail ]")
		return fmt.Errorf("different passwords for one user")
	} else {
		slog.Info("[ check 1 Success ]")
	}
	return nil
}

func CheckEmptyPassword(key, appWhere string, exclude []AppUser) error {
	vsql := fmt.Sprintf("select distinct app,user "+
		" from tb_app_priv_module where app in (%s) and psw=AES_ENCRYPT('','%s');", appWhere, key)
	slog.Info("check 2: empty password")
	err := CheckPassword(vsql, exclude, 2)
	if err != nil {
		slog.Error("CheckPassword", "error", err)
		return err
	}
	return nil
}

func CheckPasswordConsistentWithUser(key, appWhere string, exclude []AppUser) error {
	vsql := fmt.Sprintf("select distinct app,user "+
		" from tb_app_priv_module where app in (%s) and user=AES_DECRYPT(psw,'%s');", appWhere, key)
	slog.Info("check 3: password consistent with user")
	err := CheckPassword(vsql, exclude, 3)
	if err != nil {
		slog.Error("CheckPassword", "error", err)
		return err
	}
	return nil
}

func CheckPassword(vsql string, exclude []AppUser, round int) error {
	nopsw := make([]*PrivModule, 0)
	err := util.DB.Self.Debug().Raw(vsql).Scan(&nopsw).Error
	if err != nil {
		slog.Error(vsql, "execute error", err)
		return err
	}
	if len(nopsw) > 0 {
		msg := ""
		for _, user := range nopsw {
			exclude = append(exclude, AppUser{user.App, user.User})
			msg = fmt.Sprintf("%s    %s", user.App, user.User)
		}
		msg = fmt.Sprintf("app:    user: \n%s", msg)
		slog.Error(msg)
		slog.Error(fmt.Sprintf("[ check %d Fail ]", round))
		return fmt.Errorf("password check fail")
	} else {
		slog.Info(fmt.Sprintf("[ check %d Success ]", round))
	}
	return nil
}

func CheckDifferentPrivileges(appWhere string, exclude []AppUser) error {
	slog.Info("check 4: different privileges for [app user dbname]")
	vsql := fmt.Sprintf("select app,user,dbname,count(distinct(privileges)) as cnt "+
		" from tb_app_priv_module where app in (%s) group by app,user,dbname order by 1,2,3", appWhere)
	count := make([]*Count, 0)
	err := util.DB.Self.Debug().Raw(vsql).Scan(&count).Error
	if err != nil {
		slog.Error(vsql, "execute error", err)
		return err
	}
	check := make([]string, 0)
	for _, distinct := range count {
		if distinct.Cnt > 1 {
			check = append(check,
				fmt.Sprintf("%s    %s     %s     %d",
					distinct.App, distinct.User, distinct.Dbname, distinct.Cnt))
			exclude = append(exclude, AppUser{distinct.App, distinct.User})
		}
	}
	if len(check) > 0 {
		msg := "app:    user:     dbname:     different_privileges_count:"
		msg = fmt.Sprintf("\n%s\n%s", msg, strings.Join(check, "\n"))
		slog.Error(msg)
		slog.Error("[ check 4 Fail ]")
		return fmt.Errorf("different privileges")
	} else {
		slog.Info("[ check 4 Success ]")
	}
	return nil
}

func CheckPrivilegesFormat(appWhere string, exclude []AppUser) error {
	UniqMap := make(map[string]struct{})
	privPass := true
	slog.Info("check 5: check privileges")
	vsql := fmt.Sprintf("select uid,app,user,privileges "+
		" from tb_app_priv_module where app in (%s)", appWhere)
	rules := make([]*PrivModule, 0)
	err := util.DB.Self.Debug().Raw(vsql).Scan(&rules).Error
	if err != nil {
		slog.Error(vsql, "execute error", err)
		return err
	}
	for _, rule := range rules {
		_, err = FormatPriv(rule.Privileges)
		if err != nil {
			privPass = false
			slog.Error("msg", "uid", rule.Uid, "app", rule.App, "user", rule.User, "privileges", rule.Privileges, "error", err)
		}
		s := fmt.Sprintf("%s|%s", rule.App, rule.User)
		if _, isExists := UniqMap[s]; isExists == true {
			continue
		}
		UniqMap[s] = struct{}{}
		exclude = append(exclude, AppUser{rule.App, rule.User})
	}
	if !privPass {
		slog.Error("[ check 5 Fail ]")
		return fmt.Errorf("wrong privileges")
	} else {
		slog.Info("[ check 5 Success ]")
	}
	return nil
}
