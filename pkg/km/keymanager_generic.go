//go:build !darwin

package km

import (
	"fmt"
	"sshpky/pkg/utils"

	"github.com/zalando/go-keyring"
)

const (
	KEY_CHAIN_NAME = "ssh_py_default"
)

const (
	PWD_GOOGLE = "googleAuthCode"
	PWD_NORMAL = "password"
)

func getPasswordByType(key string, category string) (string, error) {
	pwdKey := fmt.Sprintf("%s/%s", key, category)
	val, err := keyring.Get(KEY_CHAIN_NAME, pwdKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", nil
		}
		return "", fmt.Errorf("keyring get error: %v", err)
	}
	return val, nil
}

func savePwdByType(key string, category string, secret string) error {
	pwdKey := fmt.Sprintf("%s/%s", key, category)
	if err := keyring.Set(KEY_CHAIN_NAME, pwdKey, secret); err != nil {
		return fmt.Errorf("keyring set error: %v", err)
	}
	return nil
}

func GenerateOTP(secret string) (string, error) {
	return utils.GenerateCode(secret)
}

func getHostKey(user, host string) string {
	return fmt.Sprintf("%s@%s", user, host)
}

func SaveMFASecret(user, host, secret string) error {
	return savePwdByType(getHostKey(user, host), PWD_GOOGLE, secret)
}
func GetMFASecret(user, host string) (string, error) {
	pwd, err := getPasswordByType(getHostKey(user, host), PWD_GOOGLE)
	if err != nil {
		return pwd, err
	}
	if pwd != "" && len(pwd) > 0 {
		return GenerateOTP(pwd)
	}
	return "", nil
}

func GetPassword(username, host string) (string, error) {
	return getPasswordByType(getHostKey(username, host), PWD_NORMAL)
}
func SavePassword(username, host, password string) error {
	return savePwdByType(getHostKey(username, host), PWD_NORMAL, password)
}
