package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
	"migrate_account_rule/util"
	"net/http"
)

func GetAccount(account AccountPara) (int64, error) {
	// ListResponse TODO
	type ListResponse struct {
		Count int64        `json:"count"`
		Items []TbAccounts `json:"items"`
	}
	var resp ListResponse
	c := util.NewClientByHosts(viper.GetString("priv.service"))
	result, err := c.Do(http.MethodPost, "/priv/get_account", account)
	if err != nil {
		slog.Error("/priv/add_account", err)
		return 0, err
	}
	if err := json.Unmarshal(result.Data, &resp); err != nil {
		slog.Error("/priv/get_account", err)
		return 0, err
	}
	if resp.Count == 0 {
		slog.Error("account query nothing return", account)
		return 0, fmt.Errorf("account not found")
	}
	return resp.Items[0].Id, nil
}

func AddAccount(account AccountPara) error {
	c := util.NewClientByHosts(viper.GetString("priv.service"))
	_, err := c.Do(http.MethodPost, "/priv/add_account", account)
	if err != nil {
		slog.Error("/priv/add_account", err)
		return err
	}
	return nil
}

func AddAccountRule(rule AccountRulePara) error {
	c := util.NewClientByHosts(viper.GetString("priv.service"))
	_, err := c.Do(http.MethodPost, "/priv/add_account_rule", rule)
	if err != nil {
		slog.Error("/priv/add_account_rule", err)
		return err
	}
	return nil

}

func DoAddAccounts(apps map[string]int64, users []*PrivModule, clusterType string) error {
	for _, user := range users {
		psw := base64.StdEncoding.EncodeToString([]byte(user.Psw))
		account := AccountPara{BkBizId: apps[user.App], User: user.User,
			Psw: psw, Operator: "migrate", ClusterType: &clusterType, MigrateFlag: true}
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
	err = AddAccountRule(AccountRulePara{BkBizId: apps[rule.App], ClusterType: &clusterType, AccountId: id,
		Dbname: rule.Dbname, Priv: priv, Operator: "migrate"})
	if err != nil {
		return fmt.Errorf("add rule failed: %s", err.Error())
	}
	return nil
}
