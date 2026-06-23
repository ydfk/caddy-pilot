package certificate

import (
	"slices"
	"testing"
)

func TestCertificateProbeDomainsIncludesWildcardProbe(t *testing.T) {
	probes := certificateProbeDomains([]string{"*.example.com"}, []string{"app.example.com"})
	for _, expected := range []string{"app.example.com", "caddypilot-probe.example.com"} {
		if !slices.Contains(probes, expected) {
			t.Fatalf("证书探测域名缺少 %s: %+v", expected, probes)
		}
	}
}
