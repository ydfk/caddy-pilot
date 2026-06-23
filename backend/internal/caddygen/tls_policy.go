package caddygen

import (
	"encoding/json"
	"fmt"
	"strings"
)

type tlsPolicyAccumulator struct {
	policies        []map[string]any
	policyIndexes   map[string]int
	subjectPolicies map[string]string
}

func newTLSPolicyAccumulator() *tlsPolicyAccumulator {
	return &tlsPolicyAccumulator{
		policies: make([]map[string]any, 0), policyIndexes: make(map[string]int),
		subjectPolicies: make(map[string]string),
	}
}

func (accumulator *tlsPolicyAccumulator) Add(policy map[string]any) error {
	subjects, ok := policy["subjects"].([]string)
	if !ok || len(subjects) == 0 {
		return fmt.Errorf("TLS 自动化策略缺少域名")
	}
	configuration, err := json.Marshal(policy["issuers"])
	if err != nil {
		return fmt.Errorf("编码 TLS 签发策略失败: %w", err)
	}
	key := string(configuration)
	for _, subject := range subjects {
		normalized := strings.ToLower(strings.TrimSpace(subject))
		if existing, exists := accumulator.subjectPolicies[normalized]; exists && existing != key {
			return fmt.Errorf("域名 %s 配置了不同的 TLS 自动化策略", subject)
		}
	}
	index, exists := accumulator.policyIndexes[key]
	if !exists {
		index = len(accumulator.policies)
		accumulator.policyIndexes[key] = index
		accumulator.policies = append(accumulator.policies, map[string]any{
			"subjects": []string{}, "issuers": policy["issuers"],
		})
	}
	merged := accumulator.policies[index]["subjects"].([]string)
	for _, subject := range subjects {
		normalized := strings.ToLower(strings.TrimSpace(subject))
		if _, exists := accumulator.subjectPolicies[normalized]; exists {
			continue
		}
		accumulator.subjectPolicies[normalized] = key
		merged = append(merged, subject)
	}
	accumulator.policies[index]["subjects"] = merged
	return nil
}

func (accumulator *tlsPolicyAccumulator) Values() []map[string]any {
	return accumulator.policies
}
