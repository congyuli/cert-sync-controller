package controller

import (
	"strings"
	"testing"

	"github.com/Deraiven/cert-controller/internal/config"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenerateTagsFiltersFluxLabelsAndSanitizesKubernetesLabels(t *testing.T) {
	c := &CertificateController{
		config: &config.Config{
			Controller: config.ControllerConfig{
				ClusterName: "eks/test",
			},
		},
	}

	cert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"kustomize.toolkit.fluxcd.io/name":      "certificates",
				"kustomize.toolkit.fluxcd.io/namespace": "flux-system",
				"app.kubernetes.io/name":                "kong/cert-sync",
				"team":                                  "platform",
			},
		},
	}

	tags := c.generateTags("cert-manager", "example-cert", cert)

	for _, tag := range tags {
		if strings.Contains(tag, "toolkit.fluxcd.io") {
			t.Fatalf("expected Flux labels to be filtered, got tags: %#v", tags)
		}
		if strings.Contains(tag, "/") {
			t.Fatalf("expected Kong tags to be sanitized, got tag %q in %#v", tag, tags)
		}
	}

	assertContainsTag(t, tags, "cluster:eks-test")
	assertContainsTag(t, tags, "app.kubernetes.io-name:kong-cert-sync")
	assertContainsTag(t, tags, "team:platform")
}

func TestSanitizeKongTag(t *testing.T) {
	tests := map[string]string{
		"kustomize.toolkit.fluxcd.io/name:certificates": "kustomize.toolkit.fluxcd.io-name:certificates",
		"app.kubernetes.io/name:api/foo":                "app.kubernetes.io-name:api-foo",
		" namespace:cert-manager ":                      "namespace:cert-manager",
		"bad tag/value":                                 "bad-tag-value",
		strings.Repeat("a", maxKongTagLength+1):         strings.Repeat("a", maxKongTagLength),
	}

	for input, want := range tests {
		if got := sanitizeKongTag(input); got != want {
			t.Fatalf("sanitizeKongTag(%q) = %q, want %q", input, got, want)
		}
	}
}

func assertContainsTag(t *testing.T, tags []string, want string) {
	t.Helper()
	for _, tag := range tags {
		if tag == want {
			return
		}
	}
	t.Fatalf("expected tags to contain %q, got %#v", want, tags)
}
