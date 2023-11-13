package service

import "migrate_account_rule/util"

// PrivModule scr、gcs账号规则的结构
type PrivModule struct {
	Uid        int64  `json:"uid" gorm:"column:uid"`
	App        string `json:"app" gorm:"column:app"`
	DbModule   string `json:"db_module" gorm:"column:db_module"`
	Module     string `json:"module" gorm:"column:module"`
	User       string `json:"user" gorm:"column:user"`
	Dbname     string `json:"dbname" gorm:"column:dbname"`
	Psw        string `json:"psw" gorm:"column:psw"`
	Privileges string `json:"privileges" gorm:"column:privileges"`
	Comment    string `json:"comment"  gorm:"column:comment"`
}

type Count struct {
	AppUser
	Dbname string `json:"dbname" gorm:"column:dbname"`
	Cnt    int64  `json:"cnt" gorm:"column:cnt"`
}

type AppUser struct {
	App  string `json:"app" gorm:"column:app"`
	User string `json:"user" gorm:"column:user"`
}

// AccountPara GetAccount、AddAccount、ModifyAccountPassword、DeleteAccount函数的入参
type AccountPara struct {
	BkBizId     int64   `json:"bk_biz_id"`
	User        string  `json:"user"`
	Psw         string  `json:"psw"`
	Operator    string  `json:"operator"`
	ClusterType *string `json:"cluster_type" `
	MigrateFlag bool    `json:"migrate_flag"`
}

// AccountRulePara AddAccountRule、ModifyAccountRule、ParaPreCheck函数的入参
type AccountRulePara struct {
	BkBizId     int64   `json:"bk_biz_id"`
	ClusterType *string `json:"cluster_type"`
	AccountId   int64   `json:"account_id"` // account的id
	Dbname      string  `json:"dbname"`
	// key为dml、ddl、global；value为逗号分隔的权限；示例{"dml":"select,update","ddl":"create","global":"REPLICATION SLAVE"}
	Priv     map[string]string `json:"priv"`
	Operator string            `json:"operator"`
}

// TbAccounts 账号表
type TbAccounts struct {
	Id          int64           `gorm:"column:id;primary_key;auto_increment" json:"id"`
	BkBizId     int64           `gorm:"column:bk_biz_id;not_null" json:"bk_biz_id"`
	ClusterType string          `gorm:"column:cluster_type;not_null" json:"cluster_type"`
	User        string          `gorm:"column:user;not_null" json:"user"`
	Psw         string          `gorm:"column:psw;not_null" json:"psw"`
	Creator     string          `gorm:"column:creator;not_null;" json:"creator"`
	CreateTime  util.TimeFormat `gorm:"column:create_time" json:"create_time"`
	Operator    string          `gorm:"column:operator" json:"operator"`
	UpdateTime  util.TimeFormat `gorm:"column:update_time" json:"update_time"`
}
