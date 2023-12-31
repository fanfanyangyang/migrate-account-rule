package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
	"migrate_account_rule/service"
	"migrate_account_rule/util"
	"os"
	"strings"
)

const mysql = "mysql"
const tendbcluster = "tendbcluster"

func main() {
	// 数据库初始化
	util.DB.Init()
	defer util.DB.Close()
	apps, key, mode := GetConfigFromEnv()
	exclude := make([]service.AppUser, 0)
	var appWhere string
	for app, _ := range apps {
		appWhere = fmt.Sprintf("%s,'%s'", appWhere, app)
	}
	appWhere = strings.TrimPrefix(appWhere, ",")

	// 检查scr、gcs中的账号规则
	pass := service.CheckOldPriv(key, appWhere, exclude)
	notPassTips := "some check not pass"
	passTips := "all check pass"
	if mode == "check" {
		if !pass {
			slog.Error(notPassTips)
		} else {
			slog.Info(passTips)
		}
		return
	} else if mode == "run" {
		if !pass {
			slog.Error(fmt.Sprintf("%s, do not migrate", notPassTips))
			return
		} else {
			slog.Info(passTips)
		}
	} else if mode == "force-run" {
		if !pass {
			slog.Warn(fmt.Sprintf("%s, but force run", notPassTips))
			slog.Warn("user can not be migrated", "users", exclude)
		} else {
			slog.Info(passTips)
		}
	}

	// 获取需要迁移的scr、gcs中的账号规则
	// db_module为spider_master/spider_slave属于spider的权限规则，
	// 其他不明确的，同时迁移到mysql和spider下
	mysqlUids, uids, err := service.FilterMigratePriv(appWhere, exclude)
	if err != nil {
		slog.Error("FilterMigratePriv", "err", err)
		return
	}

	// 获取需要迁移的权限账号
	mysqlUsers, err := service.GetUsers(key, mysqlUids)
	if err != nil {
		slog.Error("GetUsers", "err", err)
		return
	}

	allUsers, err := service.GetUsers(key, uids)
	if err != nil {
		slog.Error("GetUsers", "err", err)
		return
	}

	// 迁移账号
	err = service.DoAddAccounts(apps, allUsers, tendbcluster)
	if err != nil {
		slog.Error("DoAddAccounts", err)
		return
	}
	err = service.DoAddAccounts(apps, mysqlUsers, mysql)
	if err != nil {
		slog.Error("DoAddAccounts", err)
		return
	}

	// 获取需要迁移的规则
	mysqlRules, err := service.GetRules(mysqlUids)
	if err != nil {
		slog.Error("GetRules", "err", err)
		return
	}

	allRules, err := service.GetRules(uids)
	if err != nil {
		slog.Error("GetRules", "err", err)
		return
	}
	slog.Info("migrate account success")

	// 迁移账号规则
	for _, rule := range mysqlRules {
		priv, errInner := service.FormatPriv(rule.Privileges)
		if errInner != nil {
			slog.Error("format privileges", rule.Privileges, errInner)
			continue
		}
		errInner = service.DoAddAccountRule(rule, apps, "mysql", priv)
		if errInner != nil {
			slog.Error("AddAccountAndRule error", rule, errInner)
		}
	}

	for _, rule := range allRules {
		priv, errInner := service.FormatPriv(rule.Privileges)
		if errInner != nil {
			slog.Error("format privileges", rule.Privileges, errInner)
			continue
		}
		errInner = service.DoAddAccountRule(rule, apps, "tendbcluster", priv)
		if errInner != nil {
			slog.Error("AddAccountAndRule error", rule, errInner)
		}
	}
	slog.Info("migrate account rule success")
}

func GetConfigFromEnv() (map[string]int64, string, string) {
	tips := "环境变量APPS为空，请设置需要迁移的app列表，多个app用逗号间隔，格式如\nAPPS='{\"test\":1, \"test2\":2}',名称区分大小写"
	appsStr := viper.GetString("apps")
	if appsStr == "" {
		slog.Error(tips)
		os.Exit(1)
	}
	key := viper.GetString("key")
	if key == "" {
		slog.Error("环境变量KEY为空，请设置")
		os.Exit(1)
	}
	mode := viper.GetString("mode")
	slog.Info(mode)
	if mode != "check" && mode != "run" && mode != "force-run" {
		slog.Error("环境变量MODE为空，请设置，可选模式\ncheck --- 仅检查不实施\nrun --- 检查并且迁移\nforce-run --- 强制执行\n ")
		os.Exit(1)
	}
	apps, err := util.JsonToMap(appsStr)
	if err != nil {
		slog.Error("环境变量APPS格式错误，格式如'{\"test\":1, \"test2\":2}'", err)
		os.Exit(1)
	}
	if len(apps) == 0 {
		slog.Error(tips)
		os.Exit(1)
	}
	return apps, key, mode
}

func init() {
	viper.BindEnv("db.user", "DB_USER")
	viper.BindEnv("db.password", "DB_PASSWORD")
	viper.BindEnv("db.name", "DB_NAME")
	viper.BindEnv("db.host", "DB_HOST")
	viper.BindEnv("db.port", "DB_PORT")
	viper.BindEnv("apps", "APPS")
	viper.BindEnv("key", "KEY")
	viper.BindEnv("mode", "MODE")
	viper.BindEnv("priv.service", "MYSQL_PRIV_MANAGER_APIGW_DOMAIN")
	viper.BindEnv("debug", "DEBUG")
	viper.BindPFlags(flag.CommandLine)
	InitLog()
}

// InitLog 程序日志初始化
func InitLog() {
	var logLevel = new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)
	var logger *slog.TextHandler
	logger = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel, AddSource: viper.GetBool("debug")})
	slog.SetDefault(slog.New(logger))
}
