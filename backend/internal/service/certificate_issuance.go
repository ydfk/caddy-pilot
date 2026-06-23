package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CertificateIssuanceStatus struct {
	State     string
	Message   string
	LastError string
}

type CertificateRuntimeError struct {
	Timestamp time.Time
	Payload   string
	Message   string
}

func EvaluateCertificateIssuance(
	subjects []string,
	issued []IssuedCertificate,
	usageCount int,
	activeUsageCount int,
	runtimeConfig []byte,
	errors []CertificateRuntimeError,
) CertificateIssuanceStatus {
	if usageCount == 0 {
		return CertificateIssuanceStatus{State: "unused", Message: "未被站点使用；需由启用站点引用并成功发布后才会触发签发"}
	}
	if activeUsageCount == 0 {
		return CertificateIssuanceStatus{State: "pending_publish", Message: "引用该证书的站点尚未启用"}
	}
	if len(issued) > 0 {
		state := "issued"
		message := "证书已签发"
		for _, certificate := range issued {
			if certificate.Status == "expired" {
				return CertificateIssuanceStatus{State: "expired", Message: "证书已过期"}
			}
			if certificate.Status == "expiring" {
				state, message = "expiring", "证书即将到期，Caddy 将按策略续期"
			}
		}
		return CertificateIssuanceStatus{State: state, Message: message}
	}
	if !RuntimeConfigContainsSubjects(runtimeConfig, subjects) {
		return CertificateIssuanceStatus{State: "pending_publish", Message: "站点修改尚未发布到当前 Caddy 配置"}
	}
	if lastError := MatchCertificateError(subjects, errors); lastError != "" {
		return CertificateIssuanceStatus{State: "failed", Message: "Caddy 最近一次签发尝试失败", LastError: lastError}
	}
	return CertificateIssuanceStatus{State: "issuing", Message: "配置已发布，Caddy 正在申请或等待签发证书"}
}

func RuntimeConfigContainsSubjects(payload []byte, subjects []string) bool {
	if len(payload) == 0 || len(subjects) == 0 {
		return false
	}
	for _, subject := range subjects {
		encoded, _ := json.Marshal(subject)
		if !bytes.Contains(payload, encoded) {
			return false
		}
	}
	return true
}

func LoadRecentCertificateErrors() []CertificateRuntimeError {
	path := filepath.Join(environmentValue("CADDYPILOT_LOG_DIR", filepath.Join("data", "logs")), "caddy.log")
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	cutoff := time.Now().Add(-30 * time.Minute)
	errors := make([]CertificateRuntimeError, 0)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1<<20)
	for scanner.Scan() {
		var payload map[string]any
		line := scanner.Text()
		if json.Unmarshal([]byte(line), &payload) != nil || !isErrorLevel(payload["level"]) || !isCertificateRuntimeError(payload) {
			continue
		}
		timestamp := parseCaddyLogTime(payload["ts"])
		if !timestamp.IsZero() && timestamp.Before(cutoff) {
			continue
		}
		message := certificateErrorMessage(payload)
		errors = append(errors, CertificateRuntimeError{Timestamp: timestamp, Payload: strings.ToLower(line), Message: message})
	}
	return errors
}

func isCertificateRuntimeError(payload map[string]any) bool {
	loggerName, _ := payload["logger"].(string)
	message, _ := payload["msg"].(string)
	errorMessage, _ := payload["error"].(string)
	loggerName = strings.ToLower(loggerName)
	for _, prefix := range []string{"tls.obtain", "tls.issuance", "tls.renew"} {
		if strings.HasPrefix(loggerName, prefix) {
			return true
		}
	}
	content := strings.ToLower(message + " " + errorMessage)
	for _, fragment := range []string{"acme", "certmagic", "obtaining certificate", "certificate issuance", "certificate renewal", "challenge failed"} {
		if strings.Contains(content, fragment) {
			return true
		}
	}
	return false
}

func certificateErrorMessage(payload map[string]any) string {
	if message, _ := payload["error"].(string); strings.TrimSpace(message) != "" {
		return message
	}
	message, _ := payload["msg"].(string)
	return message
}

func MatchCertificateError(subjects []string, errors []CertificateRuntimeError) string {
	for index := len(errors) - 1; index >= 0; index-- {
		for _, subject := range subjects {
			domain := strings.ToLower(strings.TrimPrefix(subject, "*."))
			if domain != "" && strings.Contains(errors[index].Payload, domain) {
				return errors[index].Message
			}
		}
	}
	return ""
}

func isErrorLevel(value any) bool {
	level, _ := value.(string)
	return strings.EqualFold(level, "error")
}

func parseCaddyLogTime(value any) time.Time {
	switch timestamp := value.(type) {
	case float64:
		seconds := int64(timestamp)
		return time.Unix(seconds, int64((timestamp-float64(seconds))*float64(time.Second)))
	case string:
		parsed, _ := time.Parse(time.RFC3339, timestamp)
		return parsed
	default:
		return time.Time{}
	}
}
