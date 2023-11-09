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
	err3 := CheckDifferentPrivileges(appWhere, exclude)
	if err1 != nil || err2 != nil || err3 != nil {
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
	slog.Info("check 2: empty password")
	nopsw := make([]*PrivModule, 0)
	vsql := fmt.Sprintf("select distinct app,user "+
		" from tb_app_priv_module where app in (%s) and psw=AES_ENCRYPT('','%s');", appWhere, key)
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
		slog.Error("[ check 2 Fail ]")
		return fmt.Errorf("empty password")
	} else {
		slog.Info("[ check 2 Success ]")
	}
	return nil
}

func CheckDifferentPrivileges(appWhere string, exclude []AppUser) error {
	slog.Info("check 3: different privileges for [app user dbname]")
	vsql := fmt.Sprintf("select app,user,dbname,count(distinct(privileges)) as cnt "+
		" from tb_app_priv_module where app in (%s) group by app,user,dbname order by 1,2,3", appWhere)
	count := make([]*Count, 0)
	err := util.DB.Self.Debug().Raw(vsql).Scan(&count).Error
	if err != nil {
		return err
		slog.Error(vsql, "execute error", err)
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
		slog.Error("[ check 3 Fail ]")
		return fmt.Errorf("different privileges")
	} else {
		slog.Info("[ check 3 Success ]")
	}
	return nil
}
