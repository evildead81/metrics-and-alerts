package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

const HashHeaderKey = "HashSHA256"

func Hash(data []byte, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))
	if _, err := h.Write(data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
