package uuid

import (
	"crypto/md5"
	cryptorand "crypto/rand"
	"fmt"
	"strconv"
	"time"
)

// MakeNonce 生成不超过32个字符的随机nonce字串，保证全局唯一性
// 通过纳秒级时间戳(base36) + crypto/rand随机字节(base62)组合，确保唯一性
func MakeNonce() string {
	now := time.Now().UnixNano()
	rb := make([]byte, 12)
	_, _ = cryptorand.Read(rb)

	// 组合时间戳(36进制) + 随机数(base62)
	combined := strconv.FormatInt(now, 36) + base62Encode(rb)

	// 如果长度超过32，截取前32位
	if len(combined) > 32 {
		return combined[:32]
	}
	return combined
}

// MakeNonceV2 强调随机性，隐藏时间信息
func MakeNonceV2() string {
	rb := make([]byte, 20) // 160位随机数
	cryptorand.Read(rb)
	hashed := md5.Sum(rb)
	return fmt.Sprintf("%x", hashed)[:32]
}

// base62Encode 将字节切片编码为base62字符串(0-9, a-z, A-Z)
func base62Encode(data []byte) string {
	const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// 将字节序列视为大整数，逐位转换
	result := make([]byte, 0, len(data)*2)
	num := make([]byte, len(data))
	copy(num, data)

	for !isZero(num) {
		// 大整数除以62，取余数
		var remainder byte
		for i := 0; i < len(num); i++ {
			val := int(num[i]) + int(remainder)*256
			num[i] = byte(val / 62)
			remainder = byte(val % 62)
		}
		result = append(result, charset[remainder])
	}

	// 反转结果
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// isZero 检查字节切片是否全为0
func isZero(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}
