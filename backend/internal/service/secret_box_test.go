package service

import (
	"bytes"
	"testing"
)

func TestSecretBoxRoundTrip(t *testing.T) {
	t.Setenv("CADDYPILOT_SECRET_KEY", "test-secret-key")
	box, err := NewSecretBox()
	if err != nil {
		t.Fatalf("创建加密器失败: %v", err)
	}
	plaintext := []byte(`{"access_key_secret":"secret"}`)
	encrypted, err := box.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if bytes.Contains([]byte(encrypted), []byte("secret")) {
		t.Fatal("密文包含明文")
	}
	decrypted, err := box.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("解密结果不一致: %s", decrypted)
	}
}
