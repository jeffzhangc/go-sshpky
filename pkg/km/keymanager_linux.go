//go:build linux
// +build linux

package km

import (
	"errors"
	"fmt"
	"sshpky/pkg/utils"
)

var linuxError = errors.New("keychain not supported on Linux")

func getPasswordByType(key string, category string) (string, error) {

	return "", linuxError
}

func savePwdByType(key string, category string, secret string) error {

	return linuxError
}

func GenerateOTP(secret string) (string, error) {
	return utils.GenerateCode(secret)
}

func getHostKey(user, host string) string {
	return fmt.Sprintf("%s@%s", user, host)
}

func SaveMFASecret(user, host, secret string) error {
	return linuxError
}
func GetMFASecret(user, host string) (string, error) {

	return "", linuxError
}

func GetPassword(username, host string) (string, error) {
	return getPasswordByType(getHostKey(username, host), "")
}
func SavePassword(username, host, password string) error {
	return savePwdByType(getHostKey(username, host), "", password)
}
