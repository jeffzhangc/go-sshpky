package config

import (
	"sshpky/pkg/km"
)

type KeyChainManage struct {
}

var keyChainM IKeyM

func init() {
	keyChainM = &KeyChainManage{}
}

func NewKeyChainManage() IKeyM {
	return keyChainM
}

func (k *KeyChainManage) SavePwd(connConf *SshConfigItem) {
	if connConf.Password != "" {
		km.SavePassword(connConf.User, connConf.Host, connConf.Password)
		connConf.Password = ""
	}
	if connConf.MFASecret != "" {
		km.SaveMFASecret(connConf.User, connConf.Host, connConf.MFASecret)
		connConf.MFASecret = ""
	}
}

// GetPwd 从系统密钥链中获取密码
func (k *KeyChainManage) GetPwd(connConf SshConfigItem) string {
	res, err := km.GetPassword(connConf.User, connConf.Host)
	if res == "" || err != nil {
		res, _ = km.GetPassword(connConf.User, connConf.HostName)
	}
	return res
}

// GetMAFSecret 从系统密钥链中获取 mfa 密码
func (k *KeyChainManage) GetMAFSecret(connConf SshConfigItem) string {
	res, err := km.GetMFASecret(connConf.User, connConf.Host)
	if err != nil {
		res, _ = km.GetMFASecret(connConf.User, connConf.HostName)
	}
	return res
}
