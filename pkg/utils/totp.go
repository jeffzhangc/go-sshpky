package utils

import (
	"time"

	"github.com/pquerna/otp/totp"
)

func GenerateCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}
