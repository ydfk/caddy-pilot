package service

import (
	"testing"
)

func TestEvaluateCertificateIssuanceStates(t *testing.T) {
	subjects := []string{"*.example.com"}
	runtimeConfig := []byte(`{"subjects":["*.example.com"]}`)
	cases := []struct {
		name     string
		issued   []IssuedCertificate
		usage    int
		active   int
		config   []byte
		errors   []CertificateRuntimeError
		expected string
	}{
		{name: "未引用", expected: "unused"},
		{name: "等待启用", usage: 1, expected: "pending_publish"},
		{name: "等待发布", usage: 1, active: 1, expected: "pending_publish"},
		{name: "签发中", usage: 1, active: 1, config: runtimeConfig, expected: "issuing"},
		{name: "失败", usage: 1, active: 1, config: runtimeConfig, errors: []CertificateRuntimeError{{Payload: `{"domain":"example.com"}`, Message: "acme failed"}}, expected: "failed"},
		{name: "已签发", usage: 1, active: 1, config: runtimeConfig, issued: []IssuedCertificate{{Status: "valid"}}, expected: "issued"},
		{name: "即将到期", usage: 1, active: 1, config: runtimeConfig, issued: []IssuedCertificate{{Status: "expiring"}}, expected: "expiring"},
		{name: "已过期", usage: 1, active: 1, config: runtimeConfig, issued: []IssuedCertificate{{Status: "expired"}}, expected: "expired"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			status := EvaluateCertificateIssuance(subjects, testCase.issued, testCase.usage, testCase.active, testCase.config, testCase.errors)
			if status.State != testCase.expected {
				t.Fatalf("状态不正确: %+v", status)
			}
		})
	}
}
