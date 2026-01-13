package idgen

import (
	"crypto/rand"
	"math/big"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// RandomBase62 生成长度为 n 的 Base62 随机字符串（0-9A-Za-z）。
func RandomBase62(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}
	out := make([]byte, n)
	max := big.NewInt(int64(len(base62Alphabet)))
	for i := 0; i < n; i++ {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = base62Alphabet[v.Int64()]
	}
	return string(out), nil
}
