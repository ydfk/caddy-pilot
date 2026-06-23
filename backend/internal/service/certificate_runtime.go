package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type IssuedCertificate struct {
	Subjects     []string
	IssuedAt     time.Time
	ExpiresAt    time.Time
	Issuer       string
	SerialNumber string
	Status       string
}

func LoadIssuedCertificates() ([]IssuedCertificate, error) {
	dataDir := environmentValue("CADDY_DATA_DIR", filepath.Join("data", "runtime", "caddy-data"))
	certificates := make([]IssuedCertificate, 0)
	for _, root := range certificateStorageRoots(dataDir) {
		loaded, err := loadCertificatesFromRoot(root)
		if err != nil {
			return nil, err
		}
		certificates = appendUniqueCertificates(certificates, loaded...)
	}
	return certificates, nil
}

func loadCertificatesFromRoot(root string) ([]IssuedCertificate, error) {
	certificates := make([]IssuedCertificate, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".crt") {
			return nil
		}
		certificate, err := parseCertificateFile(path)
		if err != nil {
			return nil
		}
		certificates = append(certificates, certificate)
		return nil
	})
	if os.IsNotExist(err) {
		return certificates, nil
	}
	return certificates, err
}

func certificateStorageRoots(dataDir string) []string {
	return []string{
		filepath.Join(dataDir, "caddy", "certificates"),
		filepath.Join(dataDir, "certificates"),
	}
}

func LoadLiveCertificates(ctx context.Context, domains []string) []IssuedCertificate {
	address := environmentValue("CADDYPILOT_HTTPS_ADDR", "127.0.0.1:443")
	certificates := make([]IssuedCertificate, 0, len(domains))
	for _, domain := range domains {
		certificate, err := loadLiveCertificate(ctx, address, domain)
		if err == nil {
			certificates = appendUniqueCertificates(certificates, certificate)
		}
	}
	return certificates
}

func loadLiveCertificate(ctx context.Context, address, domain string) (IssuedCertificate, error) {
	dialer := &net.Dialer{Timeout: 2 * time.Second}
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return IssuedCertificate{}, err
	}
	_ = connection.SetDeadline(time.Now().Add(2 * time.Second))
	// 这里只读取本机 Caddy 返回的证书元数据，证书归属仍由 subjects 精确匹配确认。
	tlsConnection := tls.Client(connection, &tls.Config{ServerName: domain, InsecureSkipVerify: true})
	defer tlsConnection.Close()
	if err := tlsConnection.HandshakeContext(ctx); err != nil {
		return IssuedCertificate{}, err
	}
	peerCertificates := tlsConnection.ConnectionState().PeerCertificates
	if len(peerCertificates) == 0 {
		return IssuedCertificate{}, fmt.Errorf("TLS 服务未返回证书")
	}
	return issuedCertificateFromX509(peerCertificates[0]), nil
}

func appendUniqueCertificates(certificates []IssuedCertificate, values ...IssuedCertificate) []IssuedCertificate {
	for _, value := range values {
		duplicate := false
		for _, existing := range certificates {
			if existing.SerialNumber == value.SerialNumber && existing.Issuer == value.Issuer {
				duplicate = true
				break
			}
		}
		if !duplicate {
			certificates = append(certificates, value)
		}
	}
	return certificates
}

func MatchIssuedCertificates(subjects []string, certificates []IssuedCertificate) []IssuedCertificate {
	wanted := make(map[string]struct{}, len(subjects))
	for _, subject := range subjects {
		wanted[strings.ToLower(subject)] = struct{}{}
	}
	matched := make([]IssuedCertificate, 0)
	for _, certificate := range certificates {
		for _, subject := range certificate.Subjects {
			if _, ok := wanted[strings.ToLower(subject)]; ok {
				matched = append(matched, certificate)
				break
			}
		}
	}
	return matched
}

func parseCertificateFile(path string) (IssuedCertificate, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return IssuedCertificate{}, err
	}
	block, _ := pem.Decode(payload)
	if block == nil || block.Type != "CERTIFICATE" {
		return IssuedCertificate{}, fmt.Errorf("证书 PEM 无效")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return IssuedCertificate{}, err
	}
	return issuedCertificateFromX509(certificate), nil
}

func issuedCertificateFromX509(certificate *x509.Certificate) IssuedCertificate {
	subjects := certificate.DNSNames
	if len(subjects) == 0 && certificate.Subject.CommonName != "" {
		subjects = []string{certificate.Subject.CommonName}
	}
	status := "valid"
	now := time.Now()
	if now.After(certificate.NotAfter) {
		status = "expired"
	} else if certificate.NotAfter.Sub(now) <= 30*24*time.Hour {
		status = "expiring"
	}
	return IssuedCertificate{
		Subjects: subjects, IssuedAt: certificate.NotBefore, ExpiresAt: certificate.NotAfter,
		Issuer: certificate.Issuer.String(), SerialNumber: certificate.SerialNumber.Text(16), Status: status,
	}
}
