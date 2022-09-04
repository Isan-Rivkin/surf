package elastic

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LogzAccount struct {
	AccountID              int     `json:"accountId"`
	Email                  string  `json:"email"`
	AccountName            string  `json:"accountName"`
	MaxDailyGB             float64 `json:"maxDailyGB"`
	RetentionDays          int     `json:"retentionDays"`
	IsOwner                bool    `json:"isOwner"`
	Searchable             bool    `json:"searchable"`
	Accessible             bool    `json:"accessible"`
	DocSizeSetting         bool    `json:"docSizeSetting"`
	SharingObjectsAccounts []struct {
		AccountID   int    `json:"accountId"`
		AccountName string `json:"accountName"`
	} `json:"sharingObjectsAccounts"`
	UtilizationSettings struct {
		FrequencyMinutes   int  `json:"frequencyMinutes"`
		UtilizationEnabled bool `json:"utilizationEnabled"`
	} `json:"utilizationSettings"`
}

type LogzAccountsListResponse struct {
	Accounts []LogzAccount
}

type LogzApi interface {
	ListTimeBasedAccounts() (*LogzAccountsListResponse, error)
}

type LogzHttpClient struct {
	token string
	url   string
	path  string
}

func NewLogzHttpClient(url, token string) LogzApi {
	return &LogzHttpClient{
		url:   url,
		path:  "v1/account-management/time-based-accounts",
		token: token,
	}
}
func (lz *LogzHttpClient) ListTimeBasedAccounts() (*LogzAccountsListResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", lz.url, lz.path), nil)

	if err != nil {
		return nil, fmt.Errorf("failed creating get request for list accounts in logz %s", err.Error())
	}

	req.Header.Set(LogzIOTokenHeader, lz.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed executing get req for list accounts in logz %s", err.Error())
	}
	defer resp.Body.Close()

	var accounts []LogzAccount
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("failed decoding json for list accounts in logz %s - %s", err.Error(), resp.Status)
	}

	return &LogzAccountsListResponse{Accounts: accounts}, nil
}
