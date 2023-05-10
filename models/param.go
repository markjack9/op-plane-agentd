package models

type ClientData struct {
	ParameterType string `json:"parametertype" binding:"required"`
	ClientParame  `json:"clientparame"`
}
type ClientParame struct {
	Hostid       int64  `json:"hostid"`
	Hostname     string `json:"hostname"`
	OptionTime   string `json:"optiontime"`
	OptionNote   string `json:"optionnote"`
	OptionIp     string `json:"optionip"`
	OpitonParame string `json:"opitonparame"`
}

type ServerResult struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
