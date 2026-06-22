package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"go-fiber-starter/pkg/config"
)

const encryptedValuePrefix = "v1:"

type SecretBox struct {
	aead cipher.AEAD
}

func NewSecretBox() (*SecretBox, error) {
	secret := strings.TrimSpace(os.Getenv("CADDYPILOT_SECRET_KEY"))
	if secret == "" {
		secret = config.Current.Jwt.Secret
	}
	if secret == "" {
		return nil, fmt.Errorf("CADDYPILOT_SECRET_KEY 和 JWT_SECRET 不能同时为空")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &SecretBox{aead: aead}, nil
}

func (box *SecretBox) Encrypt(plaintext []byte) (string, error) {
	nonce := make([]byte, box.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	payload := box.aead.Seal(nonce, nonce, plaintext, nil)
	return encryptedValuePrefix + base64.RawStdEncoding.EncodeToString(payload), nil
}

func (box *SecretBox) Decrypt(value string) ([]byte, error) {
	if !strings.HasPrefix(value, encryptedValuePrefix) {
		return nil, fmt.Errorf("不支持的密文版本")
	}
	payload, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(value, encryptedValuePrefix))
	if err != nil {
		return nil, err
	}
	nonceSize := box.aead.NonceSize()
	if len(payload) < nonceSize {
		return nil, fmt.Errorf("密文长度无效")
	}
	return box.aead.Open(nil, payload[:nonceSize], payload[nonceSize:], nil)
}
