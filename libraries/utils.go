package libraries

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// Md5 字符串 MD5 加密
func Md5(content string) string {
	// 转16进制
	md5Byte := md5.Sum([]byte(content))
	return hex.EncodeToString(md5Byte[:])
}

// Sha256 字符串 SHA256 加密，返回二进制结果
func Sha256(content string, secret string) []byte {
	// 创建一个新的 HMAC 使用 sha256
	h := hmac.New(sha256.New, []byte(secret))
	// 写入数据
	h.Write([]byte(content))
	// 计算 HMAC 并返回二进制结果
	return h.Sum(nil)
}

// Nanoid 唯一编号生成器
func Nanoid(l ...int) string {
	id, _ := gonanoid.New(l...)
	return id
}
