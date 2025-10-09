//go:build windows
// +build windows

package km

import (
	"errors"
	"fmt"
	"sshpky/pkg/utils"
)

var windowError = errors.New("keychain not supported on window")

func getPasswordByType(key string, category string) (string, error) {

	return "", windowError
}

func savePwdByType(key string, category string, secret string) error {

	return windowError
}

func GenerateOTP(secret string) (string, error) {
	return utils.GenerateCode(secret)
}

func getHostKey(user, host string) string {
	return fmt.Sprintf("%s@%s", user, host)
}

func SaveMFASecret(user, host, secret string) error {
	return windowError
}
func GetMFASecret(user, host string) (string, error) {

	return "", windowError
}

func GetPassword(username, host string) (string, error) {
	return getPasswordByType(getHostKey(username, host), "")
}
func SavePassword(username, host, password string) error {
	return windowError
}
