package common

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5Encrypt(str string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(str))
	cipherStr := md5Ctx.Sum(nil)

	result := hex.EncodeToString(cipherStr)

	return result
}
