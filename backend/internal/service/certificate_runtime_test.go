package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadIssuedCertificates(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("CADDY_DATA_DIR", dataDir)
	directory := filepath.Join(dataDir, "caddy", "certificates", "test", "example.com")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	template := x509.Certificate{
		SerialNumber: big.NewInt(42), Subject: pkix.Name{CommonName: "*.example.com"},
		Issuer: pkix.Name{CommonName: "测试 CA"}, DNSNames: []string{"*.example.com"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(90 * 24 * time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	payload := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := os.WriteFile(filepath.Join(directory, "example.crt"), payload, 0o600); err != nil {
		t.Fatal(err)
	}

	certificates, err := LoadIssuedCertificates()
	if err != nil || len(certificates) != 1 {
		t.Fatalf("读取证书失败: %+v, %v", certificates, err)
	}
	matched := MatchIssuedCertificates([]string{"*.example.com"}, certificates)
	if len(matched) != 1 || matched[0].Status != "valid" || matched[0].SerialNumber != "2a" {
		t.Fatalf("证书信息不正确: %+v", matched)
	}
}
