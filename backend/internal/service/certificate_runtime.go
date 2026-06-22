package service

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
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
	root := filepath.Join(dataDir, "caddy", "certificates")
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
	status := "valid"
	now := time.Now()
	if now.After(certificate.NotAfter) {
		status = "expired"
	} else if certificate.NotAfter.Sub(now) <= 30*24*time.Hour {
		status = "expiring"
	}
	return IssuedCertificate{
		Subjects: certificate.DNSNames, IssuedAt: certificate.NotBefore, ExpiresAt: certificate.NotAfter,
		Issuer: certificate.Issuer.String(), SerialNumber: certificate.SerialNumber.Text(16), Status: status,
	}, nil
}
